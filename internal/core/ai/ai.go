package ai

import (
	"context"
)

type RecipeAI interface {
	AnalyzeRecipeImage(ctx context.Context, base64Image string, imageContentType ImageContentType) (*RecipeAnalysisResult, error)
	AnalyzeRecipeURL(ctx context.Context, url string) (*RecipeAnalysisResult, error)
}
