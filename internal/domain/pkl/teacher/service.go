package pkl

import (
	"errors"
	"fmt"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
	pklAdmin "github.com/Farrel44/AICademy-Backend/internal/domain/pkl/admin"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/google/uuid"
)

type TeacherPklService struct {
	repo  *pkl.PklRepository
	redis *utils.RedisClient
}

func NewTeacherPklService(repo *pkl.PklRepository, redis *utils.RedisClient) *TeacherPklService {
	return &TeacherPklService{
		repo:  repo,
		redis: redis,
	}
}

func (s *TeacherPklService) CheckRateLimit(userID string, limit int, window time.Duration) (allowed bool, remaining int, resetTime time.Time, err error) {
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

func (s *TeacherPklService) invalidateInternshipCache(internshipID uuid.UUID) {
	internshipKey := fmt.Sprintf("internship:%s", internshipID.String())
	s.redis.Delete(internshipKey)
	s.redis.Delete("internship_statistics")
	s.invalidateInternshipsListCache()
}

func (s *TeacherPklService) invalidateInternshipsListCache() {
	commonKeys := []string{
		"teacher_internships:page:1:limit:10:search:",
		"teacher_internships:page:2:limit:10:search:",
		"teacher_internships:page:1:limit:20:search:",
		"teacher_internships:page:2:limit:20:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *TeacherPklService) invalidateApplicationCache(applicationID uuid.UUID) {
	applicationKey := fmt.Sprintf("teacher_application:%s", applicationID.String())
	s.redis.Delete(applicationKey)
	s.redis.Delete("internship_statistics")
	s.invalidateApplicationsListCache()
}

func (s *TeacherPklService) invalidateApplicationsListCache() {
	commonKeys := []string{
		"teacher_applications:page:1:limit:10:search:",
		"teacher_applications:page:2:limit:10:search:",
		"teacher_applications:page:1:limit:20:search:",
		"teacher_applications:page:2:limit:20:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *TeacherPklService) invalidateSubmissionCache(submissionID uuid.UUID) {
	key := fmt.Sprintf("teacher_submission:%s", submissionID.String())
	s.redis.Delete(key)
	s.redis.Delete("internship_statistics")
	s.invalidateApplicationsListCache()

	// Clear related caches
	s.redis.Delete("teacher_submissions:internship:")
	s.redis.Delete("teacher_internships_submissions:company:")
}

func (s *TeacherPklService) invalidateAllPklCache() {
	s.redis.Delete("internship_statistics")
	s.invalidateInternshipsListCache()
	s.invalidateApplicationsListCache()
}

func (s *TeacherPklService) GetInternshipPositions(page, limit int, search string) (*pklAdmin.CleanPaginatedInternshipResponse, error) {
	cacheKey := fmt.Sprintf("teacher_internships:page:%d:limit:%d:search:%s", page, limit, search)
	var cachedResult pklAdmin.CleanPaginatedInternshipResponse

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
	var cleanInternships []pklAdmin.CleanInternshipResponse
	for _, internship := range internships {
		cleanInternship := pklAdmin.CleanInternshipResponse{
			ID:               internship.ID.String(),
			CompanyProfileID: internship.CompanyProfileID.String(),
			Title:            internship.Title,
			Description:      internship.Description,
			Type:             string(internship.Type),
			PostedAt:         internship.PostedAt,
			Deadline:         internship.Deadline,
			CompanyProfile: pklAdmin.CleanCompanyProfileResponse{
				ID:              internship.CompanyProfile.ID.String(),
				CompanyName:     internship.CompanyProfile.CompanyName,
				CompanyLogo:     internship.CompanyProfile.CompanyLogo,
				CompanyLocation: internship.CompanyProfile.CompanyLocation,
				Description:     internship.CompanyProfile.Description,
				User: pklAdmin.CleanUserResponse{
					ID:    internship.CompanyProfile.User.ID.String(),
					Email: internship.CompanyProfile.User.Email,
					Role:  string(internship.CompanyProfile.User.Role),
				},
			},
		}
		cleanInternships = append(cleanInternships, cleanInternship)
	}

	result := &pklAdmin.CleanPaginatedInternshipResponse{
		Data:       cleanInternships,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	s.redis.SetJSON(cacheKey, result, 5*time.Minute)
	return result, nil
}

func (s *TeacherPklService) GetApplicationByID(applicationID uuid.UUID) (*SubmissionResponse, error) {
	cacheKey := fmt.Sprintf("teacher_submission:%s", applicationID.String())
	var cached SubmissionResponse
	if err := s.redis.GetJSON(cacheKey, &cached); err == nil {
		return &cached, nil
	}

	sub, err := s.repo.GetSubmissionByID(applicationID)
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
			Bio:            &sub.StudentProfile.Bio,
			CvFile:         sub.StudentProfile.CVFile,
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

func (s *TeacherPklService) GetInternshipsWithSubmissionsByCompanyID(companyID uuid.UUID) ([]pklAdmin.InternshipWithSubmissionsResponse, error) {
	cacheKey := fmt.Sprintf("internships_submissions:company:%s", companyID.String())
	var cached []pklAdmin.InternshipWithSubmissionsResponse
	if err := s.redis.GetJSON(cacheKey, &cached); err == nil {
		return cached, nil
	}

	internships, err := s.repo.GetInternshipsWithSubmissionsByCompanyID(companyID)
	if err != nil {
		return nil, errors.New("failed to get internships with submissions")
	}

	var responses []pklAdmin.InternshipWithSubmissionsResponse
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

		resp := pklAdmin.InternshipWithSubmissionsResponse{
			ID:          in.ID.String(),
			Title:       in.Title,
			Description: in.Description,
			Type:        string(in.Type),
			PostedAt:    in.PostedAt,
			Deadline:    in.Deadline,
			Company: pklAdmin.CompanySummary{
				Name:        in.CompanyProfile.CompanyName,
				Email:       in.CompanyProfile.User.Email,
				Logo:        in.CompanyProfile.CompanyLogo,
				Location:    in.CompanyProfile.CompanyLocation,
				Description: in.CompanyProfile.Description,
			},
		}
		responses = append(responses, resp)
	}

	s.redis.SetJSON(cacheKey, responses, 2*time.Minute)
	return responses, nil
}

func (s *TeacherPklService) UpdateApplicationStatus(applicationID uuid.UUID, status string, approverID uuid.UUID) error {
	var newStatus pkl.ApplicationStatus
	switch status {
	case "approved":
		newStatus = pkl.ApplicationStatusApproved
	case "rejected":
		newStatus = pkl.ApplicationStatusRejected
	default:
		return errors.New("invalid status")
	}

	_, err := s.repo.GetSubmissionByID(applicationID)
	if err != nil {
		return errors.New("submission not found")
	}

	if err := s.repo.UpdateSubmissionStatus(applicationID, newStatus, &approverID, "teacher"); err != nil {
		return errors.New("failed to update submission status")
	}

	s.invalidateSubmissionCache(applicationID)
	return nil
}

func (s *TeacherPklService) GetInternshipApplications(internshipID uuid.UUID) ([]SubmissionResponse, error) {
	cacheKey := fmt.Sprintf("teacher_submissions:internship:%s", internshipID.String())
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

func (s *TeacherPklService) GetSubmissionsByInternshipID(internshipID uuid.UUID) ([]SubmissionResponse, error) {
	cacheKey := fmt.Sprintf("teacher_submissions:internship:%s", internshipID.String())
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
			ID:           sub.ID.String(),
			InternshipID: sub.InternshipID.String(),
			Status:       string(sub.Status),
			AppliedAt:    sub.AppliedAt,
			ReviewedAt:   sub.ReviewedAt,
		}

		// Handle student applications
		if sub.StudentProfileID != nil {
			resp.StudentProfileID = sub.StudentProfileID.String()
			resp.Student = StudentSummary{
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
			}
		}

		// Handle alumni applications
		if sub.AlumniProfileID != nil {
			resp.AlumniProfileID = sub.AlumniProfileID.String()
			resp.Alumni = &AlumniSummary{
				ID:             sub.AlumniProfile.ID.String(),
				Name:           sub.AlumniProfile.User.Email,
				Email:          sub.AlumniProfile.User.Email,
				Fullname:       sub.AlumniProfile.Fullname,
				ProfilePicture: &sub.AlumniProfile.ProfilePicture,
				Headline:       &sub.AlumniProfile.Headline,
				Bio:            &sub.AlumniProfile.Bio,
				CvFile:         sub.AlumniProfile.CVFile,
			}
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

func (s *TeacherPklService) GetSubmissionByID(submissionID uuid.UUID) (*SubmissionResponse, error) {
	cacheKey := fmt.Sprintf("teacher_submission:%s", submissionID.String())
	var cached SubmissionResponse
	if err := s.redis.GetJSON(cacheKey, &cached); err == nil {
		return &cached, nil
	}

	sub, err := s.repo.GetSubmissionByID(submissionID)
	if err != nil {
		return nil, errors.New("submission not found")
	}

	resp := &SubmissionResponse{
		ID:           sub.ID.String(),
		InternshipID: sub.InternshipID.String(),
		Status:       string(sub.Status),
		AppliedAt:    sub.AppliedAt,
		ReviewedAt:   sub.ReviewedAt,
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

	// Handle student applications
	if sub.StudentProfileID != nil {
		resp.StudentProfileID = sub.StudentProfileID.String()
		resp.Student = StudentSummary{
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
		}
	}

	// Handle alumni applications
	if sub.AlumniProfileID != nil {
		resp.AlumniProfileID = sub.AlumniProfileID.String()
		resp.Alumni = &AlumniSummary{
			ID:             sub.AlumniProfile.ID.String(),
			Name:           sub.AlumniProfile.User.Email,
			Email:          sub.AlumniProfile.User.Email,
			Fullname:       sub.AlumniProfile.Fullname,
			ProfilePicture: &sub.AlumniProfile.ProfilePicture,
			Headline:       &sub.AlumniProfile.Headline,
			Bio:            &sub.AlumniProfile.Bio,
			CvFile:         sub.AlumniProfile.CVFile,
		}
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

func (s *TeacherPklService) UpdateSubmissionStatus(submissionID uuid.UUID, status string, approverID uuid.UUID) error {
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

	if err := s.repo.UpdateSubmissionStatus(submissionID, newStatus, &approverID, "teacher"); err != nil {
		return errors.New("failed to update submission status")
	}

	s.invalidateSubmissionCache(submissionID)
	return nil
}
