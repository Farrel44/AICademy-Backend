package challenge

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Team struct {
	ID                        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TeamName                  string    `gorm:"type:varchar;not null" json:"team_name"`
	About                     *string   `gorm:"type:text" json:"about"`
	TeamProfilePicture        *string   `gorm:"type:text" json:"team_profile_picture"`
	CreatedByStudentProfileID uuid.UUID `gorm:"type:uuid;not null;index" json:"created_by_student_profile_id"`
	CreatedAt                 time.Time `gorm:"type:timestamptz" json:"created_at"`

	Members          []TeamMember      `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"members,omitempty"`
	WonChallenges    []Challenge       `gorm:"foreignKey:WinnerTeamID" json:"won_challenges,omitempty"`
	ChallengeWinners []ChallengeWinner `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"challenge_winners,omitempty"`
	Submissions      []Submission      `gorm:"foreignKey:TeamID;constraint:OnDelete:SET NULL" json:"submissions,omitempty"`
}

func (t *Team) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// Validate team has exactly 3 members
func (t *Team) ValidateTeamSize() bool {
	return len(t.Members) == 3
}

type TeamMember struct {
	TeamID           uuid.UUID  `gorm:"type:uuid;not null;primaryKey" json:"team_id"`
	StudentProfileID uuid.UUID  `gorm:"type:uuid;not null;primaryKey" json:"student_profile_id"`
	MemberRole       *string    `gorm:"type:varchar" json:"member_role"`
	ProfilingRoleID  *uuid.UUID `gorm:"type:uuid;index" json:"profiling_role_id"`
	JoinedAt         time.Time  `gorm:"type:timestamptz" json:"joined_at"`

	Team *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

func (TeamMember) TableName() string {
	return "team_members"
}

type Challenge struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ThumbnailImage      string     `gorm:"type:text;not null" json:"thumbnail_image"`
	Title               string     `gorm:"type:varchar;not null" json:"title"`
	Description         string     `gorm:"type:text;not null" json:"description"`
	Deadline            time.Time  `gorm:"type:timestamptz;not null" json:"deadline"`
	Prize               *string    `gorm:"type:varchar" json:"prize"`
	MaxParticipants     int        `gorm:"default:0" json:"max_participants"`
	CurrentParticipants int        `gorm:"default:0" json:"current_participants"`
	WinnerTeamID        *uuid.UUID `gorm:"type:uuid;index" json:"winner_team_id"`
	// New fields for organizer tracking
	CreatedByAdminID   *uuid.UUID `gorm:"type:uuid;index" json:"created_by_admin_id"`
	CreatedByTeacherID *uuid.UUID `gorm:"type:uuid;index" json:"created_by_teacher_id"`
	CreatedAt          time.Time  `gorm:"type:timestamptz" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"type:timestamptz" json:"updated_at"`

	WinnerTeam       *Team             `gorm:"foreignKey:WinnerTeamID" json:"winner_team,omitempty"`
	ChallengeWinners []ChallengeWinner `gorm:"foreignKey:ChallengeID;constraint:OnDelete:CASCADE" json:"challenge_winners,omitempty"`
	Submissions      []Submission      `gorm:"foreignKey:ChallengeID;constraint:OnDelete:CASCADE" json:"submissions,omitempty"`
	Judges           []ChallengeJudge  `gorm:"foreignKey:ChallengeID;constraint:OnDelete:CASCADE" json:"judges,omitempty"`
}

func (c *Challenge) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (c *Challenge) IsActive() bool {
	return time.Now().Before(c.Deadline)
}

func (c *Challenge) IsExpired() bool {
	return time.Now().After(c.Deadline)
}

func (c *Challenge) CanRegister() bool {
	return c.IsActive() && c.CurrentParticipants < c.MaxParticipants
}

func (c *Challenge) GetOrganizerName() string {
	if c.CreatedByAdminID != nil {
		return "Admin"
	}
	return "Teacher"
}

type ChallengeWinner struct {
	ChallengeID uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"challenge_id"`
	TeamID      uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"team_id"`
	Position    int       `gorm:"not null" json:"position"`

	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
	Team      *Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

func (ChallengeWinner) TableName() string {
	return "challenge_winners"
}

// Submission model
type Submission struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ChallengeID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"challenge_id"`
	Title            string     `gorm:"type:varchar;not null" json:"title"`
	TeamID           *uuid.UUID `gorm:"type:uuid;index" json:"team_id"`
	StudentProfileID *uuid.UUID `gorm:"type:uuid;index" json:"student_profile_id"`
	ImageURL         *string    `gorm:"type:text" json:"image_url"`
	RepoURL          *string    `gorm:"type:text" json:"repo_url"`
	DocsURL          *string    `gorm:"type:text" json:"docs_url"`
	SubmittedAt      time.Time  `gorm:"type:timestamptz;not null" json:"submitted_at"`
	Points           *int       `gorm:"default:0" json:"points"`
	// New fields for tracking who scored
	ScoredByAdminID   *uuid.UUID `gorm:"type:uuid;index" json:"scored_by_admin_id"`
	ScoredByTeacherID *uuid.UUID `gorm:"type:uuid;index" json:"scored_by_teacher_id"`
	ScoredAt          *time.Time `gorm:"type:timestamptz" json:"scored_at"`

	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
	Team      *Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

func (s *Submission) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *Submission) IsTeamSubmission() bool {
	return s.TeamID != nil
}

func (s *Submission) IsIndividualSubmission() bool {
	return s.StudentProfileID != nil && s.TeamID == nil
}

func (s *Submission) IsScored() bool {
	return s.Points != nil && *s.Points > 0
}

// Challenge Judge model
type ChallengeJudge struct {
	ChallengeID      uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"challenge_id"`
	TeacherProfileID uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"teacher_profile_id"`

	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
}

func (ChallengeJudge) TableName() string {
	return "challenge_judges"
}

// Leaderboard entry
type LeaderboardEntry struct {
	TeamID       uuid.UUID `json:"team_id"`
	TeamName     string    `json:"team_name"`
	TotalPoints  int       `json:"total_points"`
	Position     int       `json:"position"`
	ChallengeID  uuid.UUID `json:"challenge_id"`
	SubmissionID uuid.UUID `json:"submission_id"`
}
