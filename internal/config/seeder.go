package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire"
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"gorm.io/gorm"
)

func SeedData(db *gorm.DB) error {
	log.Println("Memulai proses seeding database...")

	if err := SeedDefaultAdmin(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding admin: %v", err)
	}

	if err := SeedDefaultTeachers(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding guru: %v", err)
	}

	if err := SeedDefaultStudents(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding siswa: %v", err)
	}

	if err := SeedDefaultAlumni(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding alumni: %v", err)
	}

	if err := SeedDefaultCompanies(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding perusahaan: %v", err)
	}

	if err := SeedDefaultQuestionnaires(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding kuesioner: %v", err)
	}

	if err := SeedRoleRecommendations(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding role recommendations: %v", err)
	}

	log.Println("Proses seeding database selesai dengan sukses")
	return nil
}
func SeedDefaultAdmin(db *gorm.DB) error {
	var existingAdmin user.User
	err := db.Where("role = ? AND email = ?", user.RoleAdmin, "admin@aicademy.com").First(&existingAdmin).Error

	if err == nil {
		log.Println("Admin default sudah ada")
		return nil
	}

	adminPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "Admin123!"
	}

	hashedPassword, err := utils.HashPassword(adminPassword)
	if err != nil {
		return err
	}

	admin := user.User{
		Email:        "admin@aicademy.com",
		PasswordHash: hashedPassword,
		Role:         user.RoleAdmin,
	}

	err = db.Create(&admin).Error
	if err != nil {
		return err
	}

	log.Printf("Admin default berhasil dibuat: admin@aicademy.com (password: %s)", adminPassword)
	return nil
}
func SeedDefaultTeachers(db *gorm.DB) error {
	teachers := []struct {
		Email    string
		Password string
		Fullname string
	}{
		{"teacher.programming@aicademy.com", "Teacher123!", "Budi Santoso"},
		{"teacher.database@aicademy.com", "Teacher123!", "Siti Rahayu"},
		{"teacher.networking@aicademy.com", "Teacher123!", "Ahmad Wijaya"},
		{"teacher.mobile@aicademy.com", "Teacher123!", "Dewi Kartika"},
		{"teacher.web@aicademy.com", "Teacher123!", "Rudi Hermawan"},
	}

	for _, teacherData := range teachers {
		var existingTeacher user.User
		err := db.Where("email = ?", teacherData.Email).First(&existingTeacher).Error

		if err == nil {
			log.Printf("Guru dengan email %s sudah ada, melewati...", teacherData.Email)
			continue
		}

		hashedPassword, err := utils.HashPassword(teacherData.Password)
		if err != nil {
			return err
		}

		teacher := user.User{
			Email:        teacherData.Email,
			PasswordHash: hashedPassword,
			Role:         user.RoleTeacher,
		}

		err = db.Create(&teacher).Error
		if err != nil {
			return err
		}

		teacherProfile := user.TeacherProfile{
			UserID:   teacher.ID,
			Fullname: teacherData.Fullname,
		}

		err = db.Create(&teacherProfile).Error
		if err != nil {
			return err
		}

		log.Printf("Guru berhasil dibuat: %s (%s)", teacherData.Email, teacherData.Fullname)
	}

	return nil
}

