package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

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
	if apiKey == "" {
		return nil, fmt.Errorf("API key tidak boleh kosong")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat client Gemini: %w", err)
	}

	model := client.GenerativeModel("gemini-2.0-flash")

	model.SetTemperature(0.7)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(8192)

	log.Println("Testing Gemini connection...")
	testResp, err := model.GenerateContent(ctx, genai.Text("Hello"))
	if err != nil {
		log.Printf("Connection test failed: %v", err)
		return nil, fmt.Errorf("gagal test koneksi Gemini: %w", err)
	}

	if len(testResp.Candidates) > 0 && len(testResp.Candidates[0].Content.Parts) > 0 {
		log.Println("Gemini connection test successful")
	}

	log.Println("Gemini AI service berhasil diinisialisasi")
	return &GeminiService{
		client: client,
		model:  model,
	}, nil
}

func (g *GeminiService) GenerateCareerRecommendations(ctx context.Context, prompt string) (*CareerAnalysisResponse, error) {
	log.Printf("Memulai generasi rekomendasi karir dengan Gemini AI...")
	log.Printf("Prompt length: %d characters", len(prompt))

	if len(prompt) == 0 {
		return nil, fmt.Errorf("prompt tidak boleh kosong")
	}

	formattedPrompt := fmt.Sprintf(`
Analisis jawaban kuesioner berikut dan berikan rekomendasi karir dalam format JSON yang tepat:

%s

Berikan respons dalam format JSON berikut (tanpa markdown atau text tambahan):
{
  "analysis": {
    "personality_traits": ["trait1", "trait2", "trait3"],
    "interests": ["interest1", "interest2", "interest3"],
    "strengths": ["strength1", "strength2", "strength3"],
    "work_style": "string_description"
  },
  "recommendations": [
    {
      "role_id": "unique_role_id",
      "role_name": "Nama Role",
      "score": 85.5,
      "justification": "Penjelasan mengapa role ini cocok"
    }
  ]
}

IMPORTANT: Berikan HANYA JSON yang valid, tanpa text lain.
`, prompt)

	resp, err := g.model.GenerateContent(ctx, genai.Text(formattedPrompt))
	if err != nil {
		log.Printf("Error calling Gemini API: %v", err)
		return nil, fmt.Errorf("gagal generate konten dari Gemini: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("tidak ada candidates dalam respons Gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("tidak ada content parts dalam respons Gemini")
	}

	var responseText string
	for _, part := range candidate.Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText += string(txt)
		}
	}

	log.Printf("Raw response length: %d", len(responseText))
	log.Printf("Finish reason: %v", candidate.FinishReason)

	responseText = cleanJSONResponse(responseText)

	if len(responseText) == 0 {
		return nil, fmt.Errorf("respons kosong setelah cleanup")
	}

	var result CareerAnalysisResponse
	err = json.Unmarshal([]byte(responseText), &result)
	if err != nil {
		log.Printf("JSON Parse Error: %v", err)
		log.Printf("Raw response (first 500 chars): %s", truncateString(responseText, 500))
		return nil, fmt.Errorf("gagal parse respons AI sebagai JSON: %w", err)
	}

	if len(result.Recommendations) == 0 {
		return nil, fmt.Errorf("tidak ada rekomendasi dalam respons AI")
	}

	log.Printf("Berhasil generate %d rekomendasi karir", len(result.Recommendations))
	return &result, nil
}

func (g *GeminiService) GenerateQuestions(ctx context.Context, prompt string) (*QuestionGenerationResponse, error) {
	log.Printf("Memulai generasi pertanyaan dengan Gemini AI...")
	log.Printf("Prompt length: %d characters", len(prompt))

	if len(prompt) == 0 {
		return nil, fmt.Errorf("prompt tidak boleh kosong")
	}

	formattedPrompt := fmt.Sprintf(`
Generate questionnaire based on the following requirements:

%s

Berikan respons dalam format JSON berikut (tanpa markdown atau text tambahan):
{
  "questions": [
    {
      "question_text": "Text pertanyaan",
      "question_type": "mcq|likert|text|case",
      "options": [
        {"label": "Option A", "value": "a"},
        {"label": "Option B", "value": "b"}
      ],
      "category": "category_name",
      "reasoning": "Alasan pertanyaan ini penting"
    }
  ],
  "metadata": {
    "total_questions": 5,
    "distribution": {"mcq": 2, "likert": 2, "text": 1},
    "target_roles_coverage": ["Backend Developer", "Frontend Developer"]
  }
}

Guidelines:
- question_type: "mcq" (multiple choice), "likert" (1-5 scale), "text" (open text), "case" (scenario)
- category: "interests", "personality", "skills", "experience", "preferences"
- options: hanya untuk "mcq" type
- Buat pertanyaan yang relevan untuk profiling karir di bidang teknologi

IMPORTANT: Berikan HANYA JSON yang valid, tanpa text lain.
`, prompt)

	resp, err := g.model.GenerateContent(ctx, genai.Text(formattedPrompt))
	if err != nil {
		log.Printf("Error calling Gemini API: %v", err)
		return nil, fmt.Errorf("gagal generate konten dari Gemini: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("tidak ada candidates dalam respons Gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("tidak ada content parts dalam respons Gemini")
	}

	var responseText string
	for _, part := range candidate.Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText += string(txt)
		}
	}

	log.Printf("Raw response length: %d", len(responseText))
	log.Printf("Finish reason: %v", candidate.FinishReason)

	responseText = cleanJSONResponse(responseText)

	if len(responseText) == 0 {
		return nil, fmt.Errorf("respons kosong setelah cleanup")
	}

	var result QuestionGenerationResponse
	err = json.Unmarshal([]byte(responseText), &result)
	if err != nil {
		log.Printf("JSON Parse Error: %v", err)
		log.Printf("Raw response (first 500 chars): %s", truncateString(responseText, 500))
		return nil, fmt.Errorf("gagal parse respons AI sebagai JSON: %w", err)
	}

	if len(result.Questions) == 0 {
		return nil, fmt.Errorf("tidak ada pertanyaan dalam respons AI")
	}

	log.Printf("Berhasil generate %d pertanyaan", len(result.Questions))
	return &result, nil
}

func (g *GeminiService) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

func cleanJSONResponse(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSpace(text)

	startIndex := strings.Index(text, "{")
	if startIndex == -1 {
		return ""
	}

	endIndex := strings.LastIndex(text, "}")
	if endIndex == -1 || endIndex <= startIndex {
		return ""
	}

	return text[startIndex : endIndex+1]
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type NoAIService struct{}

func NewNoAIService() *NoAIService {
	log.Println("Menggunakan NoAI service - fitur AI tidak tersedia")
	return &NoAIService{}
}

func (n *NoAIService) GenerateCareerRecommendations(ctx context.Context, prompt string) (*CareerAnalysisResponse, error) {
	return nil, fmt.Errorf("layanan AI tidak tersedia - GEMINI_API_KEY tidak dikonfigurasi")
}

func (n *NoAIService) GenerateQuestions(ctx context.Context, prompt string) (*QuestionGenerationResponse, error) {
	return nil, fmt.Errorf("layanan AI tidak tersedia - GEMINI_API_KEY tidak dikonfigurasi")
}
