package teacher

type TeacherChallengeStats struct {
	TotalChallenges     int `json:"total_challenges"`
	ActiveChallenges    int `json:"active_challenges"`
	CompletedChallenges int `json:"completed_challenges"`
}

type TeacherSubmissionStats struct {
	TotalSubmissions   int `json:"total_submissions"`
	ScoredSubmissions  int `json:"scored_submissions"`
	PendingSubmissions int `json:"pending_submissions"`
}

type TeacherDashboardData struct {
	ChallengeStats  TeacherChallengeStats  `json:"challenge_stats"`
	SubmissionStats TeacherSubmissionStats `json:"submission_stats"`
}
