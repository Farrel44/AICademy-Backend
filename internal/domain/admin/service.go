package admin

import (
	"errors"
	"fmt"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/auth"
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/google/uuid"
)

type AdminUserService struct {
	repo  *auth.AuthRepository
	redis *utils.RedisClient
}

func NewAdminUserService(repo *auth.AuthRepository, redis *utils.RedisClient) *AdminUserService {
	return &AdminUserService{repo: repo, redis: redis}
}

// Rate Limiting
func (s *AdminUserService) CheckRateLimit(userID string, limit int, window time.Duration) (allowed bool, remaining int, resetTime time.Time, err error) {
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

// ========== STUDENT METHODS WITH CACHING ==========

func (s *AdminUserService) GetStudents(page, limit int, search string) (*PaginatedStudentsResponse, error) {
	// Cache key dengan parameters
	cacheKey := fmt.Sprintf("students:page:%d:limit:%d:search:%s", page, limit, search)

	// Try cache first
	var cachedResult PaginatedStudentsResponse
	if err := s.redis.GetJSON(cacheKey, &cachedResult); err == nil {
		return &cachedResult, nil
	}

	// Cache miss - query database
	offset := (page - 1) * limit
	students, total, err := s.repo.GetStudents(offset, limit, search)
	if err != nil {
		return nil, errors.New("failed to get students")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	studentResponses := make([]StudentResponse, len(students))
	for i, student := range students {
		studentResponses[i] = StudentResponse{
			ID:             student.ID,
			UserID:         student.UserID,
			Email:          student.User.Email,
			Fullname:       student.Fullname,
			NIS:            student.NIS,
			Class:          student.Class,
			ProfilePicture: &student.ProfilePicture,
			Headline:       &student.Headline,
			Bio:            &student.Bio,
			CVFile:         student.CVFile,
			CreatedAt:      student.CreatedAt,
			UpdatedAt:      student.UpdatedAt,
		}
	}

	result := &PaginatedStudentsResponse{
		Data:       studentResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	s.redis.SetJSON(cacheKey, result, 5*time.Minute)

	return result, nil
}

func (s *AdminUserService) GetStudentByID(id uuid.UUID) (*StudentResponse, error) {
	// Cache key for individual student
	cacheKey := fmt.Sprintf("student:%s", id.String())

	// Try cache first
	var cachedStudent StudentResponse
	if err := s.redis.GetJSON(cacheKey, &cachedStudent); err == nil {
		return &cachedStudent, nil
	}

	// Cache miss - query database
	student, err := s.repo.GetStudentByID(id)
	if err != nil {
		return nil, errors.New("student not found")
	}

	result := &StudentResponse{
		ID:             student.ID,
		UserID:         student.UserID,
		Email:          student.User.Email,
		Fullname:       student.Fullname,
		NIS:            student.NIS,
		Class:          student.Class,
		ProfilePicture: &student.ProfilePicture,
		Headline:       &student.Headline,
		Bio:            &student.Bio,
		CVFile:         student.CVFile,
		CreatedAt:      student.CreatedAt,
		UpdatedAt:      student.UpdatedAt,
	}

	// Cache for 10 minutes
	s.redis.SetJSON(cacheKey, result, 10*time.Minute)

	return result, nil
}

func (s *AdminUserService) GetStatistics() (*StudentStatisticsResponse, error) {
	cacheKey := "student_statistics"

	// Try cache first
	var cachedStats StudentStatisticsResponse
	if err := s.redis.GetJSON(cacheKey, &cachedStats); err == nil {
		return &cachedStats, nil
	}

	// Cache miss - query database
	stats, err := s.repo.GetStudentStatistics()
	if err != nil {
		return nil, errors.New("failed to get statistics")
	}

	result := &StudentStatisticsResponse{
		TotalStudents: int(stats.TotalStudents),
		TotalRPL:      int(stats.TotalRPL),
		TotalTKJ:      int(stats.TotalTKJ),
	}

	// Cache for 15 minutes
	s.redis.SetJSON(cacheKey, result, 15*time.Minute)

	return result, nil
}

func (s *AdminUserService) UpdateStudent(id uuid.UUID, req *UpdateStudentRequest) (*StudentResponse, error) {
	student, err := s.repo.GetStudentByID(id)
	if err != nil {
		return nil, errors.New("student not found")
	}

	// Update fields if provided
	if req.Fullname != nil {
		student.Fullname = *req.Fullname
	}
	if req.NIS != nil {
		student.NIS = *req.NIS
	}
	if req.Class != nil {
		student.Class = *req.Class
	}
	if req.Headline != nil {
		student.Headline = *req.Headline
	}
	if req.Bio != nil {
		student.Bio = *req.Bio
	}

	if err := s.repo.UpdateStudentProfile(student); err != nil {
		return nil, errors.New("failed to update student")
	}

	// Invalidate cache after update
	s.invalidateStudentCache(id)

	return s.GetStudentByID(id)
}

func (s *AdminUserService) DeleteStudent(id uuid.UUID) error {
	student, err := s.repo.GetStudentByID(id)
	if err != nil {
		return errors.New("student not found")
	}

	if err := s.repo.DeleteStudent(student.UserID, student.ID); err != nil {
		return errors.New("failed to delete student")
	}

	// Invalidate cache after delete
	s.invalidateStudentCache(id)

	return nil
}

// ========== TEACHER METHODS WITH CACHING ==========

func (s *AdminUserService) GetTeachers(page, limit int, search string) (*PaginatedTeachersResponse, error) {
	// Cache key dengan parameters
	cacheKey := fmt.Sprintf("teachers:page:%d:limit:%d:search:%s", page, limit, search)

	// Try cache first
	var cachedResult PaginatedTeachersResponse
	if err := s.redis.GetJSON(cacheKey, &cachedResult); err == nil {
		return &cachedResult, nil
	}

	// Cache miss - query database
	offset := (page - 1) * limit
	teachers, total, err := s.repo.GetTeachers(offset, limit, search)
	if err != nil {
		return nil, errors.New("failed to get teachers")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	teacherResponses := make([]TeacherResponse, len(teachers))
	for i, teacher := range teachers {
		teacherResponses[i] = TeacherResponse{
			ID:             teacher.ID,
			UserID:         teacher.UserID,
			Email:          teacher.User.Email,
			Fullname:       teacher.Fullname,
			ProfilePicture: &teacher.ProfilePicture,
			CreatedAt:      teacher.CreatedAt,
		}
	}

	result := &PaginatedTeachersResponse{
		Data:       teacherResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	// Cache for 5 minutes
	s.redis.SetJSON(cacheKey, result, 5*time.Minute)

	return result, nil
}

func (s *AdminUserService) GetTeacherByID(id uuid.UUID) (*TeacherResponse, error) {
	// Cache key for individual teacher
	cacheKey := fmt.Sprintf("teacher:%s", id.String())

	// Try cache first
	var cachedTeacher TeacherResponse
	if err := s.redis.GetJSON(cacheKey, &cachedTeacher); err == nil {
		return &cachedTeacher, nil
	}

	// Cache miss - query database
	teacher, err := s.repo.GetTeacherByID(id)
	if err != nil {
		return nil, errors.New("teacher not found")
	}

	result := &TeacherResponse{
		ID:             teacher.ID,
		UserID:         teacher.UserID,
		Email:          teacher.User.Email,
		Fullname:       teacher.Fullname,
		ProfilePicture: &teacher.ProfilePicture,
		CreatedAt:      teacher.CreatedAt,
	}

	// Cache for 10 minutes
	s.redis.SetJSON(cacheKey, result, 10*time.Minute)

	return result, nil
}

func (s *AdminUserService) CreateTeacher(req *CreateTeacherRequest) (*TeacherResponse, error) {
	// Check if email already exists
	if exists, _ := s.repo.CheckEmailExists(req.Email); exists {
		return nil, errors.New("email already exists")
	}

	// Generate password hash
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	newUser := &user.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         user.RoleTeacher,
	}

	if err := s.repo.CreateUser(newUser); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Create teacher profile
	teacherProfile := &user.TeacherProfile{
		ID:             uuid.New(),
		UserID:         newUser.ID,
		Fullname:       req.Fullname,
		ProfilePicture: "",
	}

	if err := s.repo.CreateTeacherProfile(teacherProfile); err != nil {
		return nil, errors.New("failed to create teacher profile")
	}

	// Invalidate teachers list cache after create
	s.invalidateTeachersListCache()

	return s.GetTeacherByID(teacherProfile.ID)
}

func (s *AdminUserService) UpdateTeacher(id uuid.UUID, req *UpdateTeacherRequest) (*TeacherResponse, error) {
	teacher, err := s.repo.GetTeacherByID(id)
	if err != nil {
		return nil, errors.New("teacher not found")
	}

	// Update fields if provided
	if req.Fullname != nil {
		teacher.Fullname = *req.Fullname
	}
	if req.ProfilePicture != nil {
		teacher.ProfilePicture = *req.ProfilePicture
	}

	if err := s.repo.UpdateTeacherProfile(teacher); err != nil {
		return nil, errors.New("failed to update teacher")
	}

	// Invalidate cache after update
	s.invalidateTeacherCache(id)

	return s.GetTeacherByID(id)
}

func (s *AdminUserService) DeleteTeacher(id uuid.UUID) error {
	teacher, err := s.repo.GetTeacherByID(id)
	if err != nil {
		return errors.New("teacher not found")
	}

	if err := s.repo.DeleteTeacher(teacher.UserID, teacher.ID); err != nil {
		return errors.New("failed to delete teacher")
	}

	// Invalidate cache after delete
	s.invalidateTeacherCache(id)

	return nil
}

// ========== COMPANY METHODS WITH CACHING ==========

func (s *AdminUserService) GetCompanies(page, limit int, search string) (*PaginatedCompaniesResponse, error) {
	// Cache key dengan parameters
	cacheKey := fmt.Sprintf("companies:page:%d:limit:%d:search:%s", page, limit, search)

	// Try cache first
	var cachedResult PaginatedCompaniesResponse
	if err := s.redis.GetJSON(cacheKey, &cachedResult); err == nil {
		return &cachedResult, nil
	}

	// Cache miss - query database
	offset := (page - 1) * limit
	companies, total, err := s.repo.GetCompanies(offset, limit, search)
	if err != nil {
		return nil, errors.New("failed to get companies")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	companyResponses := make([]CompanyResponse, len(companies))
	for i, company := range companies {
		companyResponses[i] = CompanyResponse{
			ID:              company.ID,
			UserID:          company.UserID,
			Email:           company.User.Email,
			CompanyName:     company.CompanyName,
			CompanyLogo:     company.CompanyLogo,
			CompanyLocation: company.CompanyLocation,
			Description:     company.Description,
			CreatedAt:       company.CreatedAt,
		}
	}

	result := &PaginatedCompaniesResponse{
		Data:       companyResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	// Cache for 5 minutes
	s.redis.SetJSON(cacheKey, result, 5*time.Minute)

	return result, nil
}

func (s *AdminUserService) GetCompanyByID(id uuid.UUID) (*CompanyResponse, error) {
	// Cache key for individual company
	cacheKey := fmt.Sprintf("company:%s", id.String())

	// Try cache first
	var cachedCompany CompanyResponse
	if err := s.redis.GetJSON(cacheKey, &cachedCompany); err == nil {
		return &cachedCompany, nil
	}

	// Cache miss - query database
	company, err := s.repo.GetCompanyByID(id)
	if err != nil {
		return nil, errors.New("company not found")
	}

	result := &CompanyResponse{
		ID:              company.ID,
		UserID:          company.UserID,
		Email:           company.User.Email,
		CompanyName:     company.CompanyName,
		CompanyLogo:     company.CompanyLogo,
		CompanyLocation: company.CompanyLocation,
		Description:     company.Description,
		CreatedAt:       company.CreatedAt,
	}

	// Cache for 10 minutes
	s.redis.SetJSON(cacheKey, result, 10*time.Minute)

	return result, nil
}

func (s *AdminUserService) CreateCompany(req *CreateCompanyRequest) (*CompanyResponse, error) {
	// Check if email already exists
	if exists, _ := s.repo.CheckEmailExists(req.Email); exists {
		return nil, errors.New("email already exists")
	}

	// Generate password hash
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	newUser := &user.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         user.RoleCompany,
	}

	if err := s.repo.CreateUser(newUser); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Create company profile
	companyProfile := &user.CompanyProfile{
		ID:              uuid.New(),
		UserID:          newUser.ID,
		CompanyName:     req.CompanyName,
		CompanyLogo:     req.CompanyLogo,
		CompanyLocation: req.CompanyLocation,
		Description:     req.Description,
	}

	if err := s.repo.CreateCompanyProfile(companyProfile); err != nil {
		return nil, errors.New("failed to create company profile")
	}

	// Invalidate companies list cache after create
	s.invalidateCompaniesListCache()

	return s.GetCompanyByID(companyProfile.ID)
}

func (s *AdminUserService) UpdateCompany(id uuid.UUID, req *UpdateCompanyRequest) (*CompanyResponse, error) {
	company, err := s.repo.GetCompanyByID(id)
	if err != nil {
		return nil, errors.New("company not found")
	}

	// Update fields if provided
	if req.CompanyName != nil {
		company.CompanyName = *req.CompanyName
	}
	if req.CompanyLogo != nil {
		company.CompanyLogo = req.CompanyLogo
	}
	if req.CompanyLocation != nil {
		company.CompanyLocation = req.CompanyLocation
	}
	if req.Description != nil {
		company.Description = req.Description
	}

	if err := s.repo.UpdateCompanyProfile(company); err != nil {
		return nil, errors.New("failed to update company")
	}

	// Invalidate cache after update
	s.invalidateCompanyCache(id)

	return s.GetCompanyByID(id)
}

func (s *AdminUserService) DeleteCompany(id uuid.UUID) error {
	company, err := s.repo.GetCompanyByID(id)
	if err != nil {
		return errors.New("company not found")
	}

	if err := s.repo.DeleteCompany(company.UserID, company.ID); err != nil {
		return errors.New("failed to delete company")
	}

	// Invalidate cache after delete
	s.invalidateCompanyCache(id)

	return nil
}

// ========== ALUMNI METHODS WITH CACHING ==========

func (s *AdminUserService) GetAlumni(page, limit int, search string) (*PaginatedAlumniResponse, error) {
	// Cache key dengan parameters
	cacheKey := fmt.Sprintf("alumni:page:%d:limit:%d:search:%s", page, limit, search)

	// Try cache first
	var cachedResult PaginatedAlumniResponse
	if err := s.redis.GetJSON(cacheKey, &cachedResult); err == nil {
		return &cachedResult, nil
	}

	// Cache miss - query database
	offset := (page - 1) * limit
	alumni, total, err := s.repo.GetAlumni(offset, limit, search)
	if err != nil {
		return nil, errors.New("failed to get alumni")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	alumniResponses := make([]AlumniResponse, len(alumni))
	for i, alum := range alumni {
		alumniResponses[i] = AlumniResponse{
			ID:             alum.ID,
			UserID:         alum.UserID,
			Email:          alum.User.Email,
			Fullname:       alum.Fullname,
			ProfilePicture: &alum.ProfilePicture,
			Headline:       &alum.Headline,
			Bio:            &alum.Bio,
			CVFile:         alum.CVFile,
			CreatedAt:      alum.CreatedAt,
			UpdatedAt:      alum.UpdatedAt,
		}
	}

	result := &PaginatedAlumniResponse{
		Data:       alumniResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	// Cache for 5 minutes
	s.redis.SetJSON(cacheKey, result, 5*time.Minute)

	return result, nil
}

func (s *AdminUserService) GetAlumniByID(id uuid.UUID) (*AlumniResponse, error) {
	// Cache key for individual alumni
	cacheKey := fmt.Sprintf("alumni:%s", id.String())

	// Try cache first
	var cachedAlumni AlumniResponse
	if err := s.redis.GetJSON(cacheKey, &cachedAlumni); err == nil {
		return &cachedAlumni, nil
	}

	// Cache miss - query database
	alumni, err := s.repo.GetAlumniByID(id)
	if err != nil {
		return nil, errors.New("alumni not found")
	}

	result := &AlumniResponse{
		ID:             alumni.ID,
		UserID:         alumni.UserID,
		Email:          alumni.User.Email,
		Fullname:       alumni.Fullname,
		ProfilePicture: &alumni.ProfilePicture,
		Headline:       &alumni.Headline,
		Bio:            &alumni.Bio,
		CVFile:         alumni.CVFile,
		CreatedAt:      alumni.CreatedAt,
		UpdatedAt:      alumni.UpdatedAt,
	}

	// Cache for 10 minutes
	s.redis.SetJSON(cacheKey, result, 10*time.Minute)

	return result, nil
}

func (s *AdminUserService) UpdateAlumni(id uuid.UUID, req *UpdateAlumniRequest) (*AlumniResponse, error) {
	alumni, err := s.repo.GetAlumniByID(id)
	if err != nil {
		return nil, errors.New("alumni not found")
	}

	// Update fields if provided
	if req.Fullname != nil {
		alumni.Fullname = *req.Fullname
	}
	if req.ProfilePicture != nil {
		alumni.ProfilePicture = *req.ProfilePicture
	}
	if req.Headline != nil {
		alumni.Headline = *req.Headline
	}
	if req.Bio != nil {
		alumni.Bio = *req.Bio
	}

	if err := s.repo.UpdateAlumniProfile(alumni); err != nil {
		return nil, errors.New("failed to update alumni")
	}

	// Invalidate cache after update
	s.invalidateAlumniCache(id)

	return s.GetAlumniByID(id)
}

func (s *AdminUserService) DeleteAlumni(id uuid.UUID) error {
	alumni, err := s.repo.GetAlumniByID(id)
	if err != nil {
		return errors.New("alumni not found")
	}

	if err := s.repo.DeleteAlumni(alumni.UserID, alumni.ID); err != nil {
		return errors.New("failed to delete alumni")
	}

	// Invalidate cache after delete
	s.invalidateAlumniCache(id)

	return nil
}

// ========== CACHE INVALIDATION HELPERS ==========

func (s *AdminUserService) invalidateStudentCache(studentID uuid.UUID) {
	// Hapus individual student cache
	studentKey := fmt.Sprintf("student:%s", studentID.String())
	s.redis.Delete(studentKey)

	// Hapus statistics cache
	s.redis.Delete("student_statistics")

	// Hapus students list cache patterns
	s.invalidateStudentsListCache()
}

func (s *AdminUserService) invalidateStudentsListCache() {
	// Hapus common students list cache patterns
	commonKeys := []string{
		"students:page:1:limit:10:search:",
		"students:page:2:limit:10:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *AdminUserService) invalidateTeacherCache(teacherID uuid.UUID) {
	// Hapus individual teacher cache
	teacherKey := fmt.Sprintf("teacher:%s", teacherID.String())
	s.redis.Delete(teacherKey)

	// Hapus teachers list cache
	s.invalidateTeachersListCache()
}

func (s *AdminUserService) invalidateTeachersListCache() {
	// Hapus common teachers list cache patterns
	commonKeys := []string{
		"teachers:page:1:limit:10:search:",
		"teachers:page:2:limit:10:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *AdminUserService) invalidateCompanyCache(companyID uuid.UUID) {
	// Hapus individual company cache
	companyKey := fmt.Sprintf("company:%s", companyID.String())
	s.redis.Delete(companyKey)

	// Hapus companies list cache
	s.invalidateCompaniesListCache()
}

func (s *AdminUserService) invalidateCompaniesListCache() {
	// Hapus common companies list cache patterns
	commonKeys := []string{
		"companies:page:1:limit:10:search:",
		"companies:page:2:limit:10:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *AdminUserService) invalidateAlumniCache(alumniID uuid.UUID) {
	// Hapus individual alumni cache
	alumniKey := fmt.Sprintf("alumni:%s", alumniID.String())
	s.redis.Delete(alumniKey)

	// Hapus alumni list cache
	s.invalidateAlumniListCache()
}

func (s *AdminUserService) invalidateAlumniListCache() {
	// Hapus common alumni list cache patterns
	commonKeys := []string{
		"alumni:page:1:limit:10:search:",
		"alumni:page:2:limit:10:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}
