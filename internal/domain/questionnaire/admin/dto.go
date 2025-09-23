package admin

import (
	"time"

	"github.com/google/uuid"
)

// Target Role Management DTOs
type CreateTargetRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"required,max=500"`
	Category    string `json:"category" validate:"required,max=50"`
}

type UpdateTargetRoleRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Category    *string `json:"category,omitempty" validate:"omitempty,max=50"`
	Active      *bool   `json:"active,omitempty"`
}

type TargetRoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Questionnaire Generation DTOs
type GenerateQuestionnaireRequest struct {
	Name               string   `json:"name" validate:"required,min=2,max=100"`
	QuestionCount      int      `json:"question_count" validate:"required,min=3,max=20"`
	TargetRoles        []string `json:"target_roles" validate:"required,min=1"`
	DifficultyLevel    string   `json:"difficulty_level" validate:"required,oneof=basic intermediate advanced"`
	FocusAreas         []string `json:"focus_areas" validate:"required,min=1"`
	CustomInstructions *string  `json:"custom_instructions,omitempty"`
}

type QuestionnaireGenerationResponse struct {
	QuestionnaireID uuid.UUID `json:"questionnaire_id"`
	Status          string    `json:"status"`
	Progress        int       `json:"progress"`
	Message         string    `json:"message"`
}

type GeneratedQuestionResponse struct {
	ID           uuid.UUID   `json:"id"`
	QuestionText string      `json:"question_text"`
	QuestionType string      `json:"question_type"`
	Category     string      `json:"category"`
	Options      []OptionDTO `json:"options,omitempty"`
	MaxScore     int         `json:"max_score"`
	Order        int         `json:"order"`
}

type OptionDTO struct {
	Text  string `json:"text"`
	Value string `json:"value"`
	Score int    `json:"score"`
}

// Questionnaire Management DTOs
type QuestionnaireListResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Version     string               `json:"version"`
	TargetRoles []TargetRoleResponse `json:"target_roles"`
	Active      bool                 `json:"active"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type QuestionnaireDetailResponse struct {
	ID               uuid.UUID                   `json:"id"`
	Name             string                      `json:"name"`
	Description      string                      `json:"description"`
	Version          string                      `json:"version"`
	TargetRoles      []TargetRoleResponse        `json:"target_roles"`
	Questions        []GeneratedQuestionResponse `json:"questions"`
	Active           bool                        `json:"active"`
	TotalSubmissions int                         `json:"total_submissions"`
	CreatedAt        time.Time                   `json:"created_at"`
	UpdatedAt        time.Time                   `json:"updated_at"`
}

type ActivateQuestionnaireRequest struct {
	Active bool `json:"active"`
}

// Response Overview DTOs
type QuestionnaireResponseOverview struct {
	ID                 uuid.UUID              `json:"id"`
	QuestionnaireID    uuid.UUID              `json:"questionnaire_id"`
	QuestionnaireName  string                 `json:"questionnaire_name"`
	StudentName        string                 `json:"student_name"`
	StudentEmail       string                 `json:"student_email"`
	TotalScore         int                    `json:"total_score"`
	MaxScore           int                    `json:"max_score"`
	ScorePercentage    float64                `json:"score_percentage"`
	TopRecommendations []TopRecommendationDTO `json:"top_recommendations"`
	ProcessingStatus   string                 `json:"processing_status"`
	SubmittedAt        time.Time              `json:"submitted_at"`
}

type TopRecommendationDTO struct {
	RoleName string  `json:"role_name"`
	Score    float64 `json:"score"`
	Category string  `json:"category"`
}

type ResponseDetailResponse struct {
	ID                uuid.UUID                   `json:"id"`
	QuestionnaireID   uuid.UUID                   `json:"questionnaire_id"`
	QuestionnaireName string                      `json:"questionnaire_name"`
	Student           StudentBasicInfo            `json:"student"`
	Answers           []DetailedAnswerDTO         `json:"answers"`
	TotalScore        int                         `json:"total_score"`
	MaxScore          int                         `json:"max_score"`
	ScorePercentage   float64                     `json:"score_percentage"`
	Recommendations   []DetailedRecommendationDTO `json:"recommendations"`
	Analysis          *AnalysisResultDTO          `json:"analysis,omitempty"`
	ProcessingStatus  string                      `json:"processing_status"`
	SubmittedAt       time.Time                   `json:"submitted_at"`
}

type StudentBasicInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	NIM   string    `json:"nim"`
}

type DetailedAnswerDTO struct {
	QuestionID   uuid.UUID `json:"question_id"`
	QuestionText string    `json:"question_text"`
	Answer       string    `json:"answer"`
	Score        int       `json:"score"`
	MaxScore     int       `json:"max_score"`
	Category     string    `json:"category"`
}

type DetailedRecommendationDTO struct {
	ID            uuid.UUID `json:"id"`
	RoleName      string    `json:"role_name"`
	Score         float64   `json:"score"`
	Justification string    `json:"justification"`
	Category      string    `json:"category"`
	Active        bool      `json:"active"`
}

type AnalysisResultDTO struct {
	PersonalityTraits []string `json:"personality_traits"`
	Interests         []string `json:"interests"`
	Strengths         []string `json:"strengths"`
	WorkStyle         string   `json:"work_style"`
}

// Pagination DTOs
type PaginatedTargetRolesResponse struct {
	Data       []TargetRoleResponse `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}

type PaginatedQuestionnairesResponse struct {
	Data       []QuestionnaireListResponse `json:"data"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	Limit      int                         `json:"limit"`
	TotalPages int                         `json:"total_pages"`
}

type PaginatedResponsesResponse struct {
	Data       []QuestionnaireResponseOverview `json:"data"`
	Total      int64                           `json:"total"`
	Page       int                             `json:"page"`
	Limit      int                             `json:"limit"`
	TotalPages int                             `json:"total_pages"`
}
