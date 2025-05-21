package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type ImageContentType string

const (
	ImageContentTypeJPEG ImageContentType = "image/jpeg"
	ImageContentTypePNG  ImageContentType = "image/png"
)

const (
	_PromptRules = `
	1. Do NOT translate any content.
	2. Do NOT change the text or the order of the text.
	3. If you cannot find the information, leave the JSON field empty.
	4. Do NOT make up any information.
	5. If the recipe is not found, return an empty JSON object.`
)

// SchemaProvider defines an interface for types that can provide their own JSON schema
type SchemaProvider interface {
	JSONSchema() string
}

type OpenAI struct {
	client openai.Client
	model  string
}

func NewOpenAI(apiKey string, model string) RecipeAI {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &OpenAI{
		client: client,
		model:  model,
	}
}

func (c *OpenAI) AnalyzeRecipeImage(ctx context.Context, base64Image string, imageContentType ImageContentType) (*RecipeAnalysisResult, error) {
	result := &RecipeAnalysisResult{}

	// Create the data URI for the image
	dataURI := fmt.Sprintf("data:%s;base64,%s", imageContentType, base64Image)

	// Create the prompt with JSON format instruction
	prompt := fmt.Sprintf("Analyze the attached image of a recipe and extract the data. Important to follow the rules below.\n\nRules:\n%s\n\nOutput:\nProvide your response in JSON format following this structure:\n%s",
		_PromptRules,
		result.JSONSchema())

	// Create the request body
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Role: "user",
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
							{
								OfImageURL: &openai.ChatCompletionContentPartImageParam{
									ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
										URL: dataURI,
									},
								},
							},
							{
								OfText: &openai.ChatCompletionContentPartTextParam{
									Text: prompt,
								},
							},
						},
					},
				},
			},
		},
		Model:               c.model,
		MaxCompletionTokens: openai.Int(3000),
	}

	// Call the API
	chatCompletion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON response into the provided struct
	content := chatCompletion.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

func (c *OpenAI) AnalyzeRecipeURL(ctx context.Context, url string) (*RecipeAnalysisResult, error) {
	result := &RecipeAnalysisResult{}

	// Create the prompt with JSON format instruction
	prompt := fmt.Sprintf("Analyze the URL below that is a recipe and extract the data. Important to follow the rules below.\n\nURL: %s\n\nRules:\n%s\n\nOutput:\nProvide your response in JSON format following this structure:\n%s",
		url,
		_PromptRules,
		result.JSONSchema())

	// Create the request body
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model:               c.model,
		MaxCompletionTokens: openai.Int(3000),
	}

	// Call the API
	chatCompletion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON response into the provided struct
	content := chatCompletion.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}
