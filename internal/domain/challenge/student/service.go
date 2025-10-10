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
)

type StudentChallengeService struct {
	repo  *challenge.ChallengeRepository
	redis *utils.RedisClient
}

func NewStudentChallengeService(repo *challenge.ChallengeRepository, redis *utils.RedisClient) *StudentChallengeService {
	return &StudentChallengeService{
		repo:  repo,
		redis: redis,
	}
}

func (s *StudentChallengeService) SearchStudents(c *fiber.Ctx, req *SearchStudentRequest) ([]challenge.StudentSearchResult, error) {
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "student" {
		return nil, errors.New("access denied: student role required")
	}

	query := strings.TrimSpace(req.Query)
	if req.Limit == 0 {
		req.Limit = 10
	}

	students, err := s.repo.SearchStudents(query, req.Limit, claims.UserID)
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
	studentProfileID, err := s.repo.GetStudentProfileByUserID(claims.UserID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// Validate member count (creator + 2 others = 3 total)
	if len(req.MemberIDs) != 2 {
		return nil, errors.New("team must have exactly 3 members (including creator)")
	}

	// Validate all member IDs exist and are students
	isValid, err := s.repo.ValidateStudentProfileIDs(req.MemberIDs)
	if err != nil || !isValid {
		return nil, errors.New("invalid member IDs provided")
	}

	// Check if creator is not in member list
	for _, memberID := range req.MemberIDs {
		if memberID == *studentProfileID {
			return nil, errors.New("creator cannot be added as a member")
		}
	}

	// Create team
	team := &challenge.Team{
		TeamName:                  req.TeamName,
		About:                     req.About,
		TeamProfilePicture:        req.TeamProfilePicture,
		CreatedByStudentProfileID: *studentProfileID,
		CreatedAt:                 time.Now(),
	}

	// Create team in repository
	if err := s.repo.CreateTeam(team); err != nil {
		return nil, errors.New("failed to create team")
	}

	// Prepare team members
	var members []challenge.TeamMember

	// Add creator as leader
	members = append(members, challenge.TeamMember{
		TeamID:           team.ID,
		StudentProfileID: *studentProfileID,
		MemberRole:       stringPtr("Leader"),
		JoinedAt:         time.Now(),
	})

	// Add other members
	for _, memberID := range req.MemberIDs {
		members = append(members, challenge.TeamMember{
			TeamID:           team.ID,
			StudentProfileID: memberID,
			MemberRole:       stringPtr("Member"),
			JoinedAt:         time.Now(),
		})
	}

	// Create team members in repository
	if err := s.repo.CreateTeamMembers(members); err != nil {
		return nil, errors.New("failed to add members to team")
	}

	// Get member details for response
	var memberInfos []TeamMemberInfo
	for _, member := range members {
		memberInfo := TeamMemberInfo{
			StudentProfileID: member.StudentProfileID,
			MemberRole:       member.MemberRole,
		}
		memberInfos = append(memberInfos, memberInfo)
	}

	return &CreateTeamResponse{
		ID:                 team.ID,
		TeamName:           team.TeamName,
		About:              team.About,
		TeamProfilePicture: team.TeamProfilePicture,
		CreatedAt:          team.CreatedAt,
		Members:            memberInfos,
		MemberCount:        len(memberInfos),
	}, nil
}

// Get my teams
func (s *StudentChallengeService) GetMyTeams(c *fiber.Ctx) ([]MyTeamResponse, error) {
	// Verify student access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "student" {
		return nil, errors.New("access denied: student role required")
	}

	// Get student profile ID
	studentProfileID, err := s.repo.GetStudentProfileByUserID(claims.UserID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	teams, err := s.repo.GetTeamsByStudentProfileID(*studentProfileID)
	if err != nil {
		return nil, errors.New("failed to fetch teams")
	}

	var response []MyTeamResponse
	for _, team := range teams {
		var memberInfos []TeamMemberInfo
		for _, member := range team.Members {
			memberInfo := TeamMemberInfo{
				StudentProfileID: member.StudentProfileID,
				MemberRole:       member.MemberRole,
			}
			memberInfos = append(memberInfos, memberInfo)
		}

		// Get registered challenges for this team
		var registeredChallenges []RegisteredChallengeInfo
		// You can implement this by getting submissions for the team

		teamResponse := MyTeamResponse{
			ID:                   team.ID,
			TeamName:             team.TeamName,
			About:                team.About,
			TeamProfilePicture:   team.TeamProfilePicture,
			CreatedAt:            team.CreatedAt,
			Members:              memberInfos,
			MemberCount:          len(memberInfos),
			RegisteredChallenges: registeredChallenges,
		}
		response = append(response, teamResponse)
	}

	return response, nil
}

// Get available challenges
func (s *StudentChallengeService) GetAvailableChallenges(c *fiber.Ctx, page, limit int, search string) (*utils.PaginationResponse, error) {
	// Verify student access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "student" {
		return nil, errors.New("akses ditolak: diperlukan role siswa")
	}

	// Validate search parameters
	validation, err := utils.ValidateSearchParams(search, page, limit)
	if err != nil {
		return nil, err
	}

	page = validation.Page
	limit = validation.Limit
	search = validation.Query

	// Get student profile ID
	studentProfileID, err := s.repo.GetStudentProfileByUserID(claims.UserID)
	if err != nil {
		return nil, errors.New("profil siswa tidak ditemukan")
	}

	offset := (page - 1) * limit

	// Get cached count first
	countKey := utils.GenerateCountCacheKey("student_challenges", search)
	var total int64
	if cachedCount, err := utils.GetCachedCount(s.redis, countKey); err == nil {
		total = cachedCount
	} else {
		total, err = s.repo.CountChallenges(search)
		if err != nil {
			return nil, errors.New("gagal mengambil jumlah data tantangan")
		}
		utils.CacheCount(s.redis, countKey, total)
	}

	// Get challenges data
	challenges, err := s.repo.GetChallengesOptimized(offset, limit, search)
	if err != nil {
		return nil, errors.New("gagal mengambil data tantangan")
	}

	var result []ChallengeListResponse
	for _, ch := range challenges {
		// Check if student is already registered for this challenge
		isRegistered, _ := s.repo.IsStudentInChallengeTeam(*studentProfileID, ch.ID)

		var myTeamID *uuid.UUID
		if isRegistered {
			// Find the team ID for this student in this challenge
			for _, submission := range ch.Submissions {
				if submission.TeamID != nil {
					for _, member := range submission.Team.Members {
						if member.StudentProfileID == *studentProfileID {
							myTeamID = submission.TeamID
							break
						}
					}
				}
			}
		}

		result = append(result, ChallengeListResponse{
			ID:                  ch.ID,
			ThumbnailImage:      ch.ThumbnailImage,
			Title:               ch.Title,
			Description:         ch.Description,
			Deadline:            ch.Deadline,
			Prize:               ch.Prize,
			MaxParticipants:     ch.MaxParticipants,
			CurrentParticipants: ch.CurrentParticipants,
			CanRegister:         ch.CanRegister() && !isRegistered,
			IsActive:            ch.IsActive(),
			OrganizerName:       ch.GetOrganizerName(),
			IsRegistered:        isRegistered,
			MyTeamID:            myTeamID,
		})
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &utils.PaginationResponse{
		Data:       result,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
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
	studentProfileID, err := s.repo.GetStudentProfileByUserID(claims.UserID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// Verify team exists and student is a member
	team, err := s.repo.GetTeamByIDAndMember(req.TeamID, *studentProfileID)
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
	_, err = s.repo.GetSubmissionByTeamAndChallenge(req.TeamID, req.ChallengeID)
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

	// Create submission in repository
	if err := s.repo.CreateSubmission(submission); err != nil {
		return nil, errors.New("failed to register team")
	}

	// Update challenge participant count
	if err := s.repo.UpdateChallengeParticipants(req.ChallengeID, true); err != nil {
		return nil, errors.New("failed to update participant count")
	}

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
