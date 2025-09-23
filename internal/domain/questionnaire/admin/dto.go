package admin

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Predefined Focus Areas
var (
	// Technical Focus Areas
	TechnicalFocusAreas = []string{
		"programming", "software_development", "web_development", "mobile_development",
		"data_science", "machine_learning", "artificial_intelligence", "cybersecurity",
		"cloud_computing", "devops", "database_management", "ui_ux_design",
	}

	// Business Focus Areas
	BusinessFocusAreas = []string{
		"project_management", "business_analysis", "marketing", "sales",
		"finance", "operations", "human_resources", "strategy",
		"entrepreneurship", "consulting", "product_management",
	}

	// Industry Focus Areas
	IndustryFocusAreas = []string{
		"fintech", "healthcare", "education", "e_commerce", "gaming",
		"media", "manufacturing", "logistics", "agriculture", "energy",
		"telecommunications", "government", "non_profit",
	}

	// Soft Skills Focus Areas
	SoftSkillsFocusAreas = []string{
		"leadership", "communication", "teamwork", "problem_solving",
		"creativity", "adaptability", "time_management", "analytical_thinking",
		"customer_service", "negotiation", "presentation", "mentoring",
	}
)

// Helper functions for Focus Areas
func GetAllFocusAreas() map[string][]string {
	return map[string][]string{
		"technical":   TechnicalFocusAreas,
		"business":    BusinessFocusAreas,
		"industry":    IndustryFocusAreas,
		"soft_skills": SoftSkillsFocusAreas,
	}
}

func GetAvailableFocusAreas() []string {
	var all []string
	all = append(all, TechnicalFocusAreas...)
	all = append(all, BusinessFocusAreas...)
	all = append(all, IndustryFocusAreas...)
	all = append(all, SoftSkillsFocusAreas...)
	return all
}

// GetDefaultFocusAreas returns predefined focus areas based on role categories
func GetDefaultFocusAreas(roleNames []string) []string {
	focusAreas := make(map[string]bool)

	// Always include general assessment areas
	generalAreas := []string{"problem_solving", "analytical_thinking", "communication", "teamwork"}
	for _, area := range generalAreas {
		focusAreas[area] = true
	}

	// Add specific focus areas based on role patterns
	for _, roleName := range roleNames {
		switch {
		case contains(roleName, []string{"developer", "programmer", "engineer", "tech"}):
			techAreas := []string{"programming", "software_development", "technical_skills"}
			for _, area := range techAreas {
				focusAreas[area] = true
			}
		case contains(roleName, []string{"data", "analyst", "scientist"}):
			dataAreas := []string{"data_science", "analytical_thinking", "statistics"}
			for _, area := range dataAreas {
				focusAreas[area] = true
			}
		case contains(roleName, []string{"manager", "lead", "supervisor"}):
			mgmtAreas := []string{"leadership", "project_management", "team_management"}
			for _, area := range mgmtAreas {
				focusAreas[area] = true
			}
		case contains(roleName, []string{"design", "ui", "ux", "creative"}):
			designAreas := []string{"creativity", "ui_ux_design", "visual_thinking"}
			for _, area := range designAreas {
				focusAreas[area] = true
			}
		case contains(roleName, []string{"business", "marketing", "sales"}):
			bizAreas := []string{"business_analysis", "marketing", "customer_service"}
			for _, area := range bizAreas {
				focusAreas[area] = true
			}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(focusAreas))
	for area := range focusAreas {
		result = append(result, area)
	}

	return result
}

// Helper function to check if role name contains any of the keywords
func contains(roleName string, keywords []string) bool {
	roleNameLower := strings.ToLower(roleName)
	for _, keyword := range keywords {
		if strings.Contains(roleNameLower, keyword) {
			return true
		}
	}
	return false
}

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

// Focus Areas Response DTO
type FocusAreasResponse struct {
	Categories map[string][]string `json:"categories"`
	All        []string            `json:"all"`
}

// Questionnaire Generation DTOs
type GenerateQuestionnaireRequest struct {
	Name               string   `json:"name" validate:"required,min=2,max=100"`
	QuestionCount      int      `json:"question_count" validate:"required,min=3,max=20"`
	TargetRoleIDs      []string `json:"target_role_ids" validate:"required,min=1"`
	DifficultyLevel    string   `json:"difficulty_level" validate:"required,oneof=basic intermediate advanced"`
	CustomInstructions *string  `json:"custom_instructions,omitempty" validate:"omitempty,max=1000"`
	AIPersonality      string   `json:"ai_personality" validate:"omitempty,oneof=formal friendly professional academic creative"`
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
