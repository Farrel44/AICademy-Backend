package admin

import (
	"errors"

	"github.com/Farrel44/AICademy-Backend/internal/domain/auth"
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/google/uuid"
)

type AdminUserService struct {
	repo *auth.AuthRepository
}

func NewAdminUserService(repo *auth.AuthRepository) *AdminUserService {
	return &AdminUserService{repo: repo}
}

func (s *AdminUserService) GetStudents(page, limit int, search string) (*PaginatedStudentsResponse, error) {
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

	return &PaginatedStudentsResponse{
		Data:       studentResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminUserService) GetStudentByID(id uuid.UUID) (*StudentResponse, error) {
	student, err := s.repo.GetStudentByID(id)
	if err != nil {
		return nil, errors.New("student not found")
	}

	return &StudentResponse{
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
	}, nil
}

func (s *AdminUserService) GetStatistics() (*StudentStatisticsResponse, error) {
	stats, err := s.repo.GetStudentStatistics()
	if err != nil {
		return nil, errors.New("failed to get statistics")
	}

	return &StudentStatisticsResponse{
		TotalStudents: int(stats.TotalStudents),
		TotalRPL:      int(stats.TotalRPL),
		TotalTKJ:      int(stats.TotalTKJ),
	}, nil
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

	return nil
}

// Teacher methods
func (s *AdminUserService) GetTeachers(page, limit int, search string) (*PaginatedTeachersResponse, error) {
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

	return &PaginatedTeachersResponse{
		Data:       teacherResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminUserService) GetTeacherByID(id uuid.UUID) (*TeacherResponse, error) {
	teacher, err := s.repo.GetTeacherByID(id)
	if err != nil {
		return nil, errors.New("teacher not found")
	}

	return &TeacherResponse{
		ID:             teacher.ID,
		UserID:         teacher.UserID,
		Email:          teacher.User.Email,
		Fullname:       teacher.Fullname,
		ProfilePicture: &teacher.ProfilePicture,
		CreatedAt:      teacher.CreatedAt,
	}, nil
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

	return nil
}

// Company methods
func (s *AdminUserService) GetCompanies(page, limit int, search string) (*PaginatedCompaniesResponse, error) {
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

	return &PaginatedCompaniesResponse{
		Data:       companyResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminUserService) GetCompanyByID(id uuid.UUID) (*CompanyResponse, error) {
	company, err := s.repo.GetCompanyByID(id)
	if err != nil {
		return nil, errors.New("company not found")
	}

	return &CompanyResponse{
		ID:              company.ID,
		UserID:          company.UserID,
		Email:           company.User.Email,
		CompanyName:     company.CompanyName,
		CompanyLogo:     company.CompanyLogo,
		CompanyLocation: company.CompanyLocation,
		Description:     company.Description,
		CreatedAt:       company.CreatedAt,
	}, nil
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

	return nil
}

// Alumni methods
func (s *AdminUserService) GetAlumni(page, limit int, search string) (*PaginatedAlumniResponse, error) {
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

	return &PaginatedAlumniResponse{
		Data:       alumniResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminUserService) GetAlumniByID(id uuid.UUID) (*AlumniResponse, error) {
	alumni, err := s.repo.GetAlumniByID(id)
	if err != nil {
		return nil, errors.New("alumni not found")
	}

	return &AlumniResponse{
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
	}, nil
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

	return nil
}
