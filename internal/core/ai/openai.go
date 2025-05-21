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
	2. Do NOT change the text or the order of the text content (e.g. ingredients, steps, etc.).
	3. If you cannot find the information, leave the JSON field empty. I.e., if an ingredient is missing quantity or unit, set those to default values (0 or "").
	4. Do NOT make up any information.`
)

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
	prompt := fmt.Sprintf("Analyze the attached image of a recipe and extract the data. You must follow the rules below.\n\nOutput rules:\n%s",
		_PromptRules)

	// Create the request body
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
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
		Temperature:         openai.Float(0),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        "recipe",
					Strict:      openai.Opt(true),
					Description: openai.Opt("A JSON object representing a recipe"),
					Schema:      result.JSONSchema(),
				},
			},
		},
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

	prompt := fmt.Sprintf("Analyze the URL including a recipe and extract the data. You must follow the rules below.\n\nURL: %s\n\nOutput rules:\n%s",
		url,
		_PromptRules)

	// Create the request body
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model:               c.model,
		MaxCompletionTokens: openai.Int(3000),
		Temperature:         openai.Float(0),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        "recipe",
					Strict:      openai.Opt(true),
					Description: openai.Opt("A JSON object representing a recipe"),
					Schema:      result.JSONSchema(),
				},
			},
		},
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
