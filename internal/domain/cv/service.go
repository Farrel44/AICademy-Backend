package cv

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/services/ai"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/google/uuid"
)

type Service interface {
	GenerateCV(studentID uuid.UUID, title string) (*CV, error)
	PreviewCV(studentID uuid.UUID) (*CVContent, error)
	GetStudentCVs(studentID uuid.UUID) ([]CV, error)
	GetCVByID(id uuid.UUID) (*CV, error)
	UpdateCV(id uuid.UUID, content *CVContent) error
	DeleteCV(id uuid.UUID) error

	PublishCV(id uuid.UUID) error
	UnpublishCV(id uuid.UUID) error
	GetPublicCVs(studentID uuid.UUID) ([]CV, error)

	GeneratePDF(cvID uuid.UUID) (string, error)
	DownloadCV(cvID uuid.UUID) (string, error)

	AnalyzeATS(content *CVContent) (*ATSScore, error)
}

type CVService struct {
	repo      *CVRepository
	aiService ai.AIService
}

func NewCVService(repo *CVRepository, aiService ai.AIService) *CVService {
	return &CVService{
		repo:      repo,
		aiService: aiService,
	}
}

func (s *CVService) GenerateCV(studentID uuid.UUID, title string) (*CV, error) {
	personalInfo, err := s.repo.GetStudentProfile(studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student profile: %w", err)
	}

	projects, err := s.repo.GetStudentProjects(studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student projects: %w", err)
	}

	certifications, err := s.repo.GetStudentCertifications(studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student certifications: %w", err)
	}

	skills, err := s.repo.GetStudentSkills(studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student skills: %w", err)
	}

	education, err := s.repo.GetStudentEducation(studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student education: %w", err)
	}

	summary, err := s.generateAISummary(personalInfo, projects, skills)
	if err != nil {
		summary = s.generateDefaultSummary(personalInfo, skills)
	}

	content := CVContent{
		PersonalInfo:   *personalInfo,
		Summary:        summary,
		Skills:         skills,
		Projects:       projects,
		Certifications: certifications,
		Education:      *education,
		Keywords:       s.extractKeywords(projects, skills),
	}

	cv := &CV{
		StudentProfileID: studentID,
		Title:            title,
		Status:           CVStatusDraft,
		Content:          content,
		GeneratedAt:      time.Now(),
	}

	if err := s.repo.CreateCV(cv); err != nil {
		return nil, fmt.Errorf("failed to create CV: %w", err)
	}

	return cv, nil
}

func (s *CVService) PreviewCV(studentID uuid.UUID) (*CVContent, error) {
	personalInfo, _ := s.repo.GetStudentProfile(studentID)
	projects, _ := s.repo.GetStudentProjects(studentID)
	certifications, _ := s.repo.GetStudentCertifications(studentID)
	skills, _ := s.repo.GetStudentSkills(studentID)
	education, _ := s.repo.GetStudentEducation(studentID)

	summary, err := s.generateAISummary(personalInfo, projects, skills)
	if err != nil {
		summary = s.generateDefaultSummary(personalInfo, skills)
	}

	content := &CVContent{
		PersonalInfo:   *personalInfo,
		Summary:        summary,
		Skills:         skills,
		Projects:       projects,
		Certifications: certifications,
		Education:      *education,
		Keywords:       s.extractKeywords(projects, skills),
	}

	return content, nil
}

func (s *CVService) GetStudentCVs(studentID uuid.UUID) ([]CV, error) {
	return s.repo.GetCVsByStudentID(studentID)
}

func (s *CVService) GetCVByID(id uuid.UUID) (*CV, error) {
	return s.repo.GetCVByID(id)
}

func (s *CVService) UpdateCV(id uuid.UUID, content *CVContent) error {
	cv, err := s.repo.GetCVByID(id)
	if err != nil {
		return err
	}

	cv.Content = *content
	cv.UpdatedAt = time.Now()

	return s.repo.UpdateCV(cv)
}

func (s *CVService) DeleteCV(id uuid.UUID) error {
	return s.repo.DeleteCV(id)
}

func (s *CVService) PublishCV(id uuid.UUID) error {
	return s.repo.PublishCV(id)
}

func (s *CVService) UnpublishCV(id uuid.UUID) error {
	return s.repo.UnpublishCV(id)
}

func (s *CVService) GetPublicCVs(studentID uuid.UUID) ([]CV, error) {
	return s.repo.GetPublicCVsByStudentID(studentID)
}

func (s *CVService) GeneratePDF(cvID uuid.UUID) (string, error) {
	cv, err := s.repo.GetCVByID(cvID)
	if err != nil {
		return "", fmt.Errorf("CV not found: %w", err)
	}

	pdfGen := utils.NewPDFGenerator()
	utilsContent := s.convertToUtilsContent(&cv.Content)
	pdfPath, err := pdfGen.GenerateATSCV(&utilsContent)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	cv.PDFPath = pdfPath
	s.repo.UpdateCV(cv)

	return pdfPath, nil
}

func (s *CVService) DownloadCV(cvID uuid.UUID) (string, error) {
	cv, err := s.repo.GetCVByID(cvID)
	if err != nil {
		return "", err
	}

	if cv.PDFPath == "" {
		return s.GeneratePDF(cvID)
	}

	return cv.PDFPath, nil
}

func (s *CVService) AnalyzeATS(content *CVContent) (*ATSScore, error) {
	score := &ATSScore{}

	keywordScore := s.calculateKeywordScore(content)
	score.Keywords = keywordScore

	score.Format = 90

	structureScore := s.calculateStructureScore(content)
	score.Structure = structureScore

	score.Overall = (keywordScore + score.Format + structureScore) / 3

	score.Suggestions = s.generateATSSuggestions(content, score)

	return score, nil
}

