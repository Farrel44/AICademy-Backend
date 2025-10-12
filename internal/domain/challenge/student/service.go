package student_challenge

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/challenge"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StudentChallengeService struct {
	repo     *challenge.ChallengeRepository
	redis    *utils.RedisClient
	s3Client *s3.Client
	bucket   string
	baseURL  string
}

func NewStudentChallengeService(repo *challenge.ChallengeRepository, redis *utils.RedisClient) *StudentChallengeService {
	bucketName := os.Getenv("R2_BUCKET_NAME")
	accountId := os.Getenv("R2_ACCOUNT_ID")
	accessKeyId := os.Getenv("R2_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")
	baseURL := os.Getenv("OBJECT_STORAGE_URL")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to load AWS config: %v", err))
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	return &StudentChallengeService{
		repo:     repo,
		redis:    redis,
		s3Client: client,
		bucket:   bucketName,
		baseURL:  baseURL,
	}
}

func (s *StudentChallengeService) GetChallengeByID(c *fiber.Ctx, challengeID uuid.UUID) (*challenge.Challenge, error) {
	// Verify student access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "student" {
		return nil, errors.New("access denied: student role required")
	}

	// Use the new method with auto winner check
	challengeData, err := s.repo.GetChallengeByIDWithWinnerCheck(challengeID)
	if err != nil {
		return nil, errors.New("challenge not found or access denied")
	}

	return challengeData, nil
}

func (s *StudentChallengeService) uploadFile(file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	// Upload to R2
	_, err = s.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", s.baseURL, filename), nil
}

// SearchStudents allows students to search for other students (limited information)
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

	// Get member details with full info for response
	memberInfos, err := s.repo.GetTeamMembersWithDetails(team.ID)
	if err != nil {
		return nil, errors.New("failed to get member details")
	}

	return &CreateTeamResponse{
		ID:          team.ID,
		TeamName:    team.TeamName,
		About:       team.About,
		CreatedAt:   team.CreatedAt,
		Members:     convertToDTOMembers(memberInfos),
		MemberCount: len(memberInfos),
	}, nil
}