func SeedDefaultStudents(db *gorm.DB) error {
	students := []struct {
		Email    string
		Fullname string
		NIS      string
		Class    string
	}{
		{"student1@aicademy.com", "Andi Pratama", "12345001", "XII-RPL-1"},
		{"student2@aicademy.com", "Bella Safitri", "12345002", "XII-RPL-1"},
		{"student3@aicademy.com", "Chandra Kirana", "12345003", "XII-RPL-1"},
		{"student4@aicademy.com", "Dimas Prasetyo", "12345004", "XII-RPL-2"},
		{"student5@aicademy.com", "Eka Putri", "12345005", "XII-RPL-2"},
		{"student6@aicademy.com", "Fajar Nugroho", "12345006", "XII-RPL-2"},
		{"student7@aicademy.com", "Gita Sari", "12345007", "XII-TKJ-1"},
		{"student8@aicademy.com", "Hendra Wijaya", "12345008", "XII-TKJ-1"},
		{"student9@aicademy.com", "Indira Kusuma", "12345009", "XII-TKJ-2"},
		{"student10@aicademy.com", "Joko Susilo", "12345010", "XII-TKJ-2"},
	}

	hashedPassword, err := utils.HashPassword("telkom@2025")
	if err != nil {
		return err
	}

	for _, studentData := range students {
		var existingStudent user.User
		err := db.Where("email = ?", studentData.Email).First(&existingStudent).Error

		if err == nil {
			log.Printf("Siswa dengan email %s sudah ada, melewati...", studentData.Email)
			continue
		}

		student := user.User{
			Email:        studentData.Email,
			PasswordHash: hashedPassword,
			Role:         user.RoleStudent,
		}

		err = db.Create(&student).Error
		if err != nil {
			return err
		}

		studentProfile := user.StudentProfile{
			UserID:   student.ID,
			Fullname: studentData.Fullname,
			NIS:      studentData.NIS,
			Class:    studentData.Class,
		}

		err = db.Create(&studentProfile).Error
		if err != nil {
			return err
		}

		log.Printf("Siswa berhasil dibuat: %s (%s, NIS: %s)", studentData.Email, studentData.Fullname, studentData.NIS)
	}

	return nil
}

func SeedDefaultAlumni(db *gorm.DB) error {
	alumni := []struct {
		Email          string
		Password       string
		Fullname       string
		GraduationYear int
		Major          string
	}{
		{"alumni1@aicademy.com", "Alumni123!", "Rizki Ramadhan", 2020, "Rekayasa Perangkat Lunak"},
		{"alumni2@aicademy.com", "Alumni123!", "Maya Sari", 2021, "Teknik Komputer Jaringan"},
		{"alumni3@aicademy.com", "Alumni123!", "Bayu Setiawan", 2022, "Rekayasa Perangkat Lunak"},
		{"alumni4@aicademy.com", "Alumni123!", "Citra Dewi", 2023, "Teknik Komputer Jaringan"},
		{"alumni5@aicademy.com", "Alumni123!", "David Kurniawan", 2024, "Rekayasa Perangkat Lunak"},
	}

	for _, alumniData := range alumni {
		var existingAlumni user.User
		err := db.Where("email = ?", alumniData.Email).First(&existingAlumni).Error

		if err == nil {
			log.Printf("Alumni dengan email %s sudah ada, melewati...", alumniData.Email)
			continue
		}

		hashedPassword, err := utils.HashPassword(alumniData.Password)
		if err != nil {
			return err
		}

		alumniUser := user.User{
			Email:        alumniData.Email,
			PasswordHash: hashedPassword,
			Role:         user.RoleAlumni,
		}

		err = db.Create(&alumniUser).Error
		if err != nil {
			return err
		}

		alumniProfile := user.AlumniProfile{
			UserID:   alumniUser.ID,
			Fullname: alumniData.Fullname,
		}

		err = db.Create(&alumniProfile).Error
		if err != nil {
			return err
		}

		log.Printf("Alumni berhasil dibuat: %s (%s, Tahun Lulus: %d)", alumniData.Email, alumniData.Fullname, alumniData.GraduationYear)
	}

	return nil
}

