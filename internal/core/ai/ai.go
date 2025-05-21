package ai

import (
	"context"
)

type RecipeAI interface {
	AnalyzeRecipeImage(ctx context.Context, base64Image string, imageContentType ImageContentType) (*RecipeAnalysisResult, error)
	AnalyzeRecipeWebpage(ctx context.Context, url string) (*RecipeAnalysisResult, error)
}
