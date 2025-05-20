package ai

import (
	"context"
)

type AI interface {
	AnalyzeImage(ctx context.Context, base64Image string, imageContentType ImageContentType, prompt string) (string, error)
	AnalyzeURL(ctx context.Context, url string, prompt string) (string, error)
}