func SeedDefaultCompanies(db *gorm.DB) error {
	companies := []struct {
		Email           string
		Password        string
		CompanyName     string
		CompanyLocation string
		Description     string
	}{
		{
			"hr@techsolutions.com",
			"Company123!",
			"Tech Solutions Indonesia",
			"Jakarta Selatan",
			"Perusahaan pengembangan perangkat lunak terkemuka yang berfokus pada solusi enterprise",
		},
		{
			"recruitment@innovatech.com",
			"Company123!",
			"InnovaTech Labs",
			"Bandung",
			"Perusahaan teknologi inovatif yang berfokus pada solusi AI dan machine learning",
		},
		{
			"careers@digitalcorp.com",
			"Company123!",
			"Digital Corp",
			"Surabaya",
			"Konsultan transformasi digital yang membantu bisnis memodernisasi operasional",
		},
		{
			"jobs@smartsystems.com",
			"Company123!",
			"Smart Systems",
			"Yogyakarta",
			"Produsen perangkat IoT dan smart device dengan fokus pada solusi industri 4.0",
		},
		{
			"hiring@webstudio.com",
			"Company123!",
			"Creative Web Studio",
			"Denpasar",
			"Agensi kreatif yang berfokus pada pengembangan web dan pemasaran digital",
		},
	}

	for _, companyData := range companies {
		var existingCompany user.User
		err := db.Where("email = ?", companyData.Email).First(&existingCompany).Error

		if err == nil {
			log.Printf("Perusahaan dengan email %s sudah ada, melewati...", companyData.Email)
			continue
		}

		hashedPassword, err := utils.HashPassword(companyData.Password)
		if err != nil {
			return err
		}

		company := user.User{
			Email:        companyData.Email,
			PasswordHash: hashedPassword,
			Role:         user.RoleCompany,
		}

		err = db.Create(&company).Error
		if err != nil {
			return err
		}

		companyProfile := user.CompanyProfile{
			UserID:          company.ID,
			CompanyName:     companyData.CompanyName,
			CompanyLocation: &companyData.CompanyLocation,
			Description:     &companyData.Description,
		}

		err = db.Create(&companyProfile).Error
		if err != nil {
			return err
		}

		log.Printf("Perusahaan berhasil dibuat: %s (%s)", companyData.Email, companyData.CompanyName)
	}

	return nil
}
func SeedDefaultQuestionnaires(db *gorm.DB) error {
	questionnaires := []struct {
		Name        string
		GeneratedBy string
		Questions   []struct {
			QuestionText  string
			QuestionType  questionnaire.QuestionType
			Options       []questionnaire.QuestionOption
			MaxScore      int
			QuestionOrder int
			Category      string
		}
	}{
		{
			Name:        "Kuesioner Profiling Karir - Teknologi",
			GeneratedBy: "manual",
			Questions: []struct {
				QuestionText  string
				QuestionType  questionnaire.QuestionType
				Options       []questionnaire.QuestionOption
				MaxScore      int
				QuestionOrder int
				Category      string
			}{
				{
					QuestionText:  "Apa yang paling Anda nikmati dalam bekerja?",
					QuestionType:  questionnaire.QuestionTypeText,
					MaxScore:      0,
					QuestionOrder: 1,
					Category:      "preferences",
				},
				{
					QuestionText: "Seberapa nyaman Anda bekerja dalam tim?",
					QuestionType: questionnaire.QuestionTypeLikert,
					Options: []questionnaire.QuestionOption{
						{Label: "Sangat Tidak Nyaman", Value: "1"},
						{Label: "Tidak Nyaman", Value: "2"},
						{Label: "Netral", Value: "3"},
						{Label: "Nyaman", Value: "4"},
						{Label: "Sangat Nyaman", Value: "5"},
					},
					MaxScore:      5,
					QuestionOrder: 2,
					Category:      "personality",
				},
				{
					QuestionText: "Pilih peran yang paling menarik bagi Anda.",
					QuestionType: questionnaire.QuestionTypeMCQ,
					Options: []questionnaire.QuestionOption{
						{Label: "Frontend Developer", Value: "frontend"},
						{Label: "Backend Developer", Value: "backend"},
						{Label: "Data Scientist", Value: "data_scientist"},
					},
					MaxScore:      1,
					QuestionOrder: 3,
					Category:      "interests",
				},
			},
		},
		{
			Name:        "Kuesioner Profiling Karir - Bisnis",
			GeneratedBy: "manual",
			Questions: []struct {
				QuestionText  string
				QuestionType  questionnaire.QuestionType
				Options       []questionnaire.QuestionOption
				MaxScore      int
				QuestionOrder int
				Category      string
			}{
				{
					QuestionText:  "Apa yang memotivasi Anda untuk memulai bisnis?",
					QuestionType:  questionnaire.QuestionTypeText,
					MaxScore:      0,
					QuestionOrder: 1,
					Category:      "preferences",
				},
				{
					QuestionText: "Seberapa baik Anda dalam mengambil risiko?",
					QuestionType: questionnaire.QuestionTypeLikert,
					Options: []questionnaire.QuestionOption{
						{Label: "Sangat Buruk", Value: "1"},
						{Label: "Buruk", Value: "2"},
						{Label: "Netral", Value: "3"},
						{Label: "Baik", Value: "4"},
						{Label: "Sangat Baik", Value: "5"},
					},
					MaxScore:      5,
					QuestionOrder: 2,
					Category:      "personality",
				},
				{
					QuestionText: "Pilih bidang bisnis yang paling menarik bagi Anda.",
					QuestionType: questionnaire.QuestionTypeMCQ,
					Options: []questionnaire.QuestionOption{
						{Label: "E-commerce", Value: "ecommerce"},
						{Label: "Manufaktur", Value: "manufacturing"},
						{Label: "Jasa Keuangan", Value: "finance"},
					},
					MaxScore:      1,
					QuestionOrder: 3,
					Category:      "interests",
				},
			},
		},
	}

	for _, qData := range questionnaires {
		var existingQuestionnaire questionnaire.ProfilingQuestionnaire
		err := db.Where("name = ?", qData.Name).First(&existingQuestionnaire).Error

		if err == nil {
			log.Printf("Kuesioner dengan nama '%s' sudah ada, melewati...", qData.Name)
			continue
		}

		newQuestionnaire := questionnaire.ProfilingQuestionnaire{
			Name:        qData.Name,
			GeneratedBy: qData.GeneratedBy,
			Version:     1,
			Active:      false,
		}

		err = db.Create(&newQuestionnaire).Error
		if err != nil {
			return err
		}

		for _, q := range qData.Questions {
			var optionsJSON *string
			if len(q.Options) > 0 {
				optionsBytes, _ := json.Marshal(q.Options)
				optionsStr := string(optionsBytes)
				optionsJSON = &optionsStr
			}

			newQuestion := questionnaire.QuestionnaireQuestion{
				QuestionnaireID: newQuestionnaire.ID,
				QuestionText:    q.QuestionText,
				QuestionType:    q.QuestionType,
				Options:         optionsJSON,
				MaxScore:        q.MaxScore,
				QuestionOrder:   q.QuestionOrder,
				Category:        q.Category,
			}

			err = db.Create(&newQuestion).Error
			if err != nil {
				return err
			}
		}

		log.Printf("Kuesioner '%s' berhasil dibuat dengan %d pertanyaan", qData.Name, len(qData.Questions))
	}

	return nil
}
func SeedRoleRecommendations(db *gorm.DB) error {
	roles := []questionnaire.RoleRecommendation{
		{
			RoleName:    "Backend Developer",
			Description: "Mengembangkan aplikasi server-side, API, dan sistem database",
			Category:    "software_development",
		},
		{
			RoleName:    "Frontend Developer",
			Description: "Mengembangkan antarmuka pengguna dan pengalaman pengguna aplikasi web",
			Category:    "software_development",
		},
		{
			RoleName:    "Full Stack Developer",
			Description: "Mengembangkan aplikasi end-to-end dari frontend hingga backend",
			Category:    "software_development",
		},
		{
			RoleName:    "Mobile Developer",
			Description: "Mengembangkan aplikasi mobile untuk Android dan iOS",
			Category:    "mobile_development",
		},
		{
			RoleName:    "DevOps Engineer",
			Description: "Mengelola infrastruktur, deployment, dan operasi sistem",
			Category:    "infrastructure",
		},
		{
			RoleName:    "Data Scientist",
			Description: "Menganalisis data dan membangun model machine learning",
			Category:    "data_science",
		},
		{
			RoleName:    "UI/UX Designer",
			Description: "Merancang antarmuka dan pengalaman pengguna aplikasi",
			Category:    "design",
		},
		{
			RoleName:    "QA Engineer",
			Description: "Menguji kualitas software dan memastikan aplikasi bebas bug",
			Category:    "quality_assurance",
		},
		{
			RoleName:    "System Administrator",
			Description: "Mengelola infrastruktur IT dan sistem operasi",
			Category:    "infrastructure",
		},
		{
			RoleName:    "Cloud Engineer",
			Description: "Merancang dan mengelola infrastruktur cloud",
			Category:    "cloud_computing",
		},
		{
			RoleName:    "Cyber Security Specialist",
			Description: "Melindungi sistem dan data dari ancaman keamanan",
			Category:    "security",
		},
		{
			RoleName:    "Database Administrator",
			Description: "Mengelola dan mengoptimalkan sistem database enterprise",
			Category:    "database",
		},
	}

	for _, role := range roles {
		var existingRole questionnaire.RoleRecommendation
		err := db.Where("role_name = ?", role.RoleName).First(&existingRole).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&role).Error; err != nil {
				return fmt.Errorf("failed to create role %s: %w", role.RoleName, err)
			}
			log.Printf("Created role: %s", role.RoleName)
		}
	}

	return nil
}
