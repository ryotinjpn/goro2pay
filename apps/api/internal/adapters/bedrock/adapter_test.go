package bedrock

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBedrockClient は BedrockRuntimeAPI を satisfy する手書き mock。
type fakeBedrockClient struct {
	converseFn func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error)
	calls      int
}

func (f *fakeBedrockClient) Converse(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
	f.calls++
	return f.converseFn(ctx, params, optFns...)
}

func newSuccessOutput(text string) *bedrockruntime.ConverseOutput {
	return &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: text},
				},
			},
		},
	}
}

func TestClaudeBedrockAdapter_InferOrderPlan_Success(t *testing.T) {
	fake := &fakeBedrockClient{
		converseFn: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
			return newSuccessOutput(`{"storeName":"ゴロゴロ食堂","menuName":"おまかせ定食","amount":1000,"category":"food"}`), nil
		},
	}
	adapter := NewClaudeBedrockAdapterWithClient(fake)

	plan, err := adapter.InferOrderPlan(context.Background(), nil, "Mon", "food")
	require.NoError(t, err)
	assert.Equal(t, "ゴロゴロ食堂", plan.StoreName)
	assert.Equal(t, 1, plan.BedrockAttempt)
	assert.Equal(t, "bedrock", plan.Source)
	assert.Equal(t, 1, fake.calls)
}

func TestClaudeBedrockAdapter_InferOrderPlan_RetryThenSuccess(t *testing.T) {
	attempts := 0
	fake := &fakeBedrockClient{
		converseFn: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
			attempts++
			if attempts == 1 {
				return nil, &types.ThrottlingException{Message: aws.String("Rate exceeded")}
			}
			return newSuccessOutput(`{"storeName":"ぐうたら亭","menuName":"手抜き丼","amount":800,"category":"food"}`), nil
		},
	}
	adapter := NewClaudeBedrockAdapterWithClient(fake)

	plan, err := adapter.InferOrderPlan(context.Background(), nil, "Tue", "food")
	require.NoError(t, err)
	assert.Equal(t, "ぐうたら亭", plan.StoreName)
	assert.Equal(t, 2, plan.BedrockAttempt)
	assert.Equal(t, 2, fake.calls)
}

func TestClaudeBedrockAdapter_InferOrderPlan_TwoFailuresExhaustsRetries(t *testing.T) {
	fake := &fakeBedrockClient{
		converseFn: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
			return nil, &types.ThrottlingException{Message: aws.String("Rate exceeded")}
		},
	}
	adapter := NewClaudeBedrockAdapterWithClient(fake)

	_, err := adapter.InferOrderPlan(context.Background(), nil, "Wed", "food")
	require.Error(t, err)
	assert.Equal(t, 2, fake.calls, "max retries should be 2")
}

func TestClaudeBedrockAdapter_InferOrderPlan_PermanentErrorNoRetry(t *testing.T) {
	fake := &fakeBedrockClient{
		converseFn: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
			return nil, &types.ValidationException{Message: aws.String("invalid params")}
		},
	}
	adapter := NewClaudeBedrockAdapterWithClient(fake)

	_, err := adapter.InferOrderPlan(context.Background(), nil, "Thu", "food")
	require.Error(t, err)
	assert.Equal(t, 1, fake.calls, "permanent error should not retry")
}

func TestClaudeBedrockAdapter_InferOrderPlan_ParentContextCanceled(t *testing.T) {
	fake := &fakeBedrockClient{
		converseFn: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
			return newSuccessOutput(`{"storeName":"x","menuName":"y","amount":100,"category":"food"}`), nil
		},
	}
	adapter := NewClaudeBedrockAdapterWithClient(fake)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := adapter.InferOrderPlan(ctx, nil, "Fri", "food")
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.Equal(t, 0, fake.calls)
}

