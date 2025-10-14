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

type CVConfig struct {
	MaxSkills                int
	MaxProjects              int
	MaxCertifications        int
	ATSFormatScore           int
	StructureScorePerSection int
	KeywordThresholds        struct {
		High   int
		Medium int
		Low    int
	}
	ScoreValues struct {
		High   int
		Medium int
		Low    int
		Min    int
	}
	DefaultSkillLevel    string
	DefaultSkillCategory string
}

func defaultCVConfig() CVConfig {
	config := CVConfig{
		MaxSkills:                15,
		MaxProjects:              8,
		MaxCertifications:        15,
		ATSFormatScore:           85,
		StructureScorePerSection: 18,
		DefaultSkillLevel:        "Proficient",
		DefaultSkillCategory:     "Technical",
	}

	config.KeywordThresholds.High = 12
	config.KeywordThresholds.Medium = 8
	config.KeywordThresholds.Low = 4

	config.ScoreValues.High = 90
	config.ScoreValues.Medium = 75
	config.ScoreValues.Low = 60
	config.ScoreValues.Min = 35

	return config
}

type CVService struct {
	repo         *CVRepository
	aiService    ai.AIService
	redis        *utils.RedisClient
	cacheManager *utils.CacheManager
	config       CVConfig
}

func NewCVService(repo *CVRepository, aiService ai.AIService, redis *utils.RedisClient) *CVService {
	return &CVService{
		repo:         repo,
		aiService:    aiService,
		redis:        redis,
		cacheManager: utils.NewCacheManager(redis),
		config:       defaultCVConfig(),
	}
}

func (s *CVService) GenerateCV(userID uuid.UUID, title string) (*CV, error) {
	studentProfileID, err := s.repo.GetStudentProfileID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student profile ID: %w", err)
	}

	personalInfo, err := s.repo.GetStudentProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student profile: %w", err)
	}

	projects, err := s.repo.GetStudentProjects(userID, s.config.MaxProjects)
	if err != nil {
		return nil, fmt.Errorf("failed to get student projects: %w", err)
	}

	certifications, err := s.repo.GetStudentCertifications(userID, s.config.MaxCertifications)
	if err != nil {
		return nil, fmt.Errorf("failed to get student certifications: %w", err)
	}

	skills, err := s.repo.GetStudentSkills(userID, s.config.DefaultSkillCategory, s.config.DefaultSkillLevel, s.config.MaxSkills)
	if err != nil {
		return nil, fmt.Errorf("failed to get student skills: %w", err)
	}

	education, err := s.repo.GetStudentEducation(userID)
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
		StudentProfileID: studentProfileID,
		Title:            title,
		Status:           CVStatusDraft,
		Content:          content,
		GeneratedAt:      time.Now(),
	}

	if err := s.repo.CreateCV(cv); err != nil {
		return nil, fmt.Errorf("failed to create CV: %w", err)
	}

	s.cacheManager.InvalidateByPattern("cvs:*")
	s.cacheManager.InvalidateByPattern("user_cvs:*")
	userCacheKey := fmt.Sprintf("user_cvs:%s", userID.String())
	s.redis.Delete(userCacheKey)

	return cv, nil
}

func (s *CVService) PreviewCV(userID uuid.UUID) (*CVContent, error) {
	cacheKey := fmt.Sprintf("cv_preview:%s", userID.String())

	var content CVContent
	if err := s.redis.GetJSON(cacheKey, &content); err == nil {
		return &content, nil
	}

	personalInfo, _ := s.repo.GetStudentProfile(userID)
	projects, _ := s.repo.GetStudentProjects(userID, s.config.MaxProjects)
	certifications, _ := s.repo.GetStudentCertifications(userID, s.config.MaxCertifications)
	skills, _ := s.repo.GetStudentSkills(userID, s.config.DefaultSkillCategory, s.config.DefaultSkillLevel, s.config.MaxSkills)
	education, _ := s.repo.GetStudentEducation(userID)

	summary, err := s.generateAISummary(personalInfo, projects, skills)
	if err != nil {
		summary = s.generateDefaultSummary(personalInfo, skills)
	}

	content = CVContent{
		PersonalInfo:   *personalInfo,
		Summary:        summary,
		Skills:         skills,
		Projects:       projects,
		Certifications: certifications,
		Education:      *education,
		Keywords:       s.extractKeywords(projects, skills),
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, content, "short")

	return &content, nil
}

func (s *CVService) GetStudentCVs(userID uuid.UUID) ([]CV, error) {
	cacheKey := fmt.Sprintf("user_cvs:%s", userID.String())

	var cvs []CV
	if err := s.redis.GetJSON(cacheKey, &cvs); err == nil {
		return cvs, nil
	}

	cvs, err := s.repo.GetCVsByUserID(userID)
	if err != nil {
		return nil, err
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, cvs, "medium")
	return cvs, nil
}

func (s *CVService) GetCVByID(id uuid.UUID) (*CV, error) {
	cacheKey := fmt.Sprintf("cv:%s", id.String())

	var cv CV
	if err := s.redis.GetJSON(cacheKey, &cv); err == nil {
		return &cv, nil
	}

	cvPtr, err := s.repo.GetCVByID(id)
	if err != nil {
		return nil, err
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, *cvPtr, "medium")
	return cvPtr, nil
}

func (s *CVService) UpdateCV(id uuid.UUID, content *CVContent) error {
	cv, err := s.repo.GetCVByID(id)
	if err != nil {
		return err
	}

	cv.Content = *content
	cv.UpdatedAt = time.Now()

	err = s.repo.UpdateCV(cv)
	if err != nil {
		return err
	}

	itemKey := fmt.Sprintf("cv:%s", id.String())
	s.redis.Delete(itemKey)
	s.cacheManager.InvalidateByPattern("cvs:*")
	s.cacheManager.InvalidateByPattern("user_cvs:*")
	s.cacheManager.InvalidateByPattern("cv_preview:*")

	return nil
}

