// Package bedrock は Amazon Bedrock の呼出 Adapter を提供する。
//
// 本ファイルはリトライ判定 (P-RETRY-01) を担う RetryClassifier interface と、
// その Bedrock 専用 implementation BedrockRetryClassifier を定義する。
// AWS SDK v2 のエラー型を errors.As / errors.Is で判定し、
// 一時的エラーはリトライ、永続エラーは即フォールバックを指示する。
package bedrock

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/aws/smithy-go"
)

// RetryClassifier はエラーがリトライすべきか判定する責務を持つ。
//
// Unit C / D で共有可能 (P-RETRY-01)。テスト時は fake Classifier を注入して
// PBT (P-PBT-01) でリトライ挙動を独立検証できる。
type RetryClassifier interface {
	ShouldRetry(err error) bool
}

// BedrockRetryClassifier は Amazon Bedrock 呼出時のエラー型を判定する。
//
// NFRC-C06 で確定したリトライ対象 / 即フォールバック対象を実装する:
//   - リトライ: ThrottlingException / ServiceUnavailableException /
//     context.DeadlineExceeded / Smithy GenericAPIError (FaultServer)
//   - 即フォールバック: ValidationException / AccessDeniedException /
//     ResourceNotFoundException 等の永続エラー
type BedrockRetryClassifier struct{}

// NewBedrockRetryClassifier は BedrockRetryClassifier を返す。
func NewBedrockRetryClassifier() *BedrockRetryClassifier {
	return &BedrockRetryClassifier{}
}

// ShouldRetry は err がリトライ対象なら true を返す。
//
// nil error は false を返す (リトライ不要)。
func (c *BedrockRetryClassifier) ShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// context タイムアウトはリトライ対象
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// context cancel はリトライしない (上位の意思決定が優先)
	if errors.Is(err, context.Canceled) {
		return false
	}

	// Bedrock SDK の型固有エラー
	var throttle *types.ThrottlingException
	if errors.As(err, &throttle) {
		return true
	}
	var unavailable *types.ServiceUnavailableException
	if errors.As(err, &unavailable) {
		return true
	}
	var internal *types.InternalServerException
	if errors.As(err, &internal) {
		return true
	}

	// 永続エラー (即フォールバック)
	var validation *types.ValidationException
	if errors.As(err, &validation) {
		return false
	}
	var accessDenied *types.AccessDeniedException
	if errors.As(err, &accessDenied) {
		return false
	}
	var notFound *types.ResourceNotFoundException
	if errors.As(err, &notFound) {
		return false
	}

	// Smithy 共通の API エラー: 5xx 系はリトライ
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		if apiErr.ErrorFault() == smithy.FaultServer {
			return true
		}
		return false
	}

	// 不明なエラー: ネットワーク系の可能性があるためリトライ
	return true
}