// B-C1 修正検証: リトライ判定が true のとき RetryReporter が呼ばれること。
func TestClaudeBedrockAdapter_RetryReporterCalledOnRetry(t *testing.T) {
	attempts := 0
	fake := &fakeBedrockClient{
		converseFn: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
			attempts++
			if attempts == 1 {
				return nil, &types.ThrottlingException{Message: aws.String("Rate exceeded")}
			}
			return newSuccessOutput(`{"storeName":"X","menuName":"Y","amount":100,"category":"food"}`), nil
		},
	}
	adapter := NewClaudeBedrockAdapterWithClient(fake)

	type reportCall struct {
		attempt    int
		errorClass string
	}
	var calls []reportCall
	adapter.SetRetryReporter(func(ctx context.Context, attempt int, errorClass string, elapsedMs int64) {
		calls = append(calls, reportCall{attempt, errorClass})
	})

	_, err := adapter.InferOrderPlan(context.Background(), nil, "Mon", "food")
	require.NoError(t, err)
	require.Len(t, calls, 1, "RetryReporter must be called exactly once on retry")
	assert.Equal(t, 2, calls[0].attempt, "reported attempt should be 2 (next retry)")
	assert.Equal(t, "ThrottlingException", calls[0].errorClass)
}

// B-C1 修正検証: 永続エラー (リトライ非対象) では RetryReporter が呼ばれない。
func TestClaudeBedrockAdapter_RetryReporterNotCalledOnPermanentError(t *testing.T) {
	fake := &fakeBedrockClient{
		converseFn: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
			return nil, &types.ValidationException{Message: aws.String("invalid")}
		},
	}
	adapter := NewClaudeBedrockAdapterWithClient(fake)

	called := false
	adapter.SetRetryReporter(func(ctx context.Context, attempt int, errorClass string, elapsedMs int64) {
		called = true
	})

	_, err := adapter.InferOrderPlan(context.Background(), nil, "Mon", "food")
	require.Error(t, err)
	assert.False(t, called, "RetryReporter must NOT be called on permanent error")
}

// B-C1 修正検証: 2 回連続失敗時は最終 attempt の RetryReporter は呼ばない
// (= maxAttempts 到達後の retry は存在しないので報告不要)。1 回目失敗→2 回目挑戦の
// 通知 1 回のみ。
func TestClaudeBedrockAdapter_RetryReporterNotCalledAfterFinalAttempt(t *testing.T) {
	fake := &fakeBedrockClient{
		converseFn: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
			return nil, &types.ThrottlingException{Message: aws.String("Rate")}
		},
	}
	adapter := NewClaudeBedrockAdapterWithClient(fake)

	calls := 0
	adapter.SetRetryReporter(func(ctx context.Context, attempt int, errorClass string, elapsedMs int64) {
		calls++
	})

	_, err := adapter.InferOrderPlan(context.Background(), nil, "Mon", "food")
	require.Error(t, err)
	assert.Equal(t, 1, calls, "Reporter must be called once (attempt=2 retry kickoff), not after final exhaustion")
}

// B-C1 修正検証: SetRetryReporter(nil) 後も Noop で動作すること (panic しない)。
func TestClaudeBedrockAdapter_NilReporterDefaultsToNoop(t *testing.T) {
	fake := &fakeBedrockClient{
		converseFn: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
			return nil, &types.ThrottlingException{Message: aws.String("Rate")}
		},
	}
	adapter := NewClaudeBedrockAdapterWithClient(fake)
	adapter.SetRetryReporter(nil) // panic しないこと
	_, err := adapter.InferOrderPlan(context.Background(), nil, "Mon", "food")
	require.Error(t, err)
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"nil", nil, "Nil"},
		{"deadline", context.DeadlineExceeded, "DeadlineExceeded"},
		{"canceled", context.Canceled, "Canceled"},
		{"throttle", &types.ThrottlingException{}, "ThrottlingException"},
		{"validation", &types.ValidationException{}, "ValidationException"},
		{"access denied", &types.AccessDeniedException{}, "AccessDeniedException"},
		{"not found", &types.ResourceNotFoundException{}, "ResourceNotFoundException"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, classifyError(tt.err))
		})
	}
}

func TestMockBedrockAdapter_DefaultBehavior(t *testing.T) {
	mock := &MockBedrockAdapter{}
	plan, err := mock.InferOrderPlan(context.Background(), nil, "Sat", "food")
	require.NoError(t, err)
	assert.Equal(t, "ゴロゴロ食堂", plan.StoreName)
	assert.Equal(t, 1, mock.Calls)
}

func TestMockBedrockAdapter_CustomFunc(t *testing.T) {
	mock := &MockBedrockAdapter{
		InferOrderPlanFunc: func(ctx context.Context, history []HistoryItem, dayOfWeek string, category string) (*Plan, error) {
			return nil, errors.New("test failure")
		},
	}
	_, err := mock.InferOrderPlan(context.Background(), nil, "Sun", "food")
	require.Error(t, err)
}
