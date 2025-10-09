package student_challenge

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/challenge"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentChallengeService struct {
	repo   *challenge.ChallengeRepository
	userDB *gorm.DB // For searching students
}

func NewStudentChallengeService(repo *challenge.ChallengeRepository, db *gorm.DB) *StudentChallengeService {
	return &StudentChallengeService{
		repo:   repo,
		userDB: db,
	}
}

// Search students by NIS, name, or email
func (s *StudentChallengeService) SearchStudents(c *fiber.Ctx, req *SearchStudentRequest) ([]StudentSearchResult, error) {
	// Verify student access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "student" {
		return nil, errors.New("access denied: student role required")
	}

	var students []StudentSearchResult
	query := strings.TrimSpace(req.Query)

	// Search in users table with student_profiles
	err = s.userDB.Table("users").
		Select(`
            student_profiles.id,
            users.nis,
            users.full_name,
            users.email,
            student_profiles.profile_picture,
            users.class
        `).
		Joins("JOIN student_profiles ON users.id = student_profiles.user_id").
		Where("users.role = ? AND users.id != ?", "student", claims.UserID).
		Where(`
            users.nis ILIKE ? OR 
            users.full_name ILIKE ? OR 
            users.email ILIKE ?
        `, "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Limit(req.Limit).
		Scan(&students).Error

	if err != nil {
		return nil, errors.New("failed to search students")
	}

	return students, nil
}

// Create team
func (s *StudentChallengeService) CreateTeam(c *fiber.Ctx, req *CreateTeamRequest) (*CreateTeamResponse, error) {
	// Verify student access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "student" {
		return nil, errors.New("access denied: student role required")
	}

	// Get student profile ID
	var studentProfileID uuid.UUID
	err = s.userDB.Table("student_profiles").
		Select("id").
		Where("user_id = ?", claims.UserID).
		Scan(&studentProfileID).Error
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// Validate member count (creator + 2 others = 3 total)
	if len(req.MemberIDs) != 2 {
		return nil, errors.New("team must have exactly 3 members (including creator)")
	}

	// Validate all member IDs exist and are students
	var memberCount int64
	err = s.userDB.Table("student_profiles").
		Where("id IN ?", req.MemberIDs).
		Count(&memberCount).Error
	if err != nil || memberCount != 2 {
		return nil, errors.New("invalid member IDs provided")
	}

	// Check if creator is not in member list
	for _, memberID := range req.MemberIDs {
		if memberID == studentProfileID {
			return nil, errors.New("creator cannot be added as a member")
		}
	}

	// Create team
	team := &challenge.Team{
		TeamName:                  req.TeamName,
		About:                     req.About,
		TeamProfilePicture:        req.TeamProfilePicture,
		CreatedByStudentProfileID: studentProfileID,
		CreatedAt:                 time.Now(),
	}

	// Start transaction
	tx := s.userDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create team
	if err := tx.Create(team).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create team")
	}

	// Add creator as first member
	creatorMember := challenge.TeamMember{
		TeamID:           team.ID,
		StudentProfileID: studentProfileID,
		MemberRole:       stringPtr("Leader"),
		JoinedAt:         time.Now(),
	}
	if err := tx.Create(&creatorMember).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to add creator to team")
	}

	// Add other members
	var allMembers []TeamMemberInfo
	allMembers = append(allMembers, TeamMemberInfo{
		StudentProfileID: studentProfileID,
		MemberRole:       stringPtr("Leader"),
	})

	for _, memberID := range req.MemberIDs {
		member := challenge.TeamMember{
			TeamID:           team.ID,
			StudentProfileID: memberID,
			MemberRole:       stringPtr("Member"),
			JoinedAt:         time.Now(),
		}
		if err := tx.Create(&member).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("failed to add member to team")
		}

		allMembers = append(allMembers, TeamMemberInfo{
			StudentProfileID: memberID,
			MemberRole:       stringPtr("Member"),
		})
	}

	tx.Commit()

	return &CreateTeamResponse{
		ID:                 team.ID,
		TeamName:           team.TeamName,
		About:              team.About,
		TeamProfilePicture: team.TeamProfilePicture,
		CreatedAt:          team.CreatedAt,
		Members:            allMembers,
		MemberCount:        len(allMembers),
	}, nil
}