func (s *CVService) generateAISummary(info *PersonalInfo, projects []CVProject, skills []CVSkill) (string, error) {
	if s.aiService == nil {
		return "", fmt.Errorf("AI service not available")
	}

	prompt := fmt.Sprintf(`
		Generate a professional CV summary for a student with the following profile:
		
		Name: %s
		Skills: %v
		Recent Projects: %d projects
		
		Create an ATS-friendly professional summary (2-3 sentences) that highlights:
		1. Key technical skills
		2. Project experience
		3. Career objectives
		
		Keep it concise and keyword-rich for ATS optimization.
	`, info.FullName, s.getSkillNames(skills), len(projects))

	return s.aiService.GenerateText(context.Background(), prompt)
}

func (s *CVService) generateDefaultSummary(info *PersonalInfo, skills []CVSkill) string {
	skillNames := s.getSkillNames(skills)
	if len(skillNames) == 0 {
		return "Motivated technology student with hands-on experience in software development and a passion for creating innovative solutions."
	}

	maxSkills := 5
	if len(skillNames) < maxSkills {
		maxSkills = len(skillNames)
	}

	return fmt.Sprintf("Dedicated technology student with expertise in %s. Experienced in developing practical solutions through hands-on projects and committed to continuous learning in the field of technology.",
		strings.Join(skillNames[:maxSkills], ", "))
}

func (s *CVService) extractKeywords(projects []CVProject, skills []CVSkill) []string {
	keywordMap := make(map[string]bool)

	for _, skill := range skills {
		keywordMap[skill.Name] = true
	}

	for _, project := range projects {
		for _, tech := range project.Technologies {
			keywordMap[tech] = true
		}
	}

	var keywords []string
	for keyword := range keywordMap {
		keywords = append(keywords, keyword)
	}

	return keywords
}

func (s *CVService) calculateKeywordScore(content *CVContent) int {
	keywordCount := len(content.Keywords)

	if keywordCount >= 15 {
		return 95
	} else if keywordCount >= 10 {
		return 80
	} else if keywordCount >= 5 {
		return 65
	}
	return 40
}

func (s *CVService) calculateStructureScore(content *CVContent) int {
	score := 0

	if content.PersonalInfo.FullName != "" {
		score += 20
	}
	if content.Summary != "" {
		score += 20
	}
	if len(content.Skills) > 0 {
		score += 20
	}
	if len(content.Projects) > 0 {
		score += 20
	}
	if content.Education.School != "" {
		score += 20
	}

	return score
}

func (s *CVService) generateATSSuggestions(content *CVContent, score *ATSScore) []string {
	var suggestions []string

	if score.Keywords < 70 {
		suggestions = append(suggestions, "Add more relevant technical keywords to improve ATS matching")
	}

	if len(content.Projects) < 3 {
		suggestions = append(suggestions, "Include more projects to demonstrate practical experience")
	}

	if content.Summary == "" {
		suggestions = append(suggestions, "Add a professional summary section")
	}

	if len(content.Certifications) == 0 {
		suggestions = append(suggestions, "Include relevant certifications if available")
	}

	return suggestions
}

func (s *CVService) getSkillNames(skills []CVSkill) []string {
	var names []string
	for _, skill := range skills {
		names = append(names, skill.Name)
	}
	return names
}

func (s *CVService) convertToUtilsContent(content *CVContent) utils.CVContent {
	utilsSkills := make([]utils.CVSkill, len(content.Skills))
	for i, skill := range content.Skills {
		utilsSkills[i] = utils.CVSkill{
			Name:     skill.Name,
			Category: skill.Category,
			Level:    skill.Level,
		}
	}

	utilsProjects := make([]utils.CVProject, len(content.Projects))
	for i, project := range content.Projects {
		utilsProjects[i] = utils.CVProject{
			ID:           project.ID.String(),
			Name:         project.Name,
			Role:         project.Role,
			Description:  project.Description,
			Technologies: project.Technologies,
			StartDate:    project.StartDate,
			EndDate:      project.EndDate,
			URL:          project.URL,
			Highlights:   project.Highlights,
		}
	}

	utilsCerts := make([]utils.CVCertification, len(content.Certifications))
	for i, cert := range content.Certifications {
		utilsCerts[i] = utils.CVCertification{
			ID:                  cert.ID.String(),
			Name:                cert.Name,
			IssuingOrganization: cert.IssuingOrganization,
			IssueDate:           cert.IssueDate,
			ExpirationDate:      cert.ExpirationDate,
			CredentialID:        cert.CredentialID,
			CredentialURL:       cert.CredentialURL,
		}
	}

	return utils.CVContent{
		PersonalInfo: utils.PersonalInfo{
			FullName:  content.PersonalInfo.FullName,
			Email:     content.PersonalInfo.Email,
			Phone:     content.PersonalInfo.Phone,
			Location:  content.PersonalInfo.Location,
			LinkedIn:  content.PersonalInfo.LinkedIn,
			GitHub:    content.PersonalInfo.GitHub,
			Portfolio: content.PersonalInfo.Portfolio,
		},
		Summary:        content.Summary,
		Skills:         utilsSkills,
		Projects:       utilsProjects,
		Certifications: utilsCerts,
		Education: utils.CVEducation{
			School:    content.Education.School,
			Degree:    content.Education.Degree,
			Major:     content.Education.Major,
			StartYear: content.Education.StartYear,
			EndYear:   *content.Education.EndYear,
			GPA:       content.Education.GPA,
		},
		Keywords: content.Keywords,
	}
}
