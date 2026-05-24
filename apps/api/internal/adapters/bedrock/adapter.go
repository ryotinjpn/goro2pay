package bedrock

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// Plan は BedrockAdapter / FallbackSuggestProvider が返す注文計画。
//
// Source は計画の出処を識別する: "bedrock" | "fallback_history" | "fallback_default"
type Plan struct {
	StoreName        string
	MenuName         string
	Amount           int
	Category         string
	Source           string
	BedrockLatencyMs int64
	BedrockAttempt  int
	FallbackTriggered bool
}

// BedrockAdapter は Amazon Bedrock の Claude モデル呼出責務を担う interface。
//
// Unit C / D で共有可能 (NFRC-C20)。本番は ClaudeBedrockAdapter、テストは
// MockBedrockAdapter (P-MOCK-01) を注入する。
type BedrockAdapter interface {
	InferOrderPlan(ctx context.Context, history []HistoryItem, dayOfWeek string, category string) (*Plan, error)
}

// BedrockRuntimeAPI は ClaudeBedrockAdapter が依存する Bedrock SDK の最小 interface。
//
// テスト時は SDK Client を mock 化するためにこの interface に対する fake を
// NewClaudeBedrockAdapterWithClient で注入する。
type BedrockRuntimeAPI interface {
	Converse(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error)
}

// RetryReporter はリトライ発動時に呼ばれる callback (P-OBS-03 / NFRC-C13-2)。
//
// import cycle (bedrock → order) を避けるため、bedrock package 内で型を定義し、
// main.go の DI 配線で order.LogBedrockRetry を bridge して注入する。
// 未設定時のデフォルトは NoopRetryReporter。
type RetryReporter func(ctx context.Context, attempt int, errorClass string, elapsedMs int64)

// NoopRetryReporter は RetryReporter のデフォルト実装。何もしない。
func NoopRetryReporter(_ context.Context, _ int, _ string, _ int64) {}

// classifyError は AWS SDK / Smithy エラーを分類名 (string) で返す。
//
// CloudWatch Logs metric filter (NFRC-C13-2) で `errorClass` キーで集計するため、
// 安定した文字列ラベルを返すことが要件。
func classifyError(err error) string {
	if err == nil {
		return "Nil"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "DeadlineExceeded"
	}
	if errors.Is(err, context.Canceled) {
		return "Canceled"
	}
	var throttle *types.ThrottlingException
	if errors.As(err, &throttle) {
		return "ThrottlingException"
	}
	var unavailable *types.ServiceUnavailableException
	if errors.As(err, &unavailable) {
		return "ServiceUnavailableException"
	}
	var internal *types.InternalServerException
	if errors.As(err, &internal) {
		return "InternalServerException"
	}
	var validation *types.ValidationException
	if errors.As(err, &validation) {
		return "ValidationException"
	}
	var accessDenied *types.AccessDeniedException
	if errors.As(err, &accessDenied) {
		return "AccessDeniedException"
	}
	var notFound *types.ResourceNotFoundException
	if errors.As(err, &notFound) {
		return "ResourceNotFoundException"
	}
	return "Unknown"
}

// defaultClient は Lambda INIT フェーズで初期化される Bedrock SDK Client (P-INIT-01)。
//
// init() 失敗時は panic で fail-fast し、Lambda 起動失敗を CloudWatch で即検知する。
var defaultClient *bedrockruntime.Client

// init は Lambda INIT フェーズで AWS SDK を初期化する。
//
// arm64 + 256MB の INIT フェーズでは AWS が burst CPU を割り当てるため、
// SDK 初期化は INIT 内に押し込んで INVOKE フェーズの応答時間 (NFRC-C18 600ms)
// を確保する。
func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("bedrock: failed to load AWS config: %v", err))
	}
	defaultClient = bedrockruntime.NewFromConfig(cfg)
}

// ClaudeBedrockAdapter は Bedrock Converse API で Claude 3.5 Haiku を呼ぶ
// BedrockAdapter implementation。
//
// 単一呼出のタイムアウトは 1500ms (NFRC-C07)、リトライ判定は RetryClassifier
// に委譲する (NFRC-C06、P-RETRY-01)。retryReporter はリトライ発動時に呼ばれ、
// NFRC-C13-2 の WARN ログ出力 (P-OBS-03) に bridge される。
type ClaudeBedrockAdapter struct {
	client         BedrockRuntimeAPI
	classifier     RetryClassifier
	modelID        string
	retryReporter  RetryReporter
}

const (
	// bedrockTimeoutPerCall は Bedrock 1 呼出のタイムアウト (NFRC-C07)。
	bedrockTimeoutPerCall = 1500 * time.Millisecond

	// maxAttempts は Bedrock 呼出の最大試行回数 (NFRC-C06: 初回 + リトライ 1 回)。
	maxAttempts = 2

	// envInferenceProfileID は Bedrock 呼出に使う Inference Profile ID 環境変数。
	envInferenceProfileID = "BEDROCK_INFERENCE_PROFILE_ID"

	// defaultInferenceProfileID は env 未設定時のデフォルト Inference Profile ID。
	// ap-northeast-1 含む APAC 推奨 (NFRC-C20 / Q-I10=A)。
	defaultInferenceProfileID = "apac.anthropic.claude-3-5-haiku-20241022-v1:0"
)

