package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type ImageContentType string

const (
	ImageContentTypeJPEG ImageContentType = "image/jpeg"
	ImageContentTypePNG  ImageContentType = "image/png"
)

type OpenAI struct {
	client openai.Client
	model  string
}

func NewOpenAI(apiKey string, model string) AI {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &OpenAI{
		client: client,
		model:  model,
	}
}

func (c *OpenAI) AnalyzeImage(ctx context.Context, image bytes.Buffer, imageContentType ImageContentType, prompt string) (string, error) {
	// Encode the image as base64
	base64Image := base64.StdEncoding.EncodeToString(image.Bytes())
	dataURI := fmt.Sprintf("data:%s;base64,%s", imageContentType, base64Image)

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
						},
					},
				},
			},
		},
		Model:               c.model,
		MaxCompletionTokens: openai.Int(2000),
	}

	// Call the API
	chatCompletion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

func (c *OpenAI) AnalyzeURL(ctx context.Context, url string, prompt string) (string, error) {
	// Create the request body
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf("URL: %s\n\n%s", url, prompt)),
		},
		Model:               c.model,
		MaxCompletionTokens: openai.Int(2000),
	}

	// Call the API
	chatCompletion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}

	return chatCompletion.Choices[0].Message.Content, nil
}