func (s *CVService) DeleteCV(id uuid.UUID) error {
	err := s.repo.DeleteCV(id)
	if err != nil {
		return err
	}

	itemKey := fmt.Sprintf("cv:%s", id.String())
	s.redis.Delete(itemKey)
	s.cacheManager.InvalidateByPattern("cvs:*")
	s.cacheManager.InvalidateByPattern("user_cvs:*")
	s.cacheManager.InvalidateByPattern("cv_preview:*")

	return nil
}

func (s *CVService) PublishCV(id uuid.UUID) error {
	err := s.repo.PublishCV(id)
	if err != nil {
		return err
	}

	itemKey := fmt.Sprintf("cv:%s", id.String())
	s.redis.Delete(itemKey)
	s.cacheManager.InvalidateByPattern("cvs:*")
	s.cacheManager.InvalidateByPattern("user_cvs:*")
	s.cacheManager.InvalidateByPattern("public_cvs:*")

	return nil
}

func (s *CVService) UnpublishCV(id uuid.UUID) error {
	err := s.repo.UnpublishCV(id)
	if err != nil {
		return err
	}

	itemKey := fmt.Sprintf("cv:%s", id.String())
	s.redis.Delete(itemKey)
	s.cacheManager.InvalidateByPattern("cvs:*")
	s.cacheManager.InvalidateByPattern("user_cvs:*")
	s.cacheManager.InvalidateByPattern("public_cvs:*")

	return nil
}

func (s *CVService) GetPublicCVs(userID uuid.UUID) ([]CV, error) {
	cacheKey := fmt.Sprintf("public_cvs:%s", userID.String())

	var cvs []CV
	if err := s.redis.GetJSON(cacheKey, &cvs); err == nil {
		return cvs, nil
	}

	cvs, err := s.repo.GetPublicCVsByUserID(userID)
	if err != nil {
		return nil, err
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, cvs, "medium")
	return cvs, nil
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

	score.Format = s.config.ATSFormatScore

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

	skillsStr := strings.Join(s.getSkillNames(skills), ", ")

	prompt := fmt.Sprintf(`
		Write ONLY a professional CV summary paragraph for %s in ENGLISH. No explanations, no options, just the final summary.
		
		Profile:
		- Name: %s
		- Technical Skills: %s
		- Projects Completed: %d
		- Education: Technology/Software Development Student
		
		Requirements:
		- Write exactly 2-3 sentences in ENGLISH
		- Include the main technical skills
		- Mention project experience
		- Use professional, ATS-friendly language
		- End with career objective (seeking internship/entry-level position)
		- Must be optimized for Applicant Tracking Systems (ATS)
		
		Return ONLY the final English summary paragraph, nothing else.
	`, info.FullName, info.FullName, skillsStr, len(projects))

	return s.aiService.GenerateText(context.Background(), prompt)
}

func (s *CVService) generateDefaultSummary(info *PersonalInfo, skills []CVSkill) string {
	skillNames := s.getSkillNames(skills)
	if len(skillNames) == 0 {
		return "Motivated technology student with hands-on experience in software development and a passion for creating innovative solutions."
	}

	maxSkills := s.config.MaxSkills / 3
	if maxSkills < 3 {
		maxSkills = 3
	}
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

	if keywordCount >= s.config.KeywordThresholds.High {
		return s.config.ScoreValues.High
	} else if keywordCount >= s.config.KeywordThresholds.Medium {
		return s.config.ScoreValues.Medium
	} else if keywordCount >= s.config.KeywordThresholds.Low {
		return s.config.ScoreValues.Low
	}
	return s.config.ScoreValues.Min
}

func (s *CVService) calculateStructureScore(content *CVContent) int {
	score := 0

	if content.PersonalInfo.FullName != "" {
		score += s.config.StructureScorePerSection
	}
	if content.Summary != "" {
		score += s.config.StructureScorePerSection
	}
	if len(content.Skills) > 0 {
		score += s.config.StructureScorePerSection
	}
	if len(content.Projects) > 0 {
		score += s.config.StructureScorePerSection
	}
	if content.Education.School != "" {
		score += s.config.StructureScorePerSection
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
			EndYear: func() int {
				if content.Education.EndYear != nil {
					return *content.Education.EndYear
				}
				return 0
			}(),
			GPA: content.Education.GPA,
		},
		Keywords: content.Keywords,
	}
}

func (s *CVService) invalidateCVCache(cvID uuid.UUID, userID uuid.UUID) {
	itemKey := fmt.Sprintf("cv:%s", cvID.String())
	s.redis.Delete(itemKey)

	s.redis.Delete("cv:statistics")
	s.cacheManager.InvalidateByPattern("cvs:*")
	s.cacheManager.InvalidateByPattern("student_cvs:*")

	userCacheKey := fmt.Sprintf("user_cvs:%s", userID.String())
	s.redis.Delete(userCacheKey)
}

func (s *CVService) invalidateUserCVCache(userID uuid.UUID) {
	userCacheKey := fmt.Sprintf("user_cvs:%s", userID.String())
	s.redis.Delete(userCacheKey)

	previewKey := fmt.Sprintf("cv_preview:%s", userID.String())
	s.redis.Delete(previewKey)

	s.cacheManager.InvalidateByPattern("cvs:*")
	s.cacheManager.InvalidateByPattern("student_cvs:*")
}