// NewClaudeBedrockAdapter は production 用の ClaudeBedrockAdapter を返す。
//
// SDK Client は package-level の defaultClient (init() で初期化済み) を使用する。
func NewClaudeBedrockAdapter() *ClaudeBedrockAdapter {
	return NewClaudeBedrockAdapterWithClient(defaultClient)
}

// NewClaudeBedrockAdapterWithClient はテスト用に mock client を注入できる constructor。
//
// 統合テスト・PBT で SDK 呼出を完全制御するために使う (NFRC-C16 / P-MOCK-01)。
// retryReporter は NoopRetryReporter で初期化される。
func NewClaudeBedrockAdapterWithClient(client BedrockRuntimeAPI) *ClaudeBedrockAdapter {
	modelID := os.Getenv(envInferenceProfileID)
	if modelID == "" {
		modelID = defaultInferenceProfileID
	}
	return &ClaudeBedrockAdapter{
		client:        client,
		classifier:    NewBedrockRetryClassifier(),
		modelID:       modelID,
		retryReporter: NoopRetryReporter,
	}
}

// SetRetryReporter は retryReporter を差し替える (main.go の DI 配線で
// order.LogBedrockRetry を bridge する用途)。
func (a *ClaudeBedrockAdapter) SetRetryReporter(r RetryReporter) {
	if r == nil {
		a.retryReporter = NoopRetryReporter
		return
	}
	a.retryReporter = r
}

// InferOrderPlan は Bedrock Claude を呼んで注文計画を生成する。
//
// 実装ポイント:
//   - 各呼出に context.WithTimeout(parent, 1500ms) を適用 (NFRC-C07)
//   - エラーは RetryClassifier で分類、対象なら 1 回リトライ (NFRC-C06)
//   - 親 ctx の cancel / deadline は即時伝播する (NFRC-C10)
//   - 戻り Plan の BedrockLatencyMs / BedrockAttempt は計測値、Source="bedrock" を設定
func (a *ClaudeBedrockAdapter) InferOrderPlan(parentCtx context.Context, history []HistoryItem, dayOfWeek string, category string) (*Plan, error) {
	prompt, err := BuildPrompt(history, dayOfWeek, category)
	if err != nil {
		return nil, fmt.Errorf("build prompt: %w", err)
	}

	overall := time.Now()
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// 親 ctx の状態を都度確認、すでに cancel ならリトライしない (NFRC-C10)
		if err := parentCtx.Err(); err != nil {
			return nil, err
		}

		callCtx, cancel := context.WithTimeout(parentCtx, bedrockTimeoutPerCall)
		started := time.Now()
		out, err := a.client.Converse(callCtx, &bedrockruntime.ConverseInput{
			ModelId: aws.String(a.modelID),
			Messages: []types.Message{
				{
					Role: types.ConversationRoleUser,
					Content: []types.ContentBlock{
						&types.ContentBlockMemberText{Value: prompt},
					},
				},
			},
		})
		elapsed := time.Since(started)
		cancel()

		if err == nil {
			text, perr := extractText(out)
			if perr != nil {
				lastErr = perr
				slog.WarnContext(parentCtx, "bedrock_response_invalid", "attempt", attempt, "error", perr.Error())
				if !a.classifier.ShouldRetry(perr) || attempt >= maxAttempts {
					break
				}
				// レスポンス解釈失敗もリトライ発動 (NFRC-C13-2)
				a.retryReporter(parentCtx, attempt+1, "ResponseInvalid", elapsed.Milliseconds())
				continue
			}
			plan, perr := ParsePlanResponse(text)
			if perr != nil {
				lastErr = perr
				slog.WarnContext(parentCtx, "bedrock_response_invalid", "attempt", attempt, "error", perr.Error())
				if !a.classifier.ShouldRetry(perr) || attempt >= maxAttempts {
					break
				}
				a.retryReporter(parentCtx, attempt+1, "ResponseInvalid", elapsed.Milliseconds())
				continue
			}
			return &Plan{
				StoreName:        plan.StoreName,
				MenuName:         plan.MenuName,
				Amount:           plan.Amount,
				Category:         plan.Category,
				Source:           "bedrock",
				BedrockLatencyMs: time.Since(overall).Milliseconds(),
				BedrockAttempt:   attempt,
			}, nil
		}

		lastErr = err
		if !a.classifier.ShouldRetry(err) {
			return nil, err
		}
		// 次の attempt がある場合のみリトライ通知 (NFRC-C13-2 / P-OBS-03)
		if attempt < maxAttempts {
			a.retryReporter(parentCtx, attempt+1, classifyError(err), elapsed.Milliseconds())
		}
		// 待機なしで即リトライ (NFRC-C06、3 秒予算遵守を優先)
	}
	return nil, fmt.Errorf("bedrock: max attempts (%d) exhausted: %w", maxAttempts, lastErr)
}

// extractText は ConverseOutput から最初の text content を取り出す。
func extractText(out *bedrockruntime.ConverseOutput) (string, error) {
	if out == nil || out.Output == nil {
		return "", fmt.Errorf("bedrock: empty output")
	}
	msg, ok := out.Output.(*types.ConverseOutputMemberMessage)
	if !ok || msg == nil {
		return "", fmt.Errorf("bedrock: unexpected output type")
	}
	for _, c := range msg.Value.Content {
		if t, ok := c.(*types.ContentBlockMemberText); ok {
			return t.Value, nil
		}
	}
	return "", fmt.Errorf("bedrock: no text content in response")
}
