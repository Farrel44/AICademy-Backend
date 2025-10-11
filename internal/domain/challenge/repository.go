package challenge

import (
	"strings"
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

type TeamMemberInfo struct {
	StudentProfileID uuid.UUID `json:"student_profile_id"`
	MemberRole       *string   `json:"member_role"`
	FullName         string    `json:"full_name"`
	NIS              string    `json:"nis"`
}

type StudentSearchResult struct {
	ID             uuid.UUID `json:"id"`
	NIS            string    `json:"nis"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	ProfilePicture *string   `json:"profile_picture"`
	Class          string    `json:"class"`
}

func NewChallengeRepository(db *gorm.DB, rdb *redis.Client) *ChallengeRepository {
	return &ChallengeRepository{
		db:           db,
		rdb:          rdb,
		cacheVersion: "v1",
		cacheTTL:     5 * time.Minute,
	}
}

// Student Profile operations
func (r *ChallengeRepository) GetStudentProfileByUserID(userID uuid.UUID) (*uuid.UUID, error) {
	var studentProfile struct {
		ID uuid.UUID `gorm:"column:id"`
	}

	err := r.db.Table("student_profiles").
		Select("id").
		Where("user_id = ?", userID).
		First(&studentProfile).Error
	if err != nil {
		return nil, err
	}
	return &studentProfile.ID, nil
}

func (r *ChallengeRepository) SearchStudents(query string, limit int, excludeUserID uuid.UUID) ([]StudentSearchResult, error) {
	var students []StudentSearchResult
	err := r.db.Table("users").
		Select(`
            student_profiles.id,
            users.nis,
            users.full_name,
            users.email,
            student_profiles.profile_picture,
            users.class
        `).
		Joins("JOIN student_profiles ON users.id = student_profiles.user_id").
		Where("users.role = ? AND users.id != ?", "student", excludeUserID).
		Where(`
            users.nis ILIKE ? OR 
            users.full_name ILIKE ? OR 
            users.email ILIKE ?
        `, "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Scan(&students).Error
	return students, err
}

func (r *ChallengeRepository) ValidateStudentProfileIDs(profileIDs []uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Table("student_profiles").
		Where("id IN ?", profileIDs).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == int64(len(profileIDs)), nil
}

// Team operations
func (r *ChallengeRepository) CreateTeam(team *Team) error {
	return r.db.Create(team).Error
}

func (r *ChallengeRepository) CreateTeamMembers(members []TeamMember) error {
	return r.db.Create(&members).Error
}

func (r *ChallengeRepository) GetTeamsByStudentProfileID(studentProfileID uuid.UUID) ([]Team, error) {
	var teams []Team
	err := r.db.
		Preload("Members").
		Where(`
            id IN (
                SELECT team_id FROM team_members 
                WHERE student_profile_id = ?
            )
        `, studentProfileID).
		Find(&teams).Error
	return teams, err
}

// Add this method to get team members with user details
func (r *ChallengeRepository) GetTeamMembersWithDetails(teamID uuid.UUID) ([]TeamMemberInfo, error) {
	var memberInfos []TeamMemberInfo

	query := `
        SELECT 
            tm.student_profile_id,
            tm.member_role,
            sp.fullname as full_name,
            sp.nis
        FROM team_members tm
        JOIN student_profiles sp ON tm.student_profile_id = sp.id
        WHERE tm.team_id = ?
        ORDER BY 
            CASE WHEN tm.member_role = 'Leader' THEN 0 ELSE 1 END,
            tm.joined_at
    `

	err := r.db.Raw(query, teamID).Scan(&memberInfos).Error
	return memberInfos, err
}

func (r *ChallengeRepository) GetTeamByIDAndMember(teamID, studentProfileID uuid.UUID) (*Team, error) {
	var team Team
	err := r.db.
		Preload("Members").
		Where(`
            id = ? AND id IN (
                SELECT team_id FROM team_members 
                WHERE student_profile_id = ?
            )
        `, teamID, studentProfileID).
		First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// Challenge operations
func (r *ChallengeRepository) CreateChallenge(challenge *Challenge) error {
	return r.db.Create(challenge).Error
}

func (r *ChallengeRepository) UpdateChallenge(challenge *Challenge) error {
	return r.db.Save(challenge).Error
}

func (r *ChallengeRepository) DeleteChallenge(id uuid.UUID) error {
	return r.db.Delete(&Challenge{}, "id = ?", id).Error
}

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

func (r *ChallengeRepository) GetAllChallengesWithSearch(offset, limit int, search string) ([]Challenge, int64, error) {
	var challenges []Challenge
	var total int64

	query := r.db.Model(&Challenge{}).
		Preload("Submissions").
		Preload("Submissions.Team").
		Preload("Submissions.Team.Members").
		Preload("WinnerTeam")

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&challenges).Error
	return challenges, total, err
}

// Optimized methods for search performance
func (r *ChallengeRepository) CountChallenges(search string) (int64, error) {
	var total int64
	query := r.db.Model(&Challenge{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	err := query.Count(&total).Error
	return total, err
}

func (r *ChallengeRepository) GetChallengesOptimized(offset, limit int, search string) ([]Challenge, error) {
	var challenges []Challenge
	query := r.db.Select("challenges.*").
		Model(&Challenge{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(challenges.title) LIKE ? OR LOWER(challenges.description) LIKE ?", searchTerm, searchTerm)
	}

	err := query.Offset(offset).Limit(limit).Order("challenges.created_at DESC").Find(&challenges).Error
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

// Submission operations
func (r *ChallengeRepository) CreateSubmission(submission *Submission) error {
	return r.db.Create(submission).Error
}

func (r *ChallengeRepository) GetSubmissionByTeamAndChallenge(teamID, challengeID uuid.UUID) (*Submission, error) {
	var submission Submission
	err := r.db.Where("challenge_id = ? AND team_id = ?", challengeID, teamID).
		First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
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

func (r *ChallengeRepository) GetAllSubmissionsOptimized(offset, limit int, search string) ([]Submission, int64, error) {
	var submissions []Submission
	var total int64

	// Separate count and data queries for optimization
	countQuery := r.db.Model(&Submission{})
	dataQuery := r.db.Model(&Submission{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		searchCondition := "LOWER(submissions.title) LIKE ? OR LOWER(c.title) LIKE ? OR LOWER(t.team_name) LIKE ?"

		countQuery = countQuery.
			Joins("LEFT JOIN challenges c ON submissions.challenge_id = c.id").
			Joins("LEFT JOIN teams t ON submissions.team_id = t.id").
			Where(searchCondition, searchTerm, searchTerm, searchTerm)

		dataQuery = dataQuery.
			Joins("LEFT JOIN challenges c ON submissions.challenge_id = c.id").
			Joins("LEFT JOIN teams t ON submissions.team_id = t.id").
			Where(searchCondition, searchTerm, searchTerm, searchTerm)
	}

	// Get count
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get data
	err := dataQuery.
		Preload("Challenge").
		Preload("Team").
		Order("submissions.submitted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&submissions).Error

	return submissions, total, err
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

func (r *ChallengeRepository) GetChallengesByTeacherOptimized(teacherID uuid.UUID, offset, limit int, search string) ([]Challenge, int64, error) {
	var challenges []Challenge
	var total int64

	query := r.db.Model(&Challenge{}).Where("created_by_teacher_id = ?", teacherID)
	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("WinnerTeam").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&challenges).Error

	return challenges, total, err
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

func (r *ChallengeRepository) GetSubmissionsByTeacherOptimized(teacherID uuid.UUID, offset, limit int, search string, challengeID *uuid.UUID) ([]Submission, int64, error) {
	var submissions []Submission
	var total int64

	// Separate count query for optimization
	countQuery := r.db.Model(&Submission{}).
		Joins("JOIN challenges c ON submissions.challenge_id = c.id").
		Where("c.created_by = ?", teacherID)

	// Apply challenge filter
	if challengeID != nil {
		countQuery = countQuery.Where("submissions.challenge_id = ?", *challengeID)
	}

	// Apply search filter with case-insensitive search
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		countQuery = countQuery.Where("LOWER(c.title) LIKE ? OR LOWER(c.description) LIKE ? OR LOWER(submissions.description) LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Get count
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Data query with preloading
	dataQuery := r.db.
		Preload("Challenge", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, title").Preload("WinnerTeam", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, team_name")
			})
		}).
		Preload("Team", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, team_name").Preload("Members", func(db *gorm.DB) *gorm.DB {
				return db.Select("team_id, student_profile_id, member_role")
			})
		}).
		Joins("JOIN challenges c ON submissions.challenge_id = c.id").
		Where("c.created_by = ?", teacherID)

	// Apply challenge filter
	if challengeID != nil {
		dataQuery = dataQuery.Where("submissions.challenge_id = ?", *challengeID)
	}

	// Apply search filter
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		dataQuery = dataQuery.Where("LOWER(submissions.title) LIKE ?", searchTerm)
	}

	err := dataQuery.Offset(offset).Limit(limit).Order("submissions.submitted_at DESC").Find(&submissions).Error
	return submissions, total, err
}

// Auto announce winner for challenges that passed 7 days after deadline
func (r *ChallengeRepository) AutoAnnounceWinner(challengeID uuid.UUID) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get top team by points for this challenge
	var topSubmission struct {
		TeamID uuid.UUID `gorm:"column:team_id"`
		Points int       `gorm:"column:points"`
	}

	err := tx.Table("submissions").
		Select("team_id, points").
		Where("challenge_id = ? AND team_id IS NOT NULL AND points IS NOT NULL", challengeID).
		Order("points DESC").
		First(&topSubmission).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	// Update challenge with winner
	err = tx.Model(&Challenge{}).
		Where("id = ?", challengeID).
		Update("winner_team_id", topSubmission.TeamID).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	// Create challenge winner record
	challengeWinner := ChallengeWinner{
		ChallengeID: challengeID,
		TeamID:      topSubmission.TeamID,
		Position:    1,
	}

	err = tx.Create(&challengeWinner).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Check and auto announce winners for eligible challenges
func (r *ChallengeRepository) CheckAndAutoAnnounceWinners() error {
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)

	var eligibleChallenges []Challenge
	err := r.db.Where("deadline < ? AND winner_team_id IS NULL", sevenDaysAgo).
		Find(&eligibleChallenges).Error

	if err != nil {
		return err
	}

	for _, challenge := range eligibleChallenges {
		// Skip if no submissions with points
		var submissionCount int64
		r.db.Model(&Submission{}).
			Where("challenge_id = ? AND team_id IS NOT NULL AND points IS NOT NULL AND points > 0", challenge.ID).
			Count(&submissionCount)

		if submissionCount > 0 {
			r.AutoAnnounceWinner(challenge.ID)
		}
	}

	return nil
}

// Get challenge by ID with auto winner check
func (r *ChallengeRepository) GetChallengeByIDWithWinnerCheck(id uuid.UUID) (*Challenge, error) {
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

	// Check if this challenge needs winner announcement
	if challenge.NeedsWinnerAnnouncement() {
		r.AutoAnnounceWinner(challenge.ID)
		// Reload challenge to get updated winner info
		r.db.
			Preload("Submissions").
			Preload("Submissions.Team").
			Preload("Submissions.Team.Members").
			Preload("WinnerTeam").
			Preload("ChallengeWinners").
			First(&challenge, "id = ?", id)
	}

	return &challenge, nil
}

// Add missing method for student
func (r *ChallengeRepository) GetActiveTeamByStudentID(studentProfileID uuid.UUID) (*Team, error) {
	var team Team
	err := r.db.
		Preload("Members").
		Where(`
            id IN (
                SELECT team_id FROM team_members 
                WHERE student_profile_id = ?
            )
        `, studentProfileID).
		Order("created_at DESC").
		First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *ChallengeRepository) UpdateSubmission(submission *Submission) error {
	return r.db.Save(submission).Error
}

func (r *ChallengeRepository) CountSubmissionsByStudentID(studentProfileID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&Submission{}).
		Joins("JOIN team_members ON submissions.team_id = team_members.team_id").
		Where("team_members.student_profile_id = ?", studentProfileID).
		Count(&count).Error
	return count, err
}

func (r *ChallengeRepository) GetSubmissionsByStudentID(studentProfileID uuid.UUID, offset, limit int) ([]Submission, error) {
	var submissions []Submission
	err := r.db.
		Preload("Challenge").
		Preload("Team").
		Joins("JOIN team_members ON submissions.team_id = team_members.team_id").
		Where("team_members.student_profile_id = ?", studentProfileID).
		Offset(offset).
		Limit(limit).
		Order("submissions.submitted_at DESC").
		Find(&submissions).Error
	return submissions, err
}

func (r *ChallengeRepository) GetSubmissionsByChallengeOptimized(challengeID uuid.UUID, offset, limit int, search string) ([]Submission, int64, error) {
	var submissions []Submission
	var total int64

	// Base query for the challenge
	baseQuery := r.db.Model(&Submission{}).Where("challenge_id = ?", challengeID)

	// Apply search filter if provided
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		baseQuery = baseQuery.
			Joins("LEFT JOIN teams t ON submissions.team_id = t.id").
			Where("LOWER(submissions.title) LIKE ? OR LOWER(t.team_name) LIKE ?", searchTerm, searchTerm)
	}

	// Get count
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get data with preloading
	dataQuery := r.db.Model(&Submission{}).
		Preload("Challenge").
		Preload("Team").
		Where("challenge_id = ?", challengeID)

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		dataQuery = dataQuery.
			Joins("LEFT JOIN teams t ON submissions.team_id = t.id").
			Where("LOWER(submissions.title) LIKE ? OR LOWER(t.team_name) LIKE ?", searchTerm, searchTerm)
	}

	err := dataQuery.
		Order("submissions.submitted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&submissions).Error

	return submissions, total, err
}

func (r *ChallengeRepository) UpdateChallengeParticipants(challengeID uuid.UUID, increment bool) error {
	if increment {
		return r.db.Model(&Challenge{}).
			Where("id = ?", challengeID).
			Update("current_participants", gorm.Expr("current_participants + 1")).Error
	} else {
		return r.db.Model(&Challenge{}).
			Where("id = ?", challengeID).
			Update("current_participants", gorm.Expr("current_participants - 1")).Error
	}
}

func (r *ChallengeRepository) IsStudentInChallengeTeam(studentProfileID, challengeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&Submission{}).
		Joins("JOIN team_members tm ON submissions.team_id = tm.team_id").
		Where("submissions.challenge_id = ? AND tm.student_profile_id = ?", challengeID, studentProfileID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
