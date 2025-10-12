package teacher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/roadmap"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TeacherService struct {
	repo        *roadmap.RoadmapRepository
	redisClient *redis.Client
}

func NewTeacherService(repo *roadmap.RoadmapRepository, redisClient *redis.Client) *TeacherService {
	return &TeacherService{
		repo:        repo,
		redisClient: redisClient,
	}
}

func (s *TeacherService) GetPendingSubmissions(teacherID uuid.UUID, page, limit int, search string) (*PaginatedSubmissionsResponse, error) {
	cacheKey := fmt.Sprintf("teacher_roadmap_submissions:%s:%d:%d:%s", teacherID.String(), page, limit, search)

	if cached, err := s.redisClient.Get(context.Background(), cacheKey).Result(); err == nil {
		var result PaginatedSubmissionsResponse
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	offset := (page - 1) * limit
	submissions, total, err := s.repo.GetPendingSubmissionsOptimized(&teacherID, offset, limit, search)
	if err != nil {
		return nil, errors.New("failed to get pending submissions")
	}

	var submissionResponses []PendingSubmissionResponse
	for _, submission := range submissions {
		var studentName, studentEmail, roadmapName, stepTitle string
		var stepOrder int
		var learningObjectives, submissionGuidelines string

		if submission.StudentRoadmapProgress.StudentProfile != nil {
			studentName = submission.StudentRoadmapProgress.StudentProfile.Fullname
			if submission.StudentRoadmapProgress.StudentProfile.User.Email != "" {
				studentEmail = submission.StudentRoadmapProgress.StudentProfile.User.Email
			}
		}

		if submission.StudentRoadmapProgress.Roadmap.RoadmapName != "" {
			roadmapName = submission.StudentRoadmapProgress.Roadmap.RoadmapName
		}

		if submission.RoadmapStep.Title != "" {
			stepTitle = submission.RoadmapStep.Title
			stepOrder = submission.RoadmapStep.StepOrder
			learningObjectives = submission.RoadmapStep.LearningObjectives
			submissionGuidelines = submission.RoadmapStep.SubmissionGuidelines
		}

		submissionResponses = append(submissionResponses, PendingSubmissionResponse{
			ID:                   submission.ID,
			StudentName:          studentName,
			StudentEmail:         studentEmail,
			RoadmapName:          roadmapName,
			StepTitle:            stepTitle,
			StepOrder:            stepOrder,
			EvidenceLink:         submission.EvidenceLink,
			EvidenceType:         submission.EvidenceType,
			SubmissionNotes:      submission.SubmissionNotes,
			SubmittedAt:          submission.SubmittedAt,
			LearningObjectives:   learningObjectives,
			SubmissionGuidelines: submissionGuidelines,
		})
	}

	totalPages := (int(total) + limit - 1) / limit

	result := &PaginatedSubmissionsResponse{
		Data:       submissionResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	if resultJSON, err := json.Marshal(result); err == nil {
		s.redisClient.Set(context.Background(), cacheKey, string(resultJSON), time.Minute*5)
	}

	return result, nil
}

func (s *TeacherService) ReviewSubmission(submissionID, teacherID uuid.UUID, req ReviewSubmissionRequest) (*ReviewSubmissionResponse, error) {
	if req.Action == "approve" {
		err := s.repo.ApproveStep(submissionID, teacherID, req.ValidationScore, req.ValidationNotes)
		if err != nil {
			return nil, errors.New("failed to approve submission")
		}

		return &ReviewSubmissionResponse{
			Message:    "Submission approved successfully",
			StepStatus: string(roadmap.RoadmapProgressStatusApproved),
			ReviewedAt: "now",
		}, nil
	} else if req.Action == "reject" {
		err := s.repo.RejectStep(submissionID, teacherID, req.ValidationNotes)
		if err != nil {
			return nil, errors.New("failed to reject submission")
		}

		return &ReviewSubmissionResponse{
			Message:    "Submission rejected. Student can revise and resubmit.",
			StepStatus: string(roadmap.RoadmapProgressStatusRejected),
			ReviewedAt: "now",
		}, nil
	}

	return nil, errors.New("invalid action")
}
