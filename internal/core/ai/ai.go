package ai

import (
	"bytes"
	"context"
)

type AI interface {
	AnalyzeImage(ctx context.Context, image bytes.Buffer, imageContentType ImageContentType, prompt string) (string, error)
	AnalyzeURL(ctx context.Context, url string, prompt string) (string, error)
}