// Get my teams
func (s *StudentChallengeService) GetMyTeams(c *fiber.Ctx) ([]challenge.Team, error) {
	// Verify student access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "student" {
		return nil, errors.New("access denied: student role required")
	}

	// Get student profile ID
	var studentProfileID uuid.UUID
	err = s.userDB.Table("student_profiles").
		Select("id").
		Where("user_id = ?", claims.UserID).
		Scan(&studentProfileID).Error
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	var teams []challenge.Team
	err = s.userDB.
		Preload("Members").
		Where(`
            id IN (
                SELECT team_id FROM team_members 
                WHERE student_profile_id = ?
            )
        `, studentProfileID).
		Find(&teams).Error

	if err != nil {
		return nil, errors.New("failed to fetch teams")
	}

	return teams, nil
}

// Get available challenges
func (s *StudentChallengeService) GetAvailableChallenges(c *fiber.Ctx) ([]ChallengeListResponse, error) {
	// Verify student access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "student" {
		return nil, errors.New("access denied: student role required")
	}

	challenges, err := s.repo.GetAllChallenges()
	if err != nil {
		return nil, errors.New("failed to fetch challenges")
	}

	var result []ChallengeListResponse
	for _, ch := range challenges {
		result = append(result, ChallengeListResponse{
			ID:                  ch.ID,
			ThumbnailImage:      ch.ThumbnailImage,
			Title:               ch.Title,
			Description:         ch.Description,
			Deadline:            ch.Deadline,
			Prize:               ch.Prize,
			MaxParticipants:     ch.MaxParticipants,
			CurrentParticipants: ch.CurrentParticipants,
			CanRegister:         ch.CanRegister(),
			IsActive:            ch.IsActive(),
			OrganizerName:       ch.GetOrganizerName(),
		})
	}

	return result, nil
}

// Register team to challenge
func (s *StudentChallengeService) RegisterTeamToChallenge(c *fiber.Ctx, req *RegisterChallengeRequest) (*RegisterChallengeResponse, error) {
	// Verify student access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "student" {
		return nil, errors.New("access denied: student role required")
	}

	// Get student profile ID
	var studentProfileID uuid.UUID
	err = s.userDB.Table("student_profiles").
		Select("id").
		Where("user_id = ?", claims.UserID).
		Scan(&studentProfileID).Error
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// Verify team exists and student is a member
	var team challenge.Team
	err = s.userDB.
		Preload("Members").
		Where(`
            id = ? AND id IN (
                SELECT team_id FROM team_members 
                WHERE student_profile_id = ?
            )
        `, req.TeamID, studentProfileID).
		First(&team).Error
	if err != nil {
		return nil, errors.New("team not found or you're not a member")
	}

	// Validate team has exactly 3 members
	if !team.ValidateTeamSize() {
		return nil, errors.New("team must have exactly 3 members to register")
	}

	// Get challenge and verify it's available
	challengeData, err := s.repo.GetChallengeByID(req.ChallengeID)
	if err != nil {
		return nil, errors.New("challenge not found")
	}

	if !challengeData.CanRegister() {
		return nil, errors.New("challenge registration is not available")
	}

	// Check if team already registered for this challenge
	var existingSubmission challenge.Submission
	err = s.userDB.Where("challenge_id = ? AND team_id = ?", req.ChallengeID, req.TeamID).
		First(&existingSubmission).Error
	if err == nil {
		return nil, errors.New("team already registered for this challenge")
	}

	// Create submission entry (registration)
	submission := &challenge.Submission{
		ChallengeID: req.ChallengeID,
		TeamID:      &req.TeamID,
		Title:       fmt.Sprintf("%s - %s", team.TeamName, challengeData.Title),
		SubmittedAt: time.Now(),
	}

	// Start transaction
	tx := s.userDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create submission
	if err := tx.Create(submission).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to register team")
	}

	// Update challenge participant count
	if err := s.repo.UpdateChallengeParticipants(req.ChallengeID, true); err != nil {
		tx.Rollback()
		return nil, errors.New("failed to update participant count")
	}

	tx.Commit()

	return &RegisterChallengeResponse{
		Message:             "Team successfully registered for challenge",
		ChallengeID:         req.ChallengeID,
		TeamID:              req.TeamID,
		RegistrationDate:    submission.SubmittedAt,
		CurrentParticipants: challengeData.CurrentParticipants + 1,
	}, nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
