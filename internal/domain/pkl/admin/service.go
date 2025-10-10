package pkl

import (
	"errors"
	"fmt"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/google/uuid"
)

type AdminPklService struct {
	repo  *pkl.PklRepository
	redis *utils.RedisClient
}

func NewAdminPklService(repo *pkl.PklRepository, redis *utils.RedisClient) *AdminPklService {
	return &AdminPklService{repo: repo, redis: redis}
}

func (s *AdminPklService) CheckRateLimit(userID string, limit int, window time.Duration) (allowed bool, remaining int, resetTime time.Time, err error) {
	key := fmt.Sprintf("rate_limit:%s", userID)
	count, err := s.redis.Incr(key)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	if count == 1 {
		s.redis.SetExpire(key, window)
	}

	ttl, _ := s.redis.GetTTL(key)
	resetTime = time.Now().Add(ttl)
	remaining = limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	allowed = count <= int64(limit)
	return allowed, remaining, resetTime, nil
}

func (s *AdminPklService) CreateInternshipPosition(req *CreateInternshipRequest) (*pkl.Internship, error) {
	existingCompany, _ := s.repo.GetCompanyByID(req.CompanyID)
	if existingCompany == nil {
		return nil, errors.New("Companies Not Found")
	}

	newInternshipPosition := pkl.Internship{
		CompanyProfileID: existingCompany.ID,
		Title:            req.Title,
		Description:      req.Description,
		Type:             req.Type,
		Deadline:         req.Deadline,
		PostedAt:         time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := s.repo.CreateInternshipPosition(&newInternshipPosition)
	if err != nil {
		return nil, errors.New("failed to create intership position")
	}

	s.invalidateAllPklCache()
	return &newInternshipPosition, nil
}

func (s *AdminPklService) invalidateInternshipCache(internshipID uuid.UUID) {
	internshipKey := fmt.Sprintf("internship:%s", internshipID.String())
	s.redis.Delete(internshipKey)

	s.redis.Delete("internship_statistics")

	s.invalidateInternshipsListCache()
}

func (s *AdminPklService) invalidateInternshipsListCache() {
	commonKeys := []string{
		"internships:page:1:limit:10:search:",
		"internships:page:2:limit:10:search:",
		"internships:page:1:limit:20:search:",
		"internships:page:2:limit:20:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *AdminPklService) invalidateApplicationCache(applicationID uuid.UUID) {
	applicationKey := fmt.Sprintf("application:%s", applicationID.String())
	s.redis.Delete(applicationKey)

	s.redis.Delete("internship_statistics")

	s.invalidateApplicationsListCache()
}

func (s *AdminPklService) invalidateApplicationsListCache() {
	commonKeys := []string{
		"applications:page:1:limit:10:search:",
		"applications:page:2:limit:10:search:",
		"applications:page:1:limit:20:search:",
		"applications:page:2:limit:20:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *AdminPklService) invalidateReviewCache(reviewID uuid.UUID) {
	reviewKey := fmt.Sprintf("review:%s", reviewID.String())
	s.redis.Delete(reviewKey)

	s.redis.Delete("internship_statistics")

	s.invalidateReviewsListCache()
}

func (s *AdminPklService) invalidateReviewsListCache() {
	commonKeys := []string{
		"reviews:page:1:limit:10:search:",
		"reviews:page:2:limit:10:search:",
		"reviews:page:1:limit:20:search:",
		"reviews:page:2:limit:20:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *AdminPklService) invalidateAllPklCache() {
	s.redis.Delete("internship_statistics")
	s.invalidateInternshipsListCache()
	s.invalidateApplicationsListCache()
	s.invalidateReviewsListCache()
}

func (s *AdminPklService) GetInternshipPositions(page, limit int, search string) (*CleanPaginatedInternshipResponse, error) {
	cacheKey := fmt.Sprintf("internship:page:%d:limit:%d:search:%s", page, limit, search)
	var cachedResult CleanPaginatedInternshipResponse

	if err := s.redis.GetJSON(cacheKey, &cachedResult); err == nil {
		fmt.Print("cache return")
		return &cachedResult, nil
	}

	internships, total, err := s.repo.GetAllInternships(page, limit, search)
	if err != nil {
		return nil, errors.New("failed to get internship positions")
	}

	totalPages := int(total+int64(limit)-1) / limit

	// Convert to clean response format
	var cleanInternships []CleanInternshipResponse
	for _, internship := range internships {
		var photosStr *string
		if len(internship.CompanyProfile.Photos) > 0 {
			if internship.CompanyProfile.Photos[0].PhotoURL != "" {
				photosStr = &internship.CompanyProfile.Photos[0].PhotoURL
			}
		}

		cleanInternship := CleanInternshipResponse{
			ID:               internship.ID.String(),
			CompanyProfileID: internship.CompanyProfileID.String(),
			Title:            internship.Title,
			Description:      internship.Description,
			Type:             string(internship.Type),
			PostedAt:         internship.PostedAt,
			Deadline:         internship.Deadline,
			CompanyProfile: CleanCompanyProfileResponse{
				ID:              internship.CompanyProfile.ID.String(),
				CompanyName:     internship.CompanyProfile.CompanyName,
				CompanyLogo:     internship.CompanyProfile.CompanyLogo,
				CompanyLocation: internship.CompanyProfile.CompanyLocation,
				Description:     internship.CompanyProfile.Description,
				Photos:          photosStr,
			},
		}
		cleanInternships = append(cleanInternships, cleanInternship)
	}

	result := &CleanPaginatedInternshipResponse{
		Data:       cleanInternships,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	s.redis.SetJSON(cacheKey, result, 5*time.Minute)

	return result, nil
}

func (s *AdminPklService) GetInternshipByID(id uuid.UUID) (*CleanInternshipResponse, error) {
	cacheKey := fmt.Sprintf("internship:%s", id.String())
	var cachedInternship CleanInternshipResponse
	if err := s.redis.GetJSON(cacheKey, &cachedInternship); err == nil {
		return &cachedInternship, nil
	}

	internship, err := s.repo.GetInternshipByID(id)
	if err != nil {
		return nil, errors.New("internship position not found")
	}

	var photosStr *string
	if len(internship.CompanyProfile.Photos) > 0 {
		if internship.CompanyProfile.Photos[0].PhotoURL != "" {
			photosStr = &internship.CompanyProfile.Photos[0].PhotoURL
		}
	}

	cleanInternship := &CleanInternshipResponse{
		ID:               internship.ID.String(),
		CompanyProfileID: internship.CompanyProfileID.String(),
		Title:            internship.Title,
		Description:      internship.Description,
		Type:             string(internship.Type),
		PostedAt:         internship.PostedAt,
		Deadline:         internship.Deadline,
		CompanyProfile: CleanCompanyProfileResponse{
			ID:              internship.CompanyProfile.ID.String(),
			CompanyName:     internship.CompanyProfile.CompanyName,
			CompanyLogo:     internship.CompanyProfile.CompanyLogo,
			CompanyLocation: internship.CompanyProfile.CompanyLocation,
			Description:     internship.CompanyProfile.Description,
			Photos:          photosStr,
		},
	}

	s.redis.SetJSON(cacheKey, cleanInternship, 5*time.Minute)

	return cleanInternship, nil
}

func (s *AdminPklService) UpdateInternshipPosition(id uuid.UUID, req *UpdateInternshipRequest) error {
	existingInternship, err := s.repo.GetInternshipByID(id)
	if err != nil {
		return errors.New("internship position not found")
	}

	if req.CompanyID != nil {
		existingCompany, _ := s.repo.GetCompanyByID(*req.CompanyID)
		if existingCompany == nil {
			return errors.New("company not found")
		}
		existingInternship.CompanyProfileID = *req.CompanyID
	}

	if req.Title != nil {
		existingInternship.Title = *req.Title
	}

	if req.Description != nil {
		existingInternship.Description = *req.Description
	}

	if req.Type != nil {
		existingInternship.Type = *req.Type
	}

	if req.Deadline != nil {
		existingInternship.Deadline = req.Deadline
	}

	existingInternship.UpdatedAt = time.Now()

	err = s.repo.UpdateInternshipPosition(existingInternship)
	if err != nil {
		return errors.New("failed to update internship position")
	}

	s.invalidateInternshipCache(id)
	return nil
}

func (s *AdminPklService) DeleteInternshipPosition(id uuid.UUID) error {
	existingInternship, err := s.repo.GetInternshipByID(id)
	if err != nil {
		return errors.New("internship position not found")
	}

	err = s.repo.DeleteInternshipPosition(existingInternship.ID)
	if err != nil {
		return errors.New("failed to delete internship position")
	}

	s.invalidateInternshipCache(id)
	return nil
}

func (s *AdminPklService) GetSubmissionsByInternshipID(internshipID uuid.UUID) ([]SubmissionResponse, error) {
	cacheKey := fmt.Sprintf("submissions:internship:%s", internshipID.String())
	var cached []SubmissionResponse
	if err := s.redis.GetJSON(cacheKey, &cached); err == nil {
		return cached, nil
	}

	submissions, err := s.repo.GetSubmissionsByInternshipID(internshipID)
	if err != nil {
		return nil, errors.New("failed to get submissions")
	}

	var out []SubmissionResponse
	for _, sub := range submissions {
		resp := SubmissionResponse{
			ID:               sub.ID.String(),
			InternshipID:     sub.InternshipID.String(),
			StudentProfileID: sub.StudentProfileID.String(),
			Status:           string(sub.Status),
			AppliedAt:        sub.AppliedAt,
			ReviewedAt:       sub.ReviewedAt,
			Student: StudentSummary{
				ID:             sub.StudentProfile.ID.String(),
				Name:           sub.StudentProfile.User.Email,
				Email:          sub.StudentProfile.User.Email,
				Fullname:       sub.StudentProfile.Fullname,
				NIS:            sub.StudentProfile.NIS,
				Class:          sub.StudentProfile.Class,
				ProfilePicture: &sub.StudentProfile.ProfilePicture,
				Headline:       &sub.StudentProfile.Headline,
				Bio:            &sub.StudentProfile.Bio,
				CvFile:         sub.StudentProfile.CVFile,
			},
		}
		if sub.ApprovedByUserID != nil {
			idStr := sub.ApprovedByUserID.String()
			resp.ApprovedByUserID = &idStr
			if sub.ApprovedByRole != nil {
				role := *sub.ApprovedByRole
				resp.ApprovedByRole = &role
			}
			if sub.ApprovedByUser != nil {
				email := sub.ApprovedByUser.Email
				resp.ApproverEmail = &email
			}
		}
		out = append(out, resp)
	}

	s.redis.SetJSON(cacheKey, out, 2*time.Minute)
	return out, nil
}

func (s *AdminPklService) GetInternshipsWithSubmissionsByCompanyID(companyID uuid.UUID) ([]InternshipWithSubmissionsResponse, error) {
	cacheKey := fmt.Sprintf("internships_submissions:company:%s", companyID.String())
	var cached []InternshipWithSubmissionsResponse
	if err := s.redis.GetJSON(cacheKey, &cached); err == nil {
		return cached, nil
	}

	internships, err := s.repo.GetInternshipsWithSubmissionsByCompanyID(companyID)
	if err != nil {
		return nil, errors.New("failed to get internships with submissions")
	}

	var responses []InternshipWithSubmissionsResponse
	for _, in := range internships {
		var subsResp []SubmissionResponse
		for _, app := range in.InternshipApplications {
			sr := SubmissionResponse{
				ID:               app.ID.String(),
				InternshipID:     app.InternshipID.String(),
				StudentProfileID: app.StudentProfileID.String(),
				Status:           string(app.Status),
				AppliedAt:        app.AppliedAt,
				ReviewedAt:       app.ReviewedAt,
				Student: StudentSummary{
					ID:             app.StudentProfile.ID.String(),
					Name:           app.StudentProfile.User.Email,
					Email:          app.StudentProfile.User.Email,
					Fullname:       app.StudentProfile.Fullname,
					NIS:            app.StudentProfile.NIS,
					Class:          app.StudentProfile.Class,
					ProfilePicture: &app.StudentProfile.ProfilePicture,
					Headline:       &app.StudentProfile.Headline,
					Bio:            &app.StudentProfile.Bio,
					CvFile:         app.StudentProfile.CVFile,
				},
			}
			if app.ApprovedByUserID != nil {
				idStr := app.ApprovedByUserID.String()
				sr.ApprovedByUserID = &idStr
				if app.ApprovedByRole != nil {
					role := *app.ApprovedByRole
					sr.ApprovedByRole = &role
				}
				if app.ApprovedByUser != nil {
					email := app.ApprovedByUser.Email
					sr.ApproverEmail = &email
				}
			}
			subsResp = append(subsResp, sr)
		}

		resp := InternshipWithSubmissionsResponse{
			ID:          in.ID.String(),
			Title:       in.Title,
			Description: in.Description,
			Type:        string(in.Type),
			PostedAt:    in.PostedAt,
			Deadline:    in.Deadline,
			Company: CompanySummary{
				Name:        in.CompanyProfile.CompanyName,
				Email:       in.CompanyProfile.User.Email,
				Logo:        in.CompanyProfile.CompanyLogo,
				Location:    in.CompanyProfile.CompanyLocation,
				Description: in.CompanyProfile.Description,
			},
			Submissions:     subsResp,
			SubmissionCount: len(subsResp),
		}
		responses = append(responses, resp)
	}

	s.redis.SetJSON(cacheKey, responses, 2*time.Minute)
	return responses, nil
}

func (s *AdminPklService) GetSubmissionByID(submissionID uuid.UUID) (*SubmissionResponse, error) {
	cacheKey := fmt.Sprintf("submission:%s", submissionID.String())
	var cached SubmissionResponse
	if err := s.redis.GetJSON(cacheKey, &cached); err == nil {
		return &cached, nil
	}

	sub, err := s.repo.GetSubmissionByID(submissionID)
	if err != nil {
		return nil, errors.New("submission not found")
	}

	resp := &SubmissionResponse{
		ID:               sub.ID.String(),
		InternshipID:     sub.InternshipID.String(),
		StudentProfileID: sub.StudentProfileID.String(),
		Status:           string(sub.Status),
		AppliedAt:        sub.AppliedAt,
		ReviewedAt:       sub.ReviewedAt,
		Student: StudentSummary{
			ID:             sub.StudentProfile.ID.String(),
			Name:           sub.StudentProfile.User.Email,
			Email:          sub.StudentProfile.User.Email,
			Fullname:       sub.StudentProfile.Fullname,
			NIS:            sub.StudentProfile.NIS,
			Class:          sub.StudentProfile.Class,
			ProfilePicture: &sub.StudentProfile.ProfilePicture,
			Headline:       &sub.StudentProfile.Headline,
			CvFile:         sub.StudentProfile.CVFile,
			Bio:            &sub.StudentProfile.Bio,
		},
		Internship: &InternshipSummary{
			ID:       sub.Internship.ID.String(),
			Title:    sub.Internship.Title,
			Type:     string(sub.Internship.Type),
			PostedAt: sub.Internship.PostedAt,
			Deadline: sub.Internship.Deadline,
			Company: CompanySummary{
				Name:        sub.Internship.CompanyProfile.CompanyName,
				Email:       sub.Internship.CompanyProfile.User.Email,
				Logo:        sub.Internship.CompanyProfile.CompanyLogo,
				Location:    sub.Internship.CompanyProfile.CompanyLocation,
				Description: sub.Internship.CompanyProfile.Description,
			},
		},
	}

	if sub.ApprovedByUserID != nil {
		idStr := sub.ApprovedByUserID.String()
		resp.ApprovedByUserID = &idStr
		if sub.ApprovedByRole != nil {
			role := *sub.ApprovedByRole
			resp.ApprovedByRole = &role
		}
		if sub.ApprovedByUser != nil {
			email := sub.ApprovedByUser.Email
			resp.ApproverEmail = &email
		}
	}

	s.redis.SetJSON(cacheKey, resp, 5*time.Minute)
	return resp, nil
}

func (s *AdminPklService) UpdateSubmissionStatus(submissionID uuid.UUID, status string, approverID uuid.UUID) error {
	var newStatus pkl.ApplicationStatus
	switch status {
	case "approved":
		newStatus = pkl.ApplicationStatusApproved
	case "rejected":
		newStatus = pkl.ApplicationStatusRejected
	default:
		return errors.New("invalid status")
	}

	_, err := s.repo.GetSubmissionByID(submissionID)
	if err != nil {
		return errors.New("submission not found")
	}

	// role hardcode "admin" (endpoint admin). Jika nanti teacher dipakai, bisa ambil dari context.
	if err := s.repo.UpdateSubmissionStatus(submissionID, newStatus, &approverID, "admin"); err != nil {
		return errors.New("failed to update submission status")
	}

	s.invalidateSubmissionCache(submissionID)
	return nil
}

func (s *AdminPklService) invalidateSubmissionCache(submissionID uuid.UUID) {
	key := fmt.Sprintf("submission:%s", submissionID.String())
	s.redis.Delete(key)
	s.redis.Delete("internship_statistics")
	s.invalidateApplicationsListCache()

	// hapus list cache generik (bisa disempurnakan dengan scan pattern jika ada util)
	s.redis.Delete("submissions:internship:")
	s.redis.Delete("internships_submissions:company:")
}
