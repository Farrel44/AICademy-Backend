package ai

type CareerAnalysisResponse struct {
	Analysis struct {
		PersonalityTraits []string `json:"personality_traits"`
		Interests         []string `json:"interests"`
		Strengths         []string `json:"strengths"`
		WorkStyle         string   `json:"work_style"`
	} `json:"analysis"`
	Recommendations []CareerRecommendation `json:"recommendations"`
}

type CareerRecommendation struct {
	RoleID        string  `json:"role_id"`
	RoleName      string  `json:"role_name"`
	Score         float64 `json:"score"`
	Justification string  `json:"justification"`
}

type QuestionGenerationResponse struct {
	Questions []GeneratedQuestion `json:"questions"`
	Metadata  struct {
		TotalQuestions      int            `json:"total_questions"`
		Distribution        map[string]int `json:"distribution"`
		TargetRolesCoverage []string       `json:"target_roles_coverage"`
	} `json:"metadata"`
}

type GeneratedQuestion struct {
	QuestionText string        `json:"question_text"`
	QuestionType string        `json:"question_type"`
	Options      []QuestionOpt `json:"options,omitempty"`
	Category     string        `json:"category"`
	Reasoning    string        `json:"reasoning"`
}

type QuestionOpt struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
