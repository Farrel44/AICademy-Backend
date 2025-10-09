package challenge

import (
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ChallengeRepository struct {
	db           *gorm.DB
	rdb          *redis.Client
	cacheVersion string
	cacheTTL     time.Duration
}

func NewChallengeRepository(db *gorm.DB, rdb *redis.Client) *ChallengeRepository {
	return &ChallengeRepository{
		db:           db,
		rdb:          rdb,
		cacheVersion: "v1",
		cacheTTL:     5 * time.Minute,
	}
}

// CRUD Operations
func (r *ChallengeRepository) CreateChallenge(challenge *Challenge) error {
	return r.db.Create(challenge).Error
}

func (r *ChallengeRepository) UpdateChallenge(challenge *Challenge) error {
	return r.db.Save(challenge).Error
}

func (r *ChallengeRepository) DeleteChallenge(id uuid.UUID) error {
	return r.db.Delete(&Challenge{}, "id = ?", id).Error
}

// Admin methods
func (r *ChallengeRepository) GetAllChallenges() ([]Challenge, error) {
	var challenges []Challenge
	err := r.db.
		Preload("Submissions").
		Preload("Submissions.Team").
		Preload("Submissions.Team.Members").
		Preload("WinnerTeam").
		Find(&challenges).Error
	return challenges, err
}

func (r *ChallengeRepository) GetChallengeByID(id uuid.UUID) (*Challenge, error) {
	var challenge Challenge
	err := r.db.
		Preload("Submissions").
		Preload("Submissions.Team").
		Preload("Submissions.Team.Members").
		Preload("WinnerTeam").
		Preload("ChallengeWinners").
		First(&challenge, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *ChallengeRepository) GetAllSubmissions() ([]Submission, error) {
	var submissions []Submission
	err := r.db.
		Preload("Challenge").
		Preload("Team").
		Preload("Team.Members").
		Find(&submissions).Error
	return submissions, err
}

func (r *ChallengeRepository) GetSubmissionsByChallenge(challengeID uuid.UUID) ([]Submission, error) {
	var submissions []Submission
	err := r.db.
		Preload("Team").
		Preload("Team.Members").
		Where("challenge_id = ?", challengeID).
		Find(&submissions).Error
	return submissions, err
}

func (r *ChallengeRepository) ScoreSubmission(submissionID uuid.UUID, points int, scorerID uuid.UUID, isAdmin bool) error {
	updates := map[string]interface{}{
		"points":    points,
		"scored_at": time.Now(),
	}

	if isAdmin {
		updates["scored_by_admin_id"] = scorerID
	} else {
		updates["scored_by_teacher_id"] = scorerID
	}

	return r.db.Model(&Submission{}).
		Where("id = ?", submissionID).
		Updates(updates).Error
}

// Teacher methods
func (r *ChallengeRepository) GetChallengesByTeacher(teacherID uuid.UUID) ([]Challenge, error) {
	var challenges []Challenge
	err := r.db.
		Preload("Submissions").
		Preload("Submissions.Team").
		Preload("Submissions.Team.Members").
		Preload("WinnerTeam").
		Where("created_by_teacher_id = ?", teacherID).
		Find(&challenges).Error
	return challenges, err
}

func (r *ChallengeRepository) GetTeacherChallengeByID(challengeID, teacherID uuid.UUID) (*Challenge, error) {
	var challenge Challenge
	err := r.db.
		Preload("Submissions").
		Preload("Submissions.Team").
		Preload("Submissions.Team.Members").
		Preload("WinnerTeam").
		Preload("ChallengeWinners").
		Where("id = ? AND created_by_teacher_id = ?", challengeID, teacherID).
		First(&challenge).Error
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *ChallengeRepository) GetAdminChallengeByID(challengeID, adminID uuid.UUID) (*Challenge, error) {
	var challenge Challenge
	err := r.db.
		Preload("Submissions").
		Preload("Submissions.Team").
		Preload("Submissions.Team.Members").
		Preload("WinnerTeam").
		Preload("ChallengeWinners").
		Where("id = ? AND created_by_admin_id = ?", challengeID, adminID).
		First(&challenge).Error
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

// Leaderboard methods
func (r *ChallengeRepository) GetLeaderboard(challengeID uuid.UUID) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry

	query := `
        SELECT 
            s.team_id,
            t.team_name,
            COALESCE(s.points, 0) as total_points,
            s.challenge_id,
            s.id as submission_id,
            ROW_NUMBER() OVER (ORDER BY COALESCE(s.points, 0) DESC) as position
        FROM submissions s
        JOIN teams t ON s.team_id = t.id
        WHERE s.challenge_id = ? AND s.points IS NOT NULL
        ORDER BY total_points DESC
    `

	err := r.db.Raw(query, challengeID).Scan(&entries).Error
	return entries, err
}

func (r *ChallengeRepository) GetGlobalLeaderboard() ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry

	query := `
        SELECT 
            s.team_id,
            t.team_name,
            SUM(COALESCE(s.points, 0)) as total_points,
            ROW_NUMBER() OVER (ORDER BY SUM(COALESCE(s.points, 0)) DESC) as position
        FROM submissions s
        JOIN teams t ON s.team_id = t.id
        WHERE s.points IS NOT NULL
        GROUP BY s.team_id, t.team_name
        ORDER BY total_points DESC
    `

	err := r.db.Raw(query).Scan(&entries).Error
	return entries, err
}

func (r *ChallengeRepository) UpdateChallengeParticipants(challengeID uuid.UUID, increment bool) error {
	var change int
	if increment {
		change = 1
	} else {
		change = -1
	}

	return r.db.Model(&Challenge{}).
		Where("id = ?", challengeID).
		Update("current_participants", gorm.Expr("current_participants + ?", change)).Error
}
