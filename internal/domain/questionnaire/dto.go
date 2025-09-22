package questionnaire

import (
	"time"

	"github.com/google/uuid"
)

type GenerateQuestionnaireRequest struct {
	Name               string   `json:"name" validate:"required,min=2,max=100"`
	QuestionCount      int      `json:"question_count" validate:"required,min=3,max=20"`
	TargetRoles        []string `json:"target_roles" validate:"required,min=1"`
	DifficultyLevel    string   `json:"difficulty_level" validate:"required,oneof=basic intermediate advanced"`
	FocusAreas         []string `json:"focus_areas" validate:"required,min=1"`
	CustomInstructions *string  `json:"custom_instructions,omitempty"`
}

type GenerateQuestionnaireResponse struct {
	QuestionnaireID uuid.UUID `json:"questionnaire_id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	Message         string    `json:"message"`
	EstimatedTime   string    `json:"estimated_time"`
}

type GenerationStatusResponse struct {
	QuestionnaireID uuid.UUID  `json:"questionnaire_id"`
	Status          string     `json:"status"`
	Progress        int        `json:"progress"`
	Message         string     `json:"message"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type ActiveQuestionnaireResponse struct {
	ID        uuid.UUID                  `json:"id"`
	Name      string                     `json:"name"`
	Version   int                        `json:"version"`
	Questions []QuestionnaireQuestionDTO `json:"questions"`
}

type QuestionnaireQuestionDTO struct {
	ID           uuid.UUID    `json:"id"`
	QuestionText string       `json:"question_text"`
	QuestionType QuestionType `json:"question_type"`
	Options      []string     `json:"options,omitempty"`
	MaxScore     int          `json:"max_score"`
	Category     string       `json:"category"`
	Order        int          `json:"order"`
}

type SubmitQuestionnaireRequest struct {
	QuestionnaireID uuid.UUID    `json:"questionnaire_id" validate:"required"`
	Answers         []AnswerItem `json:"answers" validate:"required,dive"`
}

type AnswerItem struct {
	QuestionID     uuid.UUID `json:"question_id" validate:"required"`
	SelectedOption *string   `json:"selected_option,omitempty"`
	Score          *int      `json:"score,omitempty"`
	TextAnswer     *string   `json:"text_answer,omitempty"`
}

type SubmitQuestionnaireResponse struct {
	ResponseID      uuid.UUID `json:"response_id"`
	QuestionnaireID uuid.UUID `json:"questionnaire_id"`
	Message         string    `json:"message"`
	ProcessingTime  string    `json:"processing_time"`
}

type PersonalityAnalysis struct {
	PersonalityTraits []string `json:"personality_traits"`
	Interests         []string `json:"interests"`
	Strengths         []string `json:"strengths"`
	WorkStyle         string   `json:"work_style"`
}

type AIRecommendation struct {
	RoleID        string  `json:"role_id"`
	RoleName      string  `json:"role_name"`
	Score         float64 `json:"score"`
	Justification string  `json:"justification"`
	Category      string  `json:"category"`
}

type RoleRecommendationResponse struct {
	ID          uuid.UUID `json:"id"`
	RoleName    string    `json:"role_name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type QuestionnaireListResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Version     int       `json:"version"`
	Active      bool      `json:"active"`
	GeneratedBy string    `json:"generated_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type QuestionnaireResponseSummary struct {
	ID              uuid.UUID  `json:"id"`
	StudentName     string     `json:"student_name"`
	StudentEmail    string     `json:"student_email"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	TotalScore      *int       `json:"total_score,omitempty"`
	RecommendedRole *string    `json:"recommended_role,omitempty"`
	Status          string     `json:"status"`
}

type QuestionOption struct {
	Text  string `json:"text"`
	Label string `json:"label"`
	Value string `json:"value"`
	Score int    `json:"score,omitempty"`
}

type QuestionnaireQuestionResponse struct {
	ID             uuid.UUID        `json:"id"`
	ResponseID     uuid.UUID        `json:"response_id,omitempty"`
	QuestionID     uuid.UUID        `json:"question_id,omitempty"`
	QuestionText   string           `json:"question_text"`
	QuestionType   string           `json:"question_type"`
	MaxScore       int              `json:"max_score"`
	QuestionOrder  int              `json:"question_order"`
	Category       string           `json:"category"`
	Options        []QuestionOption `json:"options,omitempty"`
	SelectedOption *string          `json:"selected_option,omitempty"`
	Score          *int             `json:"score,omitempty"`
	TextAnswer     *string          `json:"text_answer,omitempty"`
}

type QuestionnaireResultResponse struct {
	ID                  uuid.UUID            `json:"id"`
	ResponseID          uuid.UUID            `json:"response_id,omitempty"`
	QuestionnaireID     uuid.UUID            `json:"questionnaire_id"`
	QuestionnaireName   string               `json:"questionnaire_name,omitempty"`
	SubmittedAt         time.Time            `json:"submitted_at"`
	ProcessedAt         *time.Time           `json:"processed_at,omitempty"`
	TotalScore          *int                 `json:"total_score,omitempty"`
	MaxPossibleScore    int                  `json:"max_possible_score,omitempty"`
	RecommendedRole     *string              `json:"recommended_role,omitempty"`
	PersonalityAnalysis *PersonalityAnalysis `json:"personality_analysis,omitempty"`
	AIRecommendations   []AIRecommendation   `json:"recommendations,omitempty"`
	Status              string               `json:"status,omitempty"`
}

type RoleRecommendationDTO struct {
	ID          uuid.UUID `json:"id"`
	RoleName    string    `json:"role_name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Active      bool      `json:"active"`
}

type QuestionnaireDetailResponse struct {
	ID          uuid.UUID                       `json:"id"`
	Name        string                          `json:"name"`
	Version     int                             `json:"version"`
	GeneratedBy string                          `json:"generated_by"`
	Active      bool                            `json:"active"`
	Questions   []QuestionnaireQuestionResponse `json:"questions"`
	CreatedAt   time.Time                       `json:"created_at"`
	UpdatedAt   time.Time                       `json:"updated_at"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
