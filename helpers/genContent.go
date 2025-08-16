package helpers

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/genai"
)

var modelName = "gemini-2.0-flash"

// GenerateContent queries Gemini with the given prompt
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

	resp, err := client.Models.GenerateContent(ctx, modelName, genai.Text(prompt), cfg)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("generate content returned nil")
	}
	return resp.Text(), nil
}

// FormatPrompt builds the actual instruction for the AI
func FormatPrompt(query, username string, history []string, wantLong bool) string {
	if wantLong {
		return fmt.Sprintf(
			"History: %v\n\nThe user \"%v\" asked: \"%v\".\n"+
				"Provide a detailed, well-structured response, up to 1800 characters (including spaces).",
			history, username, query,
		)
	}
	return fmt.Sprintf(
		"Conversation history: %v\n\n"+
			"User \"%v\" says: \"%v\".\n"+
			"Respond naturally as if you are chatting in Discord. "+
			"Keep it friendly, concise (max 2 sentences), and DO NOT say you lack information. "+
			"If the message isn’t a question, just reply casually like a person would.",
		history, username, query,
	)
}

// DecideLength checks if user explicitly wants a long answer
func DecideLength(query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(q, "detailed") ||
		strings.Contains(q, "long") ||
		strings.Contains(q, "explain fully")
}