func convertToDTOMembers(members []challenge.TeamMemberInfo) []TeamMemberInfo {
	dtoMembers := make([]TeamMemberInfo, len(members))
	for i, m := range members {
		dtoMembers[i] = TeamMemberInfo{
			StudentProfileID: m.StudentProfileID,
			MemberRole:       m.MemberRole,
			FullName:         m.FullName,
			NIS:              m.NIS,
		}
	}
	return dtoMembers
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
		// Get member details with full info
		memberInfos, err := s.repo.GetTeamMembersWithDetails(team.ID)
		if err != nil {
			return nil, errors.New("failed to get member details")
		}

		// Get registered challenges for this team
		var registeredChallenges []RegisteredChallengeInfo
		// You can implement this by getting submissions for the team

		teamResponse := MyTeamResponse{
			ID:                   team.ID,
			TeamName:             team.TeamName,
			About:                team.About,
			CreatedAt:            team.CreatedAt,
			Members:              convertToDTOMembers(memberInfos), // <-- use conversion
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

	// Check and auto announce winners before getting challenges
	s.repo.CheckAndAutoAnnounceWinners()

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

		// Update result building in GetAvailableChallenges method
		var winnerTeamName *string
		if ch.WinnerTeam != nil {
			winnerTeamName = &ch.WinnerTeam.TeamName
		}

		result = append(result, ChallengeListResponse{
			ID:                  ch.ID,
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
			WinnerTeamID:        ch.WinnerTeamID,
			WinnerTeamName:      winnerTeamName,
			IsWinnerAnnounced:   ch.WinnerTeamID != nil,
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

// Auto register to challenge without team_id
func (s *StudentChallengeService) AutoRegisterToChallenge(c *fiber.Ctx, req *AutoRegisterChallengeRequest) (*AutoRegisterChallengeResponse, error) {
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

	// Check if student already registered for this challenge
	isRegistered, _ := s.repo.IsStudentInChallengeTeam(*studentProfileID, req.ChallengeID)
	if isRegistered {
		return nil, errors.New("already registered for this challenge")
	}

	// Get student's active team (latest created team where student is member)
	team, err := s.repo.GetActiveTeamByStudentID(*studentProfileID)
	if err != nil {
		return nil, errors.New("no active team found, please create a team first")
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

	// Check if team already registered
	_, err = s.repo.GetSubmissionByTeamAndChallenge(team.ID, req.ChallengeID)
	if err == nil {
		return nil, errors.New("team already registered for this challenge")
	}

	// Create submission entry (registration)
	submission := &challenge.Submission{
		ChallengeID: req.ChallengeID,
		TeamID:      &team.ID,
		Title:       fmt.Sprintf("%s - %s Registration", team.TeamName, challengeData.Title),
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

	return &AutoRegisterChallengeResponse{
		Message:             "Team successfully registered for challenge",
		ChallengeID:         req.ChallengeID,
		TeamID:              team.ID,
		TeamName:            team.TeamName,
		RegistrationDate:    submission.SubmittedAt,
		CurrentParticipants: challengeData.CurrentParticipants + 1,
	}, nil
}

// Submit challenge
func (s *StudentChallengeService) SubmitChallenge(c *fiber.Ctx, req *SubmitChallengeRequest) (*SubmitChallengeResponse, error) {
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

	// Get student's active team
	team, err := s.repo.GetActiveTeamByStudentID(*studentProfileID)
	if err != nil {
		return nil, errors.New("no active team found")
	}

	// Check if team is registered for this challenge
	existingSubmission, err := s.repo.GetSubmissionByTeamAndChallenge(team.ID, req.ChallengeID)
	if err != nil {
		return nil, errors.New("team not registered for this challenge")
	}

	// Verify challenge is still active
	challengeData, err := s.repo.GetChallengeByID(req.ChallengeID)
	if err != nil {
		return nil, errors.New("challenge not found")
	}

	if challengeData.IsExpired() {
		return nil, errors.New("challenge deadline has passed")
	}

	// Upload files if provided
	var docsURL, imageURL *string

	if req.DocsFile != nil {
		url, err := s.uploadFile(req.DocsFile, "challenge-docs")
		if err != nil {
			return nil, fmt.Errorf("failed to upload docs file: %v", err)
		}
		docsURL = &url
	}

	if req.ImageFile != nil {
		url, err := s.uploadFile(req.ImageFile, "challenge-images")
		if err != nil {
			return nil, fmt.Errorf("failed to upload image file: %v", err)
		}
		imageURL = &url
	}

	// Update submission with actual submission data
	existingSubmission.Title = req.Title
	existingSubmission.RepoURL = req.RepoURL
	existingSubmission.DocsURL = docsURL
	existingSubmission.ImageURL = imageURL
	existingSubmission.SubmittedAt = time.Now()

	// Update submission in repository
	if err := s.repo.UpdateSubmission(existingSubmission); err != nil {
		return nil, errors.New("failed to update submission")
	}

	return &SubmitChallengeResponse{
		ID:          existingSubmission.ID,
		ChallengeID: req.ChallengeID,
		TeamID:      team.ID,
		Title:       req.Title,
		RepoURL:     req.RepoURL,
		DocsURL:     docsURL,
		ImageURL:    imageURL,
		SubmittedAt: existingSubmission.SubmittedAt,
		Message:     "Challenge submitted successfully",
	}, nil
}

// Get my submissions
func (s *StudentChallengeService) GetMySubmissions(c *fiber.Ctx, page, limit int) (*utils.PaginationResponse, error) {
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

	offset := (page - 1) * limit

	// Get submissions by student
	total, err := s.repo.CountSubmissionsByStudentID(*studentProfileID)
	if err != nil {
		return nil, errors.New("failed to count submissions")
	}

	submissions, err := s.repo.GetSubmissionsByStudentID(*studentProfileID, offset, limit)
	if err != nil {
		return nil, errors.New("failed to get submissions")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &utils.PaginationResponse{
		Data:       submissions,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
