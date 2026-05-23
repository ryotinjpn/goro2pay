package bedrock

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
)

func TestBedrockRetryClassifier_ShouldRetry(t *testing.T) {
	c := NewBedrockRetryClassifier()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"context deadline exceeded", context.DeadlineExceeded, true},
		{"context canceled (上位の意思決定優先)", context.Canceled, false},
		{
			"ThrottlingException",
			&types.ThrottlingException{Message: aws.String("Rate exceeded")},
			true,
		},
		{
			"ServiceUnavailableException",
			&types.ServiceUnavailableException{Message: aws.String("temporarily unavailable")},
			true,
		},
		{
			"InternalServerException",
			&types.InternalServerException{Message: aws.String("server error")},
			true,
		},
		{
			"ValidationException (永続)",
			&types.ValidationException{Message: aws.String("invalid input")},
			false,
		},
		{
			"AccessDeniedException (永続)",
			&types.AccessDeniedException{Message: aws.String("forbidden")},
			false,
		},
		{
			"ResourceNotFoundException (永続)",
			&types.ResourceNotFoundException{Message: aws.String("model not found")},
			false,
		},
		{"unknown error (ネットワーク想定でリトライ)", errors.New("dial tcp: i/o timeout"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, c.ShouldRetry(tt.err))
		})
	}
}
