package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AIService interface {
	GenerateCareerRecommendations(ctx context.Context, prompt string) (*CareerAnalysisResponse, error)
	GenerateQuestions(ctx context.Context, prompt string) (*QuestionGenerationResponse, error)
}

type GeminiService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewGeminiService(apiKey string) (*GeminiService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel("gemini-pro")
	model.SetTemperature(0.7)
	model.SetTopK(40)
	model.SetTopP(0.95)

	return &GeminiService{
		client: client,
		model:  model,
	}, nil
}

func (g *GeminiService) GenerateCareerRecommendations(ctx context.Context, prompt string) (*CareerAnalysisResponse, error) {
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gagal generate konten: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("tidak ada respons yang dihasilkan")
	}

	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText += string(txt)
		}
	}

	var result CareerAnalysisResponse
	err = json.Unmarshal([]byte(responseText), &result)
	if err != nil {
		return nil, fmt.Errorf("gagal parse respons AI: %w", err)
	}

	return &result, nil
}

func (g *GeminiService) GenerateQuestions(ctx context.Context, prompt string) (*QuestionGenerationResponse, error) {
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gagal generate konten: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("tidak ada respons yang dihasilkan")
	}

	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText += string(txt)
		}
	}

	var result QuestionGenerationResponse
	err = json.Unmarshal([]byte(responseText), &result)
	if err != nil {
		return nil, fmt.Errorf("gagal parse respons AI: %w", err)
	}

	return &result, nil
}

func (g *GeminiService) Close() error {
	return g.client.Close()
}
