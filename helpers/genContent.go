package helpers

import (
	"context"
	"flag"
	"os"

	"google.golang.org/genai"
)

var model = flag.String("model", "gemini-2.0-flash", "Model name (e.g. gemini-2.0-flash)")

func GenerateContent(ctx context.Context, prompt string) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", err
	}

	cfg := &genai.GenerateContentConfig{
		Temperature: genai.Ptr[float32](0),
	}

	resp, err := client.Models.GenerateContent(ctx, *model, genai.Text(prompt), cfg)
	if err != nil {
		return "", err
	}
	return resp.Text(), nil
}
