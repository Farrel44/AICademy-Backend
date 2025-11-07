package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/challenge"
	"github.com/Farrel44/AICademy-Backend/internal/domain/cv"
	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
	"github.com/Farrel44/AICademy-Backend/internal/domain/project"
	"github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire"
	"github.com/Farrel44/AICademy-Backend/internal/domain/roadmap"
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/google/uuid"

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

	if err := SeedTargetRoles(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding target roles: %v", err)
	}

	if err := SeedFeatureRoadmaps(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding roadmaps: %v", err)
	}

	if err := SeedStudentProjects(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding student projects: %v", err)
	}

	if err := SeedStudentCertifications(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding student certifications: %v", err)
	}

	if err := SeedQuestionnaireResponses(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding questionnaire responses: %v", err)
	}

	if err := SeedStudentRoadmapProgress(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding student roadmap progress: %v", err)
	}

	if err := SeedChallenges(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding challenges: %v", err)
	}

	if err := SeedInternships(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding internships: %v", err)
	}

	if err := SeedStudentCVs(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding student CVs: %v", err)
	}

	if err := SeedTeams(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding teams: %v", err)
	}

	if err := SeedChallengeSubmissions(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding challenge submissions: %v", err)
	}

	if err := SeedChallengeJudges(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding challenge judges: %v", err)
	}

	if err := SeedInternshipApplications(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding internship applications: %v", err)
	}

	if err := SeedInternshipReviews(db); err != nil {
		return fmt.Errorf("gagal melakukan seeding internship reviews: %v", err)
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

func SeedTargetRoles(db *gorm.DB) error {
	roles := []project.TargetRole{
		{ //0
			Name:        "Backend Developer",
			Description: "Mengembangkan aplikasi server-side, API, dan sistem database",
			Category:    "Technology",  
		},
		{//1
			Name:        "Frontend Developer",
			Description: "Mengembangkan antarmuka pengguna dan pengalaman pengguna aplikasi web",
			Category:    "Technology",
		},
		{//2
			Name:        "Full Stack Developer",
			Description: "Mengembangkan aplikasi end-to-end dari frontend hingga backend",
			Category:    "Technology",
		},
		{//3
			Name:        "Mobile Developer",
			Description: "Mengembangkan aplikasi mobile untuk Android dan iOS",
			Category:    "Technology",
		},
		{//4
			Name:        "DevOps Engineer",
			Description: "Mengelola infrastruktur, deployment, dan operasi sistem",
			Category:    "Technology",
		},
		{//5
			Name:        "Data Scientist",
			Description: "Menganalisis data dan membangun model machine learning",
			Category:    "Technology",
		},
		{//6
			Name:        "Data Analyst",
			Description: "Menganalisis data bisnis untuk menghasilkan insight dan laporan",
			Category:    "Technology",
		},
		{//7
			Name:        "Machine Learning Engineer",
			Description: "Mengimplementasikan dan deploy model machine learning ke production",
			Category:    "Technology",
		},
		{//8
			Name:        "UI/UX Designer",
			Description: "Merancang antarmuka dan pengalaman pengguna aplikasi",
			Category:    "Creative",
		},
		{//9
			Name:        "Graphic Designer",
			Description: "Membuat desain visual untuk berbagai media dan platform",
			Category:    "Creative",
		},
		{//10
			Name:        "Product Designer",
			Description: "Merancang produk digital dari konsep hingga implementasi",
			Category:    "Creative",
		},
		{//11
			Name:        "QA Engineer",
			Description: "Menguji kualitas software dan memastikan aplikasi bebas bug",
			Category:    "Technology",
		},
		{//12
			Name:        "Test Automation Engineer",
			Description: "Mengembangkan dan mengelola automated testing frameworks",
			Category:    "Technology",
		},
		{//13
			Name:        "System Administrator",
			Description: "Mengelola infrastruktur IT dan sistem operasi",
			Category:    "Technology",
		},
		{//14
			Name:        "Cloud Engineer",
			Description: "Merancang dan mengelola infrastruktur cloud",
			Category:    "Technology",
		},
		{//15
			Name:        "Cloud Architect",
			Description: "Merancang arsitektur cloud untuk aplikasi enterprise",
			Category:    "Technology",
		},
		{//16
			Name:        "Cyber Security Specialist",
			Description: "Melindungi sistem dan data dari ancaman keamanan",
			Category:    "Technology",
		},
		{//17
			Name:        "Security Analyst",
			Description: "Menganalisis ancaman keamanan dan implementasi solusi keamanan",
			Category:    "Technology",
		},
		{//18
			Name:        "Database Administrator",
			Description: "Mengelola dan mengoptimalkan sistem database enterprise",
			Category:    "Technology",
		},
		{//19
			Name:        "Database Developer",
			Description: "Mengembangkan struktur database dan stored procedures",
			Category:    "Technology",
		},
		{//20
			Name:        "Product Manager",
			Description: "Mengelola pengembangan produk dari perencanaan hingga peluncuran",
			Category:    "Business",
		},
		{//21
			Name:        "Technical Product Manager",
			Description: "Mengelola produk teknologi dengan fokus pada aspek teknis",
			Category:    "Business",
		},
		{//22
			Name:        "Business Analyst",
			Description: "Menganalisis kebutuhan bisnis dan merancang solusi teknologi",
			Category:    "Business",
		},
		{//23
			Name:        "Systems Analyst",
			Description: "Menganalisis sistem informasi dan merancang perbaikan",
			Category:    "Business",
		},
		{//24
			Name:        "Game Developer",
			Description: "Mengembangkan game untuk berbagai platform",
			Category:    "Technology",
		},
		{//25
			Name:        "Game Designer",
			Description: "Merancang gameplay, level, dan experience dalam game",
			Category:    "Creative",
		},
		{//26
			Name:        "Blockchain Developer",
			Description: "Mengembangkan aplikasi dan smart contracts berbasis blockchain",
			Category:    "Technology",
		},
		{//27
			Name:        "IoT Developer",
			Description: "Mengembangkan sistem Internet of Things dan embedded systems",
			Category:    "Technology",
		},
		{//28
			Name:        "AI Engineer",
			Description: "Mengembangkan solusi artificial intelligence dan deep learning",
			Category:    "Technology",
		},
		{//29
			Name:        "Robotics Engineer",
			Description: "Merancang dan mengembangkan sistem robotika",
			Category:    "Technology",
		},
	}

	for _, role := range roles {
		var existingRole project.TargetRole
		err := db.Where("name = ?", role.Name).First(&existingRole).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&role).Error; err != nil {
				return fmt.Errorf("failed to create target role %s: %w", role.Name, err)
			}
			log.Printf("Created target role: %s", role.Name)
		}
	}

	return nil
}

// SeedFeatureRoadmaps seeds roadmap templates for different roles
func SeedFeatureRoadmaps(db *gorm.DB) error {
	log.Println("Seeding feature roadmaps...")

	// Get admin user for created_by
	var admin user.User
	if err := db.Where("role = ?", user.RoleAdmin).First(&admin).Error; err != nil {
		return fmt.Errorf("admin user not found: %v", err)
	}

	// Get some target roles
	var targetRoles []project.TargetRole
	if err := db.Find(&targetRoles).Error; err != nil {
		return fmt.Errorf("target roles not found: %v", err)
	}

	roadmaps := []struct {
		Name        string
		Description string
		RoleIndex   int
		Steps       []struct {
			Order                int
			Title                string
			Description          string
			LearningObjectives   string
			SubmissionGuidelines string
			EstimatedDuration    int
			DifficultyLevel      string
		}
	}{
		{
			Name:        "Roadmap Backend Developer",
			Description: "Jalur pembelajaran untuk menjadi Backend Developer yang handal",
			RoleIndex:   0,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Dasar-dasar Programming",
					Description:          "Pelajari fundamental programming dan konsep dasar algoritma",
					LearningObjectives:   "Memahami variabel, function, loop, dan conditional statement",
					SubmissionGuidelines: "Upload screenshot atau video demo program sederhana yang dibuat",
					EstimatedDuration:    20,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Database Management",
					Description:          "Pelajari cara mengelola database dan membuat query SQL",
					LearningObjectives:   "Memahami DDL, DML, dan konsep normalisasi database",
					SubmissionGuidelines: "Submit database schema dan sample queries untuk aplikasi sederhana",
					EstimatedDuration:    30,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "RESTful API Development",
					Description:          "Belajar membuat API backend dengan konsep REST",
					LearningObjectives:   "Mampu membuat CRUD API dengan authentication",
					SubmissionGuidelines: "Deploy API ke server dan berikan dokumentasi lengkap",
					EstimatedDuration:    40,
					DifficultyLevel:      "advanced",
				},
				{
					Order:                4,
					Title:                "Testing & Documentation",
					Description:          "Pelajari unit testing dan API documentation",
					LearningObjectives:   "Membuat test coverage minimal 80% dan dokumentasi API lengkap",
					SubmissionGuidelines: "Submit link repository dengan test coverage report dan dokumentasi API",
					EstimatedDuration:    25,
					DifficultyLevel:      "intermediate",
				},
			},
		},
		{
			Name:        "Roadmap Frontend Developer",
			Description: "Jalur pembelajaran untuk menjadi Frontend Developer yang kompeten",
			RoleIndex:   1,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "HTML & CSS Fundamentals",
					Description:          "Kuasai dasar-dasar HTML dan CSS untuk membangun website",
					LearningObjectives:   "Membuat layout responsive dengan HTML5 dan CSS3",
					SubmissionGuidelines: "Buat website portfolio sederhana yang responsive",
					EstimatedDuration:    25,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "JavaScript & DOM Manipulation",
					Description:          "Pelajari JavaScript dan cara memanipulasi DOM",
					LearningObjectives:   "Membuat interactive web dengan vanilla JavaScript",
					SubmissionGuidelines: "Upload project interactive website dengan fitur dinamis",
					EstimatedDuration:    35,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Modern Frontend Framework",
					Description:          "Pelajari React.js atau Vue.js untuk building aplikasi modern",
					LearningObjectives:   "Membuat single page application dengan state management",
					SubmissionGuidelines: "Deploy aplikasi web ke hosting dan berikan source code",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},

		{
			Name:        "Roadmap Full Stack Developer",
			Description: "Jalur pembelajaran untuk menjadi Full Stack Developer yang handal",
			RoleIndex:   2, // Full Stack Developer
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Frontend Fundamentals",
					Description:          "Pelajari HTML, CSS, dan JavaScript untuk development frontend",
					LearningObjectives:   "Mampu membuat website responsive dengan vanilla JavaScript",
					SubmissionGuidelines: "Upload portfolio website dengan minimal 3 halaman interaktif",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Backend Development",
					Description:          "Pelajari server-side programming dan database management",
					LearningObjectives:   "Membangun RESTful API dengan authentication dan database",
					SubmissionGuidelines: "Deploy backend API dengan dokumentasi lengkap",
					EstimatedDuration:    45,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Full Stack Integration",
					Description:          "Integrasikan frontend dan backend untuk aplikasi lengkap",
					LearningObjectives:   "Membuat aplikasi web full stack yang production-ready",
					SubmissionGuidelines: "Deploy aplikasi full stack dengan user authentication",
					EstimatedDuration:    50,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Mobile Developer",
			Description: "Jalur pembelajaran untuk menjadi Mobile Developer yang mahir",
			RoleIndex:   3,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Mobile UI/UX Design",
					Description:          "Pelajari prinsip desain mobile dan user experience",
					LearningObjectives:   "Memahami design patterns mobile dan membuat wireframe",
					SubmissionGuidelines: "Upload mockup aplikasi mobile dengan justifikasi design choices",
					EstimatedDuration:    30,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Cross-Platform Development",
					Description:          "Belajar Flutter atau React Native untuk development mobile",
					LearningObjectives:   "Membuat aplikasi mobile yang bisa berjalan di Android dan iOS",
					SubmissionGuidelines: "Submit APK/IPA file dan source code aplikasi mobile",
					EstimatedDuration:    50,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap DevOps Engineer",
			Description: "Jalur pembelajaran untuk menjadi DevOps Engineer yang kompeten",
			RoleIndex:   4,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Linux System Administration",
					Description:          "Kuasai administrasi sistem Linux dan command line",
					LearningObjectives:   "Mampu mengelola server Linux dan automation script",
					SubmissionGuidelines: "Upload dokumentasi setup server dan automation scripts",
					EstimatedDuration:    35,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                2,
					Title:                "Containerization & Orchestration",
					Description:          "Pelajari Docker dan Kubernetes untuk container management",
					LearningObjectives:   "Deploy aplikasi menggunakan Docker dan Kubernetes",
					SubmissionGuidelines: "Berikan link repository dengan Docker compose dan K8s manifests",
					EstimatedDuration:    40,
					DifficultyLevel:      "advanced",
				},
				{
					Order:                3,
					Title:                "CI/CD Pipeline",
					Description:          "Buat pipeline otomatis untuk testing dan deployment",
					LearningObjectives:   "Mengimplementasikan CI/CD dengan GitHub Actions atau Jenkins",
					SubmissionGuidelines: "Setup working CI/CD pipeline untuk aplikasi sample",
					EstimatedDuration:    30,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Data Scientist",
			Description: "Jalur pembelajaran untuk menjadi Data Scientist yang kompeten",
			RoleIndex:   5, // Data Scientist
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Statistics & Mathematics",
					Description:          "Pelajari statistik, probabilitas, dan matematika untuk data science",
					LearningObjectives:   "Memahami konsep statistik dan dapat melakukan analisis data dasar",
					SubmissionGuidelines: "Upload laporan analisis statistik dataset dengan Python/R",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                2,
					Title:                "Data Analysis & Visualization",
					Description:          "Pelajari pandas, matplotlib, seaborn untuk analisis dan visualisasi data",
					LearningObjectives:   "Mampu melakukan EDA dan membuat visualisasi data yang informatif",
					SubmissionGuidelines: "Submit Jupyter notebook dengan analisis dataset kompleks",
					EstimatedDuration:    45,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Machine Learning Implementation",
					Description:          "Implementasikan algoritma machine learning untuk solving business problems",
					LearningObjectives:   "Membangun dan deploy model ML untuk prediksi atau klasifikasi",
					SubmissionGuidelines: "Deploy model ML dengan API endpoint dan dokumentasi",
					EstimatedDuration:    60,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Data Analyst",
			Description: "Jalur pembelajaran untuk menjadi Data Analyst yang solid dalam SQL, EDA, dan dashboarding",
			RoleIndex:   6,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order: 1,
					Title: "Fundamental Data & SQL",
					Description: "Pelajari tipe data, normalisasi, dan SQL untuk query analitis",
					LearningObjectives: "Menguasai SELECT, JOIN, GROUP BY, window function dasar",
					SubmissionGuidelines: "Kirim notebook/skrip SQL yang menjawab 10 pertanyaan analitis pada dataset retail",
					EstimatedDuration: 30,
					DifficultyLevel:   "beginner",
				},
				{
					Order: 2,
					Title: "Exploratory Data Analysis (EDA)",
					Description: "Analisis data menggunakan Python (pandas, matplotlib/seaborn) atau R",
					LearningObjectives: "Membersihkan data, membuat visualisasi deskriptif, menarik insight",
					SubmissionGuidelines: "Upload notebook EDA dengan narasi dan minimal 5 chart yang relevan",
					EstimatedDuration: 40,
					DifficultyLevel:   "intermediate",
				},
				{
					Order: 3,
					Title: "Dashboard & Reporting",
					Description: "Membangun dashboard interaktif (Metabase/Power BI/Tableau)",
					LearningObjectives: "Merancang KPI, filter, dan storytelling data",
					SubmissionGuidelines: "Publish dashboard dan lampirkan link + definisi KPI",
					EstimatedDuration: 45,
					DifficultyLevel:   "advanced",
				},
			},
		},

		{
			Name:        "Roadmap Machine Learning Engineer",
			Description: "Jalur pembelajaran untuk menjadi ML Engineer yang production-ready",
			RoleIndex:   7, // Machine Learning Engineer
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "ML Fundamentals & Programming",
					Description:          "Pelajari Python, data structures, dan machine learning algorithms",
					LearningObjectives:   "Memahami algoritma ML dan dapat implement dari scratch",
					SubmissionGuidelines: "Upload implementation algoritma ML dasar dengan Python",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                2,
					Title:                "ML Pipeline Development",
					Description:          "Belajar data preprocessing, feature engineering, dan model evaluation",
					LearningObjectives:   "Mampu membangun ML pipeline yang robust dan scalable",
					SubmissionGuidelines: "Submit end-to-end ML pipeline dengan proper evaluation metrics",
					EstimatedDuration:    50,
					DifficultyLevel:      "advanced",
				},
				{
					Order:                3,
					Title:                "ML Model Deployment",
					Description:          "Deploy ML models ke production dengan monitoring dan maintenance",
					LearningObjectives:   "Memahami MLOps dan dapat deploy model ke cloud platform",
					SubmissionGuidelines: "Deploy ML model dengan API, monitoring, dan CI/CD pipeline",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap UI/UX Designer",
			Description: "Jalur pembelajaran untuk menjadi UI/UX Designer yang kreatif",
			RoleIndex:   8, // UI/UX Designer
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Design Fundamentals",
					Description:          "Pelajari prinsip dasar design, typography, color theory, dan layout",
					LearningObjectives:   "Memahami elemen design dan dapat membuat composition yang baik",
					SubmissionGuidelines: "Upload portfolio design dasar dengan 5 design variations",
					EstimatedDuration:    30,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "User Research & Wireframing",
					Description:          "Pelajari user research methods, persona creation, dan wireframing",
					LearningObjectives:   "Mampu melakukan user research dan membuat wireframe yang user-centered",
					SubmissionGuidelines: "Submit case study dengan user research findings dan wireframes",
					EstimatedDuration:    35,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Prototyping & User Testing",
					Description:          "Buat prototype interaktif dan lakukan user testing",
					LearningObjectives:   "Membuat prototype high-fidelity dan menguji usability dengan users",
					SubmissionGuidelines: "Upload prototype interaktif dengan user testing report",
					EstimatedDuration:    40,
					DifficultyLevel:      "advanced",
				},
			},
				},
		{
			Name:        "Roadmap Graphic Designer",
			Description: "Jalur pembelajaran untuk menjadi Graphic Designer yang kuat di brand & marketing assets",
			RoleIndex:   9,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Dasar Desain & Tools",
					Description:          "Prinsip desain, tipografi, warna; pengenalan Figma/Adobe",
					LearningObjectives:   "Menerapkan prinsip kontras, hierarki, dan grid",
					SubmissionGuidelines: "Upload 3 poster sederhana dengan variasi tipografi/warna",
					EstimatedDuration:    25,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Branding & Ilustrasi Vektor",
					Description:          "Logo, palet warna, ikon/ilustrasi vektor",
					LearningObjectives:   "Mendesain identitas visual yang konsisten",
					SubmissionGuidelines: "Kirim brand kit (logo, palet, ikon) dalam 1 paket",
					EstimatedDuration:    35,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Marketing Assets & Export",
					Description:          "Layout social media, banner, cetak; format & ekspor",
					LearningObjectives:   "Menyiapkan file siap produksi",
					SubmissionGuidelines: "Upload 5 aset kampanye beserta file sumber",
					EstimatedDuration:    40,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Product Designer",
			Description: "Jalur pembelajaran untuk Product Designer dari riset hingga handoff",
			RoleIndex:   10,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Discovery & UX Research",
					Description:          "Persona, problem framing, user interview, JTBD",
					LearningObjectives:   "Merumuskan problem statement & insight riset",
					SubmissionGuidelines: "Case study ringkas berisi temuan riset dan persona",
					EstimatedDuration:    30,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Wireframing & Prototyping",
					Description:          "Low–hi fidelity di Figma; arsitektur informasi",
					LearningObjectives:   "Mendesain alur end-to-end yang usable",
					SubmissionGuidelines: "Prototype interaktif + skenario tugas",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Design System & Handoff",
					Description:          "Komponen, tokens, dokumentasi, kolaborasi dev",
					LearningObjectives:   "Menyusun design system & spesifikasi",
					SubmissionGuidelines: "Link design system + file handoff lengkap",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap QA Engineer",
			Description: "Jalur pembelajaran untuk menjadi QA Engineer yang detail-oriented",
			RoleIndex:   11, // QA Engineer
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Software Testing Fundamentals",
					Description:          "Pelajari testing methodologies, test cases, dan bug reporting",
					LearningObjectives:   "Memahami SDLC, testing types, dan dapat membuat test cases yang comprehensive",
					SubmissionGuidelines: "Submit test plan dan test cases untuk aplikasi web sample",
					EstimatedDuration:    30,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Manual Testing Excellence",
					Description:          "Kuasai manual testing, exploratory testing, dan defect management",
					LearningObjectives:   "Mampu melakukan testing manual yang thorough dan reporting bugs yang clear",
					SubmissionGuidelines: "Upload bug reports dan testing documentation untuk real application",
					EstimatedDuration:    35,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Test Automation Basics",
					Description:          "Pelajari automation testing tools seperti Selenium dan API testing",
					LearningObjectives:   "Membuat automated test scripts untuk UI dan API testing",
					SubmissionGuidelines: "Deploy automation test suite dengan CI/CD integration",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Test Automation Engineer",
			Description: "Jalur untuk spesialisasi automation testing (UI, API, CI)",
			RoleIndex:   12,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Automation Fundamentals",
					Description:          "Selenium/WebDriver, locator strategy, Page Object",
					LearningObjectives:   "Menulis skrip UI automation yang stabil",
					SubmissionGuidelines: "Repo dengan 5 test UI berbasis Page Object",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "API Test Automation",
					Description:          "REST Assured/Newman, schema validation, data-driven test",
					LearningObjectives:   "Mengotomasi test API & assertions",
					SubmissionGuidelines: "Suite API tests + laporan hasil run",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "CI Integration & Reporting",
					Description:          "Integrasi GitHub Actions/Jenkins, flaky test handling",
					LearningObjectives:   "Pipeline otomatis & laporan (mis. Allure)",
					SubmissionGuidelines: "Pipeline CI yang menjalankan test & publish report",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap System Administrator",
			Description: "Jalur pembelajaran untuk SysAdmin on-prem & cloud dasar",
			RoleIndex:   13,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Linux & Networking Basics",
					Description:          "User/group, filesystem, systemd; TCP/IP, DNS, firewall",
					LearningObjectives:   "Mengelola server & layanan jaringan dasar",
					SubmissionGuidelines: "Dokumentasi setup server + hardening dasar",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Services, Backup, Monitoring",
					Description:          "Web/proxy, database basic; backup & alert",
					LearningObjectives:   "Menyusun backup & monitoring server",
					SubmissionGuidelines: "Konfigurasi backup + dashboard monitoring",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Troubleshooting & Scripting",
					Description:          "Diagnostik performa & shell scripting",
					LearningObjectives:   "Menyelesaikan insiden & automasi rutin",
					SubmissionGuidelines: "Kumpulan skrip otomasi + postmortem 1 insiden",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Cloud Engineer",
			Description: "Jalur pembelajaran untuk menjadi Cloud Engineer yang reliable",
			RoleIndex:   14, // Cloud Engineer
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Cloud Fundamentals",
					Description:          "Pelajari cloud computing concepts, AWS/Azure/GCP basics",
					LearningObjectives:   "Memahami cloud services dan dapat deploy aplikasi sederhana",
					SubmissionGuidelines: "Deploy web application ke cloud platform dengan dokumentasi",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Infrastructure as Code",
					Description:          "Pelajari Terraform, CloudFormation untuk infrastructure automation",
					LearningObjectives:   "Mampu manage infrastructure menggunakan code dan version control",
					SubmissionGuidelines: "Submit IaC scripts untuk deploy complete infrastructure",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Cloud Architecture & Security",
					Description:          "Design scalable cloud architecture dengan security best practices",
					LearningObjectives:   "Merancang dan implement secure, scalable cloud solutions",
					SubmissionGuidelines: "Design dan deploy production-ready cloud architecture",
					EstimatedDuration:    50,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Cloud Architect",
			Description: "Jalur untuk merancang arsitektur cloud yang scalable & cost-aware",
			RoleIndex:   15,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Architecture Patterns",
					Description:          "Microservices, event-driven, serverless; trade-off",
					LearningObjectives:   "Memilih pola arsitektur sesuai kebutuhan",
					SubmissionGuidelines: "Dokumen arsitektur solusi studi kasus",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Reliability & Cost",
					Description:          "Well-Architected, HA/DR, autoscaling, FinOps",
					LearningObjectives:   "Menyeimbangkan SLO, biaya, performa",
					SubmissionGuidelines: "DR plan + perhitungan biaya komponen",
					EstimatedDuration:    45,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Security & Governance",
					Description:          "IAM, jaringan, enkripsi, compliance",
					LearningObjectives:   "Merancang kontrol keamanan menyeluruh",
					SubmissionGuidelines: "Rancangan kontrol keamanan & guardrails",
					EstimatedDuration:    50,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Cyber Security Specialist",
			Description: "Jalur spesialis keamanan: dari dasar sampai operasi biru",
			RoleIndex:   16,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Security Fundamentals",
					Description:          "CIA triad, threat modeling, risk assessment",
					LearningObjectives:   "Menyusun baseline kebijakan keamanan",
					SubmissionGuidelines: "Dokumen threat model untuk web app sederhana",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Network & Web Security",
					Description:          "Segmentation, TLS, OWASP Top 10, secure coding",
					LearningObjectives:   "Mengidentifikasi & memitigasi kerentanan umum",
					SubmissionGuidelines: "Laporan temuan + rekomendasi mitigasi",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Blue Team Operations",
					Description:          "SIEM, detections, playbook IR, forensik dasar",
					LearningObjectives:   "Menangani insiden dari deteksi hingga recovery",
					SubmissionGuidelines: "Playbook IR + simulasi insiden terekam",
					EstimatedDuration:    50,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Security Analyst",
			Description: "Jalur analis keamanan berfokus monitoring, vulnerability, dan IR",
			RoleIndex:   17,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "SIEM & Log Analysis",
					Description:          "Kumpulan log, parsing, korelasi, alerting",
					LearningObjectives:   "Membangun rule deteksi efektif",
					SubmissionGuidelines: "Rule SIEM + laporan alert triage",
					EstimatedDuration:    30,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Vulnerability Management",
					Description:          "Scanning, CVSS, prioritisasi & remedi",
					LearningObjectives:   "Menyusun siklus vuln management",
					SubmissionGuidelines: "Report pemindaian + rencana remedi",
					EstimatedDuration:    35,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Incident Response",
					Description:          "Proses IR, komunikasi, pelaporan pasca-insiden",
					LearningObjectives:   "Menangani insiden end-to-end",
					SubmissionGuidelines: "Post-incident report (template disediakan)",
					EstimatedDuration:    40,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Database Administrator",
			Description: "Jalur untuk menjadi DBA yang andal",
			RoleIndex:   18,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "RDBMS Fundamentals",
					Description:          "Instalasi, skema, permission, backup dasar",
					LearningObjectives:   "Mengelola instance & keamanan dasar",
					SubmissionGuidelines: "Dokumentasi setup + skenario backup/restore",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "High Availability & Replication",
					Description:          "Streaming replication, failover, recovery",
					LearningObjectives:   "Menyusun HA & DR untuk database",
					SubmissionGuidelines: "Simulasi failover + hasil uji",
					EstimatedDuration:    45,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Performance & Monitoring",
					Description:          "Indexing, EXPLAIN, tuning, observability",
					LearningObjectives:   "Mengoptimalkan query & resource",
					SubmissionGuidelines: "Laporan tuning dari workload nyata",
					EstimatedDuration:    50,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Database Developer",
			Description: "Jalur untuk developer yang fokus pada desain & logika DB",
			RoleIndex:   19,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "SQL Lanjut & Desain Skema",
					Description:          "Constraint, indexing, normalisasi/denormalisasi",
					LearningObjectives:   "Membuat skema efisien & konsisten",
					SubmissionGuidelines: "ERD + skrip pembuatan skema",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Stored Procedures & Transaksi",
					Description:          "Function, trigger, ACID, concurrency",
					LearningObjectives:   "Menulis SP/trigger aman & performan",
					SubmissionGuidelines: "Repo berisi SP/trigger + test",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Optimasi & Migrasi",
					Description:          "EXPLAIN, partisi, migrasi skema/data",
					LearningObjectives:   "Mengoptimasi query & rencana migrasi",
					SubmissionGuidelines: "Rencana migrasi + hasil benchmark",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Product Manager",
			Description: "Jalur PM dari discovery sampai delivery & iterasi",
			RoleIndex:   20,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Product Discovery & Strategy",
					Description:          "Market/user research, value proposition, OKR",
					LearningObjectives:   "Merumuskan strategi & tujuan produk",
					SubmissionGuidelines: "PRD ringkas + OKR Q1 untuk produk contoh",
					EstimatedDuration:    30,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Prioritization & Roadmapping",
					Description:          "Kano, RICE, roadmap & dependency",
					LearningObjectives:   "Menyusun roadmap realistis",
					SubmissionGuidelines: "Roadmap 2 kuartal + rationale prioritas",
					EstimatedDuration:    35,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Delivery & Analytics",
					Description:          "Backlog, release plan, eksperimen & telemetry",
					LearningObjectives:   "Mengukur dampak dan iterasi",
					SubmissionGuidelines: "Plan eksperimen + metrik keberhasilan",
					EstimatedDuration:    40,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Technical Product Manager",
			Description: "Jalur TPM yang menjembatani bisnis & teknis",
			RoleIndex:   21,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "System Fundamentals & APIs",
					Description:          "Arsitektur web, REST/gRPC, integrasi",
					LearningObjectives:   "Membaca/membuat spesifikasi teknis",
					SubmissionGuidelines: "Spec API + kontrak & skenario integrasi",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Translating Requirements",
					Description:          "Dari problem → requirement → desain teknis",
					LearningObjectives:   "Menulis PRD teknis dan acceptance criteria",
					SubmissionGuidelines: "PRD teknis + acceptance tests",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Tech Roadmap & Lifecycle",
					Description:          "Skalabilitas, keamanan, deprecations",
					LearningObjectives:   "Mengelola lifecycle komponen",
					SubmissionGuidelines: "Tech roadmap 12 bulan + risk register",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Business Analyst",
			Description: "Jalur BA untuk analisis kebutuhan & solusi",
			RoleIndex:   22,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Requirements Elicitation",
					Description:          "Interview, workshop, dokumentasi kebutuhan",
					LearningObjectives:   "Mendefinisikan kebutuhan yang testable",
					SubmissionGuidelines: "Dokumen kebutuhan + use cases",
					EstimatedDuration:    30,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Modeling & Process",
					Description:          "BPMN/UML, mapping proses AS-IS/TO-BE",
					LearningObjectives:   "Menyusun model proses yang jelas",
					SubmissionGuidelines: "Diagram proses + analisis gap",
					EstimatedDuration:    35,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Solution Evaluation & UAT",
					Description:          "Kriteria evaluasi, UAT, sign-off",
					LearningObjectives:   "Memvalidasi solusi & nilai bisnis",
					SubmissionGuidelines: "Rencana UAT + laporan hasil",
					EstimatedDuration:    40,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Systems Analyst",
			Description: "Jalur analis sistem berfokus desain & integrasi",
			RoleIndex:   23,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Analisis Kebutuhan & Spesifikasi",
					Description:          "Teknik analisis & dokumentasi teknis",
					LearningObjectives:   "Menulis spesifikasi sistem yang jelas",
					SubmissionGuidelines: "Dokumen spesifikasi untuk modul inti",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Desain Arsitektur & Integrasi",
					Description:          "Komponen, antarmuka, integrasi antar sistem",
					LearningObjectives:   "Merancang integrasi yang robust",
					SubmissionGuidelines: "Diagram arsitektur + kontrak integrasi",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Dokumentasi & Handover",
					Description:          "Dokumentasi teknis, knowledge transfer",
					LearningObjectives:   "Menjamin kelangsungan pemeliharaan",
					SubmissionGuidelines: "Paket dokumentasi + sesi knowledge transfer",
					EstimatedDuration:    40,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Game Developer",
			Description: "Jalur dev game menggunakan Unity/Godot",
			RoleIndex:   24,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Fundamental Pemrograman Game",
					Description:          "Game loop, input, scene, physics dasar",
					LearningObjectives:   "Membangun prototype gameplay dasar",
					SubmissionGuidelines: "Prototype mini game + kode sumber",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Gameplay & Systems",
					Description:          "Spawner, inventory, AI sederhana",
					LearningObjectives:   "Mengimplementasi sistem gameplay inti",
					SubmissionGuidelines: "Demo fitur gameplay + penjelasan",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Optimasi & Build",
					Description:          "Profiling, asset management, build multi-platform",
					LearningObjectives:   "Membuat build yang stabil & efisien",
					SubmissionGuidelines: "Build + catatan optimasi",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Game Designer",
			Description: "Jalur perancangan game: mekanik, level, dan balancing",
			RoleIndex:   25,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Game Design Principles",
					Description:          "Core loop, mekanik, motivasi pemain",
					LearningObjectives:   "Mendefinisikan high-level concept & loop",
					SubmissionGuidelines: "GDD ringkas untuk game casual",
					EstimatedDuration:    30,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Level Design & Balancing",
					Description:          "Pacing, difficulty curve, metrics",
					LearningObjectives:   "Merancang level yang engaging",
					SubmissionGuidelines: "2 level playable + rationale balancing",
					EstimatedDuration:    35,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Prototyping & Playtesting",
					Description:          "Rapid prototyping, feedback loop",
					LearningObjectives:   "Mengiterasi desain dari data playtest",
					SubmissionGuidelines: "Laporan playtest + iterasi desain",
					EstimatedDuration:    40,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Blockchain Developer",
			Description: "Jalur pengembangan blockchain & smart contract",
			RoleIndex:   26,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Dasar Blockchain",
					Description:          "Blok, konsensus, wallet, transaksi",
					LearningObjectives:   "Memahami arsitektur blockchain modern",
					SubmissionGuidelines: "Ringkasan konsep + diagram alur transaksi",
					EstimatedDuration:    30,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Smart Contract Development",
					Description:          "Solidity, standard ERC, testing",
					LearningObjectives:   "Menulis & menguji smart contract",
					SubmissionGuidelines: "Repo kontrak + unit test lulus",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "DApp & Security",
					Description:          "Integrasi web3, audit dasar, best practices",
					LearningObjectives:   "Membangun DApp end-to-end yang aman",
					SubmissionGuidelines: "DApp live + catatan audit dasar",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap IoT Developer",
			Description: "Jalur pengembangan sistem IoT dari device ke cloud",
			RoleIndex:   27,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Embedded & Sensor Dasar",
					Description:          "Microcontroller, GPIO, sensor/actuator",
					LearningObjectives:   "Membaca sensor & mengontrol actuator",
					SubmissionGuidelines: "Demo pembacaan 2 sensor + kode",
					EstimatedDuration:    35,
					DifficultyLevel:      "beginner",
				},
				{
					Order:                2,
					Title:                "Konektivitas & Edge",
					Description:          "MQTT/HTTP, payload, edge processing",
					LearningObjectives:   "Mengirim data ke broker/cloud",
					SubmissionGuidelines: "Publikasi data realtime ke broker MQTT",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                3,
					Title:                "Cloud Integration & Dashboard",
					Description:          "Ingest, storage, alerting, visualisasi",
					LearningObjectives:   "Menyusun pipeline device→cloud→dashboard",
					SubmissionGuidelines: "Dashboard IoT + alert rule aktif",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap AI Engineer",
			Description: "Jalur AI Engineer fokus deployment & skalabilitas model",
			RoleIndex:   28,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Deep Learning Foundations",
					Description:          "Neural network dasar, training loop, PyTorch/TensorFlow",
					LearningObjectives:   "Membangun & melatih model DL sederhana",
					SubmissionGuidelines: "Notebook training + hasil evaluasi",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                2,
					Title:                "Serving & Retrieval",
					Description:          "ONNX/Triton/TF Serving, vector DB, RAG",
					LearningObjectives:   "Menyajikan model dengan latensi rendah",
					SubmissionGuidelines: "Endpoint inferensi + latensi tercatat",
					EstimatedDuration:    50,
					DifficultyLevel:      "advanced",
				},
				{
					Order:                3,
					Title:                "MLOps & Monitoring",
					Description:          "CI/CD model, drift, observabilitas",
					LearningObjectives:   "Mengelola siklus hidup model produksi",
					SubmissionGuidelines: "Pipeline CI/CD + dashboard monitoring",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
			},
		},
		{
			Name:        "Roadmap Robotics Engineer",
			Description: "Jalur rekayasa robotika: kontrol, persepsi, navigasi",
			RoleIndex:   29,
			Steps: []struct {
				Order                int
				Title                string
				Description          string
				LearningObjectives   string
				SubmissionGuidelines string
				EstimatedDuration    int
				DifficultyLevel      string
			}{
				{
					Order:                1,
					Title:                "Kinematika & Kontrol Dasar",
					Description:          "Kinematika forward/inverse, kontrol PID",
					LearningObjectives:   "Mengendalikan aktuator dengan stabil",
					SubmissionGuidelines: "Demo kontrol motor + analisis respons",
					EstimatedDuration:    40,
					DifficultyLevel:      "intermediate",
				},
				{
					Order:                2,
					Title:                "ROS & Integrasi Sensor",
					Description:          "ROS/ROS2, topic, tf, kamera/LiDAR/IMU",
					LearningObjectives:   "Membangun node & pipeline sensor",
					SubmissionGuidelines: "Proyek ROS sederhana + bagfile",
					EstimatedDuration:    45,
					DifficultyLevel:      "advanced",
				},
				{
					Order:                3,
					Title:                "Navigasi & SLAM",
					Description:          "Path planning, mapping, SLAM",
					LearningObjectives:   "Menavigasi lingkungan baru secara otonom",
					SubmissionGuidelines: "Demo SLAM di simulator + laporan",
					EstimatedDuration:    50,
					DifficultyLevel:      "advanced",
				},
			},
		},
	}

	for _, roadmapData := range roadmaps {
		if roadmapData.RoleIndex >= len(targetRoles) {
			continue
		}

		var existingRoadmap roadmap.FeatureRoadmap
		err := db.Where("roadmap_name = ?", roadmapData.Name).First(&existingRoadmap).Error
		if err == nil {
			log.Printf("Roadmap '%s' sudah ada, melewati...", roadmapData.Name)
			continue
		}

		newRoadmap := roadmap.FeatureRoadmap{
			ProfilingRoleID: targetRoles[roadmapData.RoleIndex].ID,
			RoadmapName:     roadmapData.Name,
			Description:     &roadmapData.Description,
			Status:          roadmap.RoadmapStatusActive,
			CreatedBy:       admin.ID,
		}

		if err := db.Create(&newRoadmap).Error; err != nil {
			return fmt.Errorf("failed to create roadmap %s: %v", roadmapData.Name, err)
		}

		// Create steps
		for _, stepData := range roadmapData.Steps {
			step := roadmap.RoadmapStep{
				RoadmapID:            newRoadmap.ID,
				StepOrder:            stepData.Order,
				Title:                stepData.Title,
				Description:          stepData.Description,
				LearningObjectives:   stepData.LearningObjectives,
				SubmissionGuidelines: stepData.SubmissionGuidelines,
				EstimatedDuration:    stepData.EstimatedDuration,
				DifficultyLevel:      stepData.DifficultyLevel,
			}

			if err := db.Create(&step).Error; err != nil {
				return fmt.Errorf("failed to create roadmap step: %v", err)
			}
		}

		log.Printf("Roadmap '%s' berhasil dibuat dengan %d steps", roadmapData.Name, len(roadmapData.Steps))
	}

	return nil
}

// SeedStudentProjects seeds sample projects for students
func SeedStudentProjects(db *gorm.DB) error {
	log.Println("Seeding student projects...")

	// Get student profiles
	var studentProfiles []user.StudentProfile
	if err := db.Limit(5).Find(&studentProfiles).Error; err != nil {
		return fmt.Errorf("student profiles not found: %v", err)
	}

	// Get target roles
	var targetRoles []project.TargetRole
	if err := db.Limit(3).Find(&targetRoles).Error; err != nil {
		return fmt.Errorf("target roles not found: %v", err)
	}

	projects := []struct {
		Name        string
		Description string
		TechStack   []string
		StartDate   time.Time
		EndDate     time.Time
		LinkURL     string
	}{
		{
			Name:        "E-Commerce Website",
			Description: "Website e-commerce lengkap dengan fitur keranjang belanja, payment gateway, dan admin dashboard",
			TechStack:   []string{"React", "Node.js", "PostgreSQL", "Express"},
			StartDate:   time.Now().AddDate(0, -3, 0),
			EndDate:     time.Now().AddDate(0, -1, 0),
			LinkURL:     "https://github.com/student/ecommerce-project",
		},
		{
			Name:        "Task Management App",
			Description: "Aplikasi manajemen tugas dengan fitur real-time collaboration dan notification",
			TechStack:   []string{"Vue.js", "Firebase", "Tailwind CSS"},
			StartDate:   time.Now().AddDate(0, -4, 0),
			EndDate:     time.Now().AddDate(0, -2, 0),
			LinkURL:     "https://github.com/student/task-management",
		},
		{
			Name:        "Weather Dashboard",
			Description: "Dashboard cuaca dengan visualisasi data dan prediksi menggunaan machine learning",
			TechStack:   []string{"Python", "Django", "Chart.js", "OpenWeather API"},
			StartDate:   time.Now().AddDate(0, -2, 0),
			EndDate:     time.Now().AddDate(0, 0, -15),
			LinkURL:     "https://github.com/student/weather-dashboard",
		},
	}

	for i, projectData := range projects {
		if i >= len(studentProfiles) {
			break
		}

		var existingProject project.Project
		err := db.Where("project_name = ? AND owner_student_profile_id = ?",
			projectData.Name, studentProfiles[i].ID).First(&existingProject).Error
		if err == nil {
			log.Printf("Project '%s' sudah ada untuk student %s, melewati...", projectData.Name, studentProfiles[i].Fullname)
			continue
		}

		// Create project using raw SQL to handle PostgreSQL array properly
		projectID := uuid.New()

		// Build array string for PostgreSQL
		techStackArray := "ARRAY["
		for j, tech := range projectData.TechStack {
			if j > 0 {
				techStackArray += ","
			}
			techStackArray += "'" + strings.ReplaceAll(tech, "'", "''") + "'"
		}
		techStackArray += "]"

		err = db.Exec(fmt.Sprintf(`
			INSERT INTO projects (id, owner_student_profile_id, project_name, description, link_url, tech_stack, start_date, end_date, created_at)
			VALUES (?, ?, ?, ?, ?, %s, ?, ?, ?)`,
			techStackArray),
			projectID, studentProfiles[i].ID, projectData.Name, projectData.Description, projectData.LinkURL,
			projectData.StartDate, projectData.EndDate, time.Now()).Error

		if err != nil {
			return fmt.Errorf("failed to create project %s: %v", projectData.Name, err)
		}

		newProject := project.Project{
			ID:                    projectID,
			OwnerStudentProfileID: studentProfiles[i].ID,
			ProjectName:           projectData.Name,
			Description:           projectData.Description,
			LinkURL:               &projectData.LinkURL,
			TechStack:             projectData.TechStack,
			StartDate:             projectData.StartDate,
			EndDate:               projectData.EndDate,
		}

		// Add contributors (including owner and some other students)
		contributors := []project.ProjectContributor{
			{
				ProjectID:        newProject.ID,
				StudentProfileID: studentProfiles[i].ID,
				ProjectRole:      stringPtr("Project Lead"),
				RoleID:           targetRoles[i%len(targetRoles)].ID,
			},
		}

		// Add 1-2 more contributors if available
		for j := 1; j <= 2 && (i+j) < len(studentProfiles); j++ {
			contributors = append(contributors, project.ProjectContributor{
				ProjectID:        newProject.ID,
				StudentProfileID: studentProfiles[i+j].ID,
				ProjectRole:      stringPtr("Developer"),
				RoleID:           targetRoles[(i+j)%len(targetRoles)].ID,
			})
		}

		for _, contributor := range contributors {
			if err := db.Create(&contributor).Error; err != nil {
				log.Printf("Warning: failed to create contributor: %v", err)
			}
		}

		log.Printf("Project '%s' berhasil dibuat untuk student %s dengan %d contributors",
			projectData.Name, studentProfiles[i].Fullname, len(contributors))
	}

	return nil
}

// SeedStudentCertifications seeds sample certifications for students
func SeedStudentCertifications(db *gorm.DB) error {
	log.Println("Seeding student certifications...")

	// Get student profiles
	var studentProfiles []user.StudentProfile
	if err := db.Limit(5).Find(&studentProfiles).Error; err != nil {
		return fmt.Errorf("student profiles not found: %v", err)
	}

	certifications := []struct {
		Name        string
		Issuer      string
		IssuedDate  time.Time
		ExpiryDate  *time.Time
		CertURL     string
		Description string
	}{
		{
			Name:        "AWS Certified Cloud Practitioner",
			Issuer:      "Amazon Web Services",
			IssuedDate:  time.Now().AddDate(0, -6, 0),
			ExpiryDate:  timePtr(time.Now().AddDate(3, -6, 0)),
			CertURL:     "https://aws.amazon.com/certification/certified-cloud-practitioner/",
			Description: "Entry-level AWS cloud certification covering basic cloud concepts",
		},
		{
			Name:        "Google Data Analytics Certificate",
			Issuer:      "Google",
			IssuedDate:  time.Now().AddDate(0, -4, 0),
			ExpiryDate:  nil,
			CertURL:     "https://grow.google/certificates/data-analytics/",
			Description: "Professional certificate in data analytics covering tools and techniques",
		},
		{
			Name:        "Microsoft Azure Fundamentals",
			Issuer:      "Microsoft",
			IssuedDate:  time.Now().AddDate(0, -8, 0),
			ExpiryDate:  timePtr(time.Now().AddDate(2, -8, 0)),
			CertURL:     "https://docs.microsoft.com/en-us/learn/certifications/azure-fundamentals/",
			Description: "Foundational knowledge of cloud services and Microsoft Azure",
		},
	}

	for i, certData := range certifications {
		if i >= len(studentProfiles) {
			break
		}

		var existingCert user.Certification
		err := db.Where("name = ? AND student_profile_id = ?",
			certData.Name, studentProfiles[i].ID).First(&existingCert).Error
		if err == nil {
			log.Printf("Certification '%s' sudah ada untuk student %s, melewati...", certData.Name, studentProfiles[i].Fullname)
			continue
		}

		newCert := user.Certification{
			ID:                  uuid.New(),
			StudentProfileID:    studentProfiles[i].ID,
			Name:                certData.Name,
			IssuingOrganization: certData.Issuer,
			IssueDate:           certData.IssuedDate,
			ExpirationDate:      certData.ExpiryDate,
			CredentialURL:       &certData.CertURL,
		}

		if err := db.Create(&newCert).Error; err != nil {
			return fmt.Errorf("failed to create certification %s: %v", certData.Name, err)
		}

		log.Printf("Certification '%s' berhasil dibuat untuk student %s", certData.Name, studentProfiles[i].Fullname)
	}

	return nil
}

// SeedQuestionnaireResponses seeds sample questionnaire responses
func SeedQuestionnaireResponses(db *gorm.DB) error {
	log.Println("Seeding questionnaire responses...")

	// Get students and questionnaires
	var studentProfiles []user.StudentProfile
	if err := db.Limit(3).Find(&studentProfiles).Error; err != nil {
		return fmt.Errorf("student profiles not found: %v", err)
	}

	var questionnaires []questionnaire.ProfilingQuestionnaire
	if err := db.Find(&questionnaires).Error; err != nil {
		return fmt.Errorf("questionnaires not found: %v", err)
	}

	// Get target roles for recommendations
	var targetRoles []project.TargetRole
	if err := db.Limit(3).Find(&targetRoles).Error; err != nil {
		return fmt.Errorf("target roles not found: %v", err)
	}

	responses := []map[string]interface{}{
		{
			"question_1": "Saya paling menikmati memecahkan masalah teknis dan menciptakan solusi yang efisien",
			"question_2": "4",
			"question_3": "backend",
		},
		{
			"question_1": "Saya senang membuat tampilan yang menarik dan user-friendly",
			"question_2": "5",
			"question_3": "frontend",
		},
		{
			"question_1": "Saya tertarik menganalisis data untuk menemukan insights bisnis",
			"question_2": "3",
			"question_3": "data_scientist",
		},
	}

	for i, studentProfile := range studentProfiles {
		if i >= len(questionnaires) || i >= len(responses) || i >= len(targetRoles) {
			continue
		}

		var existingResponse questionnaire.QuestionnaireResponse
		err := db.Where("student_profile_id = ? AND questionnaire_id = ?",
			studentProfile.ID, questionnaires[i].ID).First(&existingResponse).Error
		if err == nil {
			log.Printf("Response sudah ada untuk student %s, melewati...", studentProfile.Fullname)
			continue
		}

		answersJSON, _ := json.Marshal(responses[i])

		newResponse := questionnaire.QuestionnaireResponse{
			StudentProfileID:           studentProfile.ID,
			QuestionnaireID:            questionnaires[i].ID,
			Answers:                    string(answersJSON),
			SubmittedAt:                time.Now().AddDate(0, 0, -randomInt(30)),
			ProcessedAt:                timePtr(time.Now()),
			TotalScore:                 intPtr(85 + randomInt(15)),
			AIAnalysis:                 stringPtr("Berdasarkan jawaban, kandidat menunjukkan ketertarikan tinggi pada teknologi dan pemecahan masalah"),
			AIRecommendations:          stringPtr("Disarankan untuk fokus pada pengembangan skill programming dan project-based learning"),
			AIModelVersion:             stringPtr("gemini-1.5-pro"),
			RecommendedProfilingRoleID: &targetRoles[i].ID,
		}

		if err := db.Create(&newResponse).Error; err != nil {
			return fmt.Errorf("failed to create questionnaire response: %v", err)
		}

		log.Printf("Questionnaire response berhasil dibuat untuk student %s", studentProfile.Fullname)
	}

	return nil
}

// SeedStudentRoadmapProgress seeds roadmap progress for students
func SeedStudentRoadmapProgress(db *gorm.DB) error {
	log.Println("Seeding student roadmap progress...")

	// Get all students and roadmaps and teachers
	var studentProfiles []user.StudentProfile
	if err := db.Find(&studentProfiles).Error; err != nil {
		return fmt.Errorf("student profiles not found: %v", err)
	}

	var featureRoadmaps []roadmap.FeatureRoadmap
	if err := db.Preload("Steps").Find(&featureRoadmaps).Error; err != nil {
		return fmt.Errorf("roadmaps not found: %v", err)
	}

	var teacherProfiles []user.TeacherProfile
	if err := db.Find(&teacherProfiles).Error; err != nil {
		return fmt.Errorf("teacher profiles not found: %v", err)
	}

	if len(featureRoadmaps) == 0 {
		log.Println("No roadmaps found, skipping student roadmap progress seeding")
		return nil
	}

	// Create diverse scenarios for different students
	scenarios := []struct {
		Description      string
		CompletedSteps   int
		HasPendingSteps  bool
		HasRejectedSteps bool
		ProgressPattern  string // "linear", "mixed", "advanced"
	}{
		{"Beginner student - just started", 0, false, false, "linear"},
		{"Active student - making progress", 1, true, false, "linear"},
		{"Progressing student - some submissions pending", 2, true, false, "linear"},
		{"Advanced student - mixed progress with some rejections", 3, true, true, "mixed"},
		{"Struggling student - multiple rejections", 1, true, true, "mixed"},
		{"High achiever - consistent progress", 4, false, false, "linear"},
		{"Inconsistent student - sporadic submissions", 2, true, true, "mixed"},
		{"Recent starter - first submission pending", 1, true, false, "linear"},
		{"Fast learner - rapid progression", 3, true, false, "advanced"},
		{"Methodical student - steady progress", 2, false, false, "linear"},
	}

	for i, studentProfile := range studentProfiles {
		// Assign roadmap based on student index (cycling through available roadmaps)
		roadmapIndex := i % len(featureRoadmaps)
		roadmapData := featureRoadmaps[roadmapIndex]

		// Use scenario based on student index
		scenario := scenarios[i%len(scenarios)]

		var existingProgress roadmap.StudentRoadmapProgress
		err := db.Where("student_profile_id = ? AND roadmap_id = ?",
			studentProfile.ID, roadmapData.ID).First(&existingProgress).Error
		if err == nil {
			log.Printf("Roadmap progress sudah ada untuk student %s, melewati...", studentProfile.Fullname)
			continue
		}

		// Determine completed steps based on scenario and available steps
		maxSteps := len(roadmapData.Steps)
		completedSteps := scenario.CompletedSteps
		if completedSteps > maxSteps {
			completedSteps = maxSteps
		}

		progressPercent := float64(completedSteps) / float64(maxSteps) * 100

		newProgress := roadmap.StudentRoadmapProgress{
			RoadmapID:        roadmapData.ID,
			StudentProfileID: studentProfile.ID,
			TotalSteps:       maxSteps,
			CompletedSteps:   completedSteps,
			ProgressPercent:  progressPercent,
			StartedAt:        timePtr(time.Now().AddDate(0, 0, -randomInt(90))),
			LastActivityAt:   timePtr(time.Now().AddDate(0, 0, -randomInt(7))),
		}

		if completedSteps == maxSteps {
			newProgress.CompletedAt = timePtr(time.Now().AddDate(0, 0, -randomInt(30)))
		}

		if err := db.Create(&newProgress).Error; err != nil {
			return fmt.Errorf("failed to create roadmap progress: %v", err)
		}

		// Create step progress with diverse statuses
		for j, step := range roadmapData.Steps {
			status := roadmap.RoadmapProgressStatusLocked
			var startedAt, submittedAt, completedAt *time.Time
			var evidenceLink, submissionNotes *string
			var validatedByTeacherID *uuid.UUID
			var validationNotes *string
			var validationScore *int

			if j < completedSteps {
				// Completed steps
				status = roadmap.RoadmapProgressStatusApproved
				startedAt = timePtr(time.Now().AddDate(0, 0, -(randomInt(60) + 20)))
				submittedAt = timePtr(time.Now().AddDate(0, 0, -(randomInt(45) + 15)))
				completedAt = timePtr(time.Now().AddDate(0, 0, -(randomInt(30) + 10)))
				evidenceLink = stringPtr(fmt.Sprintf("https://github.com/%s/project-step-%d",
					strings.ToLower(strings.ReplaceAll(studentProfile.Fullname, " ", "")), j+1))
				submissionNotes = stringPtr(getRandomSubmissionNote())

				if len(teacherProfiles) > 0 {
					validatedByTeacherID = &teacherProfiles[randomInt(len(teacherProfiles))].ID
					validationNotes = stringPtr(getRandomValidationNote(true))
					validationScore = intPtr(75 + randomInt(25)) // 75-100 for approved
				}
			} else if j == completedSteps && scenario.HasPendingSteps && completedSteps < maxSteps {
				// Pending submission
				status = roadmap.RoadmapProgressStatusSubmitted
				startedAt = timePtr(time.Now().AddDate(0, 0, -randomInt(20)))
				submittedAt = timePtr(time.Now().AddDate(0, 0, -randomInt(5)))
				evidenceLink = stringPtr(fmt.Sprintf("https://github.com/%s/project-step-%d",
					strings.ToLower(strings.ReplaceAll(studentProfile.Fullname, " ", "")), j+1))
				submissionNotes = stringPtr(getRandomSubmissionNote())
			} else if j == completedSteps && scenario.HasRejectedSteps && completedSteps < maxSteps {
				// Rejected submission
				status = roadmap.RoadmapProgressStatusRejected
				startedAt = timePtr(time.Now().AddDate(0, 0, -(randomInt(25) + 10)))
				submittedAt = timePtr(time.Now().AddDate(0, 0, -(randomInt(15) + 5)))
				evidenceLink = stringPtr(fmt.Sprintf("https://github.com/%s/project-step-%d",
					strings.ToLower(strings.ReplaceAll(studentProfile.Fullname, " ", "")), j+1))
				submissionNotes = stringPtr(getRandomSubmissionNote())

				if len(teacherProfiles) > 0 {
					validatedByTeacherID = &teacherProfiles[randomInt(len(teacherProfiles))].ID
					validationNotes = stringPtr(getRandomValidationNote(false))
					validationScore = intPtr(40 + randomInt(35)) // 40-75 for rejected
				}
			} else if j == completedSteps+1 && scenario.ProgressPattern == "advanced" && completedSteps < maxSteps-1 {
				// In progress for advanced students
				status = roadmap.RoadmapProgressStatusInProgress
				startedAt = timePtr(time.Now().AddDate(0, 0, -randomInt(10)))
			} else if j <= completedSteps+1 && scenario.ProgressPattern == "mixed" {
				// Unlocked for mixed pattern students
				status = roadmap.RoadmapProgressStatusUnlocked
			}

			stepProgress := roadmap.StudentStepProgress{
				StudentRoadmapProgressID: newProgress.ID,
				RoadmapStepID:            step.ID,
				Status:                   status,
				EvidenceLink:             evidenceLink,
				EvidenceType:             stringPtr("url"),
				SubmissionNotes:          submissionNotes,
				ValidatedByTeacherID:     validatedByTeacherID,
				ValidationNotes:          validationNotes,
				ValidationScore:          validationScore,
				StartedAt:                startedAt,
				SubmittedAt:              submittedAt,
				CompletedAt:              completedAt,
			}

			if err := db.Create(&stepProgress).Error; err != nil {
				log.Printf("Warning: failed to create step progress: %v", err)
			}
		}

		log.Printf("Roadmap progress berhasil dibuat untuk student %s (%s) dengan %d/%d steps completed - %s",
			studentProfile.Fullname, roadmapData.RoadmapName, completedSteps, maxSteps, scenario.Description)
	}

	return nil
}

// SeedChallenges seeds sample challenges
func SeedChallenges(db *gorm.DB) error {
	log.Println("Seeding challenges...")

	// Get admin and teacher
	var admin user.User
	if err := db.Where("role = ?", user.RoleAdmin).First(&admin).Error; err != nil {
		return fmt.Errorf("admin user not found: %v", err)
	}

	var teacherProfiles []user.TeacherProfile
	if err := db.Find(&teacherProfiles).Error; err != nil {
		return fmt.Errorf("teacher profiles not found: %v", err)
	}

	challenges := []struct {
		Title           string
		Description     string
		Deadline        time.Time
		Prize           string
		MaxParticipants int
		CreatedByAdmin  bool
	}{
		{
			Title:           "Hackathon AICademy 2025",
			Description:     "Kompetisi pengembangan aplikasi dengan tema AI untuk pendidikan. Peserta diminta untuk membuat solusi inovatif menggunakan teknologi AI.",
			Deadline:        time.Now().AddDate(0, 2, 0),
			Prize:           "Rp 10,000,000 + Laptop Gaming",
			MaxParticipants: 30,
			CreatedByAdmin:  true,
		},
		{
			Title:           "Web Development Challenge",
			Description:     "Buat website e-learning interaktif dengan fitur real-time collaboration dan gamification elements.",
			Deadline:        time.Now().AddDate(0, 1, 15),
			Prize:           "Rp 5,000,000 + Sertifikat",
			MaxParticipants: 20,
			CreatedByAdmin:  false,
		},
		{
			Title:           "Mobile App Innovation Contest",
			Description:     "Develop a mobile application that solves real-world problems for students and educators.",
			Deadline:        time.Now().AddDate(0, 3, 0),
			Prize:           "Rp 7,500,000 + Internship Opportunity",
			MaxParticipants: 25,
			CreatedByAdmin:  true,
		},
	}

	for i, challengeData := range challenges {
		var existingChallenge challenge.Challenge
		err := db.Where("title = ?", challengeData.Title).First(&existingChallenge).Error
		if err == nil {
			log.Printf("Challenge '%s' sudah ada, melewati...", challengeData.Title)
			continue
		}

		newChallenge := challenge.Challenge{
			Title:               challengeData.Title,
			Description:         challengeData.Description,
			Deadline:            challengeData.Deadline,
			Prize:               &challengeData.Prize,
			MaxParticipants:     challengeData.MaxParticipants,
			CurrentParticipants: randomInt(challengeData.MaxParticipants/2) + 3, // Random participants
		}

		if challengeData.CreatedByAdmin {
			newChallenge.CreatedByAdminID = &admin.ID
		} else if len(teacherProfiles) > 0 {
			teacherID := teacherProfiles[i%len(teacherProfiles)].ID
			newChallenge.CreatedByTeacherID = &teacherID
		}

		if err := db.Create(&newChallenge).Error; err != nil {
			return fmt.Errorf("failed to create challenge %s: %v", challengeData.Title, err)
		}

		log.Printf("Challenge '%s' berhasil dibuat", challengeData.Title)
	}

	return nil
}

// SeedInternships seeds sample internship opportunities
func SeedInternships(db *gorm.DB) error {
	log.Println("Seeding internships...")

	// Get company profiles
	var companyProfiles []user.CompanyProfile
	if err := db.Find(&companyProfiles).Error; err != nil {
		return fmt.Errorf("company profiles not found: %v", err)
	}

	if len(companyProfiles) == 0 {
		log.Println("No company profiles found, skipping internship seeding")
		return nil
	}

	internships := []struct {
		Title       string
		Description string
		Type        pkl.InternshipType
		Deadline    *time.Time
	}{
		{
			Title:       "Backend Developer Intern",
			Description: "Bergabung dengan tim development untuk mengembangkan sistem backend yang scalable menggunakan Go dan PostgreSQL.",
			Type:        pkl.InternshipTypePKL,
			Deadline:    timePtr(time.Now().AddDate(0, 1, 0)),
		},
		{
			Title:       "Frontend Developer Intern",
			Description: "Kesempatan untuk belajar dan berkontribusi dalam pengembangan aplikasi web modern menggunakan React dan TypeScript.",
			Type:        pkl.InternshipTypePKL,
			Deadline:    timePtr(time.Now().AddDate(0, 0, 20)),
		},
		{
			Title:       "Data Analyst Intern",
			Description: "Analisis data bisnis dan pembuatan dashboard untuk mendukung decision making menggunakan Python dan SQL.",
			Type:        pkl.InternshipTypeJob,
			Deadline:    timePtr(time.Now().AddDate(0, 2, 0)),
		},
		{
			Title:       "UI/UX Designer Intern",
			Description: "Desain user interface dan user experience untuk aplikasi mobile dan web dengan tools modern seperti Figma.",
			Type:        pkl.InternshipTypeFreelance,
			Deadline:    timePtr(time.Now().AddDate(0, 1, 15)),
		},
	}

	for i, internshipData := range internships {
		companyIndex := i % len(companyProfiles)

		var existingInternship pkl.Internship
		err := db.Where("title = ? AND company_profile_id = ?",
			internshipData.Title, companyProfiles[companyIndex].ID).First(&existingInternship).Error
		if err == nil {
			log.Printf("Internship '%s' sudah ada untuk company %s, melewati...",
				internshipData.Title, companyProfiles[companyIndex].CompanyName)
			continue
		}

		newInternship := pkl.Internship{
			CompanyProfileID: companyProfiles[companyIndex].ID,
			Title:            internshipData.Title,
			Description:      internshipData.Description,
			Type:             internshipData.Type,
			Deadline:         internshipData.Deadline,
		}

		if err := db.Create(&newInternship).Error; err != nil {
			return fmt.Errorf("failed to create internship %s: %v", internshipData.Title, err)
		}

		log.Printf("Internship '%s' berhasil dibuat untuk company %s",
			internshipData.Title, companyProfiles[companyIndex].CompanyName)
	}

	return nil
}

// SeedStudentCVs seeds sample CVs for students
func SeedStudentCVs(db *gorm.DB) error {
	log.Println("Seeding student CVs...")

	// Get student profiles with their projects and certifications
	var studentProfiles []user.StudentProfile
	if err := db.Limit(3).Find(&studentProfiles).Error; err != nil {
		return fmt.Errorf("student profiles not found: %v", err)
	}

	for i, studentProfile := range studentProfiles {
		var existingCV cv.CV
		err := db.Where("student_profile_id = ?", studentProfile.ID).First(&existingCV).Error
		if err == nil {
			log.Printf("CV sudah ada untuk student %s, melewati...", studentProfile.Fullname)
			continue
		}

		// Get student's projects using raw SQL to handle PostgreSQL arrays properly
		var studentProjects []struct {
			ID          uuid.UUID `json:"id"`
			ProjectName string    `json:"project_name"`
			Description string    `json:"description"`
			LinkURL     *string   `json:"link_url"`
			StartDate   time.Time `json:"start_date"`
			EndDate     time.Time `json:"end_date"`
			TechStack   string    `json:"tech_stack"` // Raw string from PostgreSQL array
		}
		db.Raw("SELECT id, project_name, description, link_url, start_date, end_date, array_to_string(tech_stack, ',') as tech_stack FROM projects WHERE owner_student_profile_id = ?", studentProfile.ID).Scan(&studentProjects)

		// Get student's certifications
		var studentCerts []user.Certification
		db.Where("student_profile_id = ?", studentProfile.ID).Find(&studentCerts)

		// Create CV content
		cvContent := cv.CVContent{
			PersonalInfo: cv.PersonalInfo{
				FullName: studentProfile.Fullname,
				Email:    "student" + fmt.Sprintf("%d", i+1) + "@aicademy.com",
				Phone:    "+62812345678" + fmt.Sprintf("%d", i),
				Location: "Jakarta, Indonesia",
				LinkedIn: "linkedin.com/in/" + strings.ToLower(strings.ReplaceAll(studentProfile.Fullname, " ", "")),
				GitHub:   "github.com/" + strings.ToLower(strings.ReplaceAll(studentProfile.Fullname, " ", "")),
			},
			Summary: "Fresh graduate dengan passion di bidang teknologi dan pengembangan software. Berpengalaman dalam project development dan selalu eager to learn teknologi baru.",
			Skills: []cv.CVSkill{
				{Name: "JavaScript", Level: "Advanced", Category: "Programming"},
				{Name: "React", Level: "Intermediate", Category: "Framework"},
				{Name: "Node.js", Level: "Intermediate", Category: "Backend"},
				{Name: "PostgreSQL", Level: "Intermediate", Category: "Database"},
				{Name: "Git", Level: "Advanced", Category: "Tools"},
			},
			Education: cv.CVEducation{
				School:  "SMK Telkom Jakarta",
				Degree:  "Rekayasa Perangkat Lunak",
				Major:   "Rekayasa Perangkat Lunak",
				EndYear: intPtr(time.Now().Year()),
				GPA:     "3.75",
			},
			Keywords: []string{"JavaScript", "React", "Node.js", "Full Stack", "Web Development"},
		}

		// Add projects to CV
		for _, proj := range studentProjects {
			// Convert comma-separated string back to slice
			var techStack []string
			if proj.TechStack != "" {
				techStack = strings.Split(proj.TechStack, ",")
				// Trim spaces from each element
				for i, tech := range techStack {
					techStack[i] = strings.TrimSpace(tech)
				}
			}

			cvProject := cv.CVProject{
				Name:         proj.ProjectName,
				Description:  proj.Description,
				Technologies: techStack,
				StartDate:    proj.StartDate,
				EndDate:      &proj.EndDate,
				Role:         "Developer",
				Highlights:   []string{"Implemented core features", "Collaborated with team"},
			}
			if proj.LinkURL != nil {
				cvProject.URL = *proj.LinkURL
			}
			cvContent.Projects = append(cvContent.Projects, cvProject)
		}

		// Add certifications to CV
		for _, cert := range studentCerts {
			cvCert := cv.CVCertification{
				Name:                cert.Name,
				IssuingOrganization: cert.IssuingOrganization,
				IssueDate:           cert.IssueDate,
			}
			if cert.CredentialURL != nil {
				cvCert.CredentialURL = *cert.CredentialURL
			}
			if cert.ExpirationDate != nil {
				cvCert.ExpirationDate = cert.ExpirationDate
			}
			cvContent.Certifications = append(cvContent.Certifications, cvCert)
		}

		newCV := cv.CV{
			StudentProfileID: studentProfile.ID,
			Title:            studentProfile.Fullname + " - Software Developer",
			Status:           cv.CVStatusPublished,
			Content:          cvContent,
			IsPublic:         true,
			GeneratedAt:      time.Now(),
			PublishedAt:      timePtr(time.Now()),
		}

		if err := db.Create(&newCV).Error; err != nil {
			return fmt.Errorf("failed to create CV for %s: %v", studentProfile.Fullname, err)
		}

		log.Printf("CV berhasil dibuat untuk student %s", studentProfile.Fullname)
	}

	return nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func intPtr(i int) *int {
	return &i
}

func randomInt(n int) int {
	return int(time.Now().UnixNano()) % n
}

func SeedTeams(db *gorm.DB) error {
	log.Println("Seeding teams...")

	var studentProfiles []user.StudentProfile
	if err := db.Limit(9).Find(&studentProfiles).Error; err != nil {
		return fmt.Errorf("student profiles not found: %v", err)
	}

	if len(studentProfiles) < 6 {
		log.Println("Not enough students for team creation")
		return nil
	}

	var targetRoles []project.TargetRole
	if err := db.Limit(3).Find(&targetRoles).Error; err != nil {
		return fmt.Errorf("target roles not found: %v", err)
	}

	teams := []struct {
		TeamName   string
		About      string
		LeaderIdx  int
		MemberIdxs []int
	}{
		{
			TeamName:   "Code Warriors",
			About:      "Tim yang berfokus pada pengembangan aplikasi web dengan teknologi modern",
			LeaderIdx:  0,
			MemberIdxs: []int{0, 1, 2},
		},
		{
			TeamName:   "Data Miners",
			About:      "Tim yang berspesialisasi dalam analisis data dan machine learning",
			LeaderIdx:  3,
			MemberIdxs: []int{3, 4, 5},
		},
		{
			TeamName:   "Tech Innovators",
			About:      "Tim yang menciptakan solusi teknologi inovatif untuk berbagai masalah",
			LeaderIdx:  6,
			MemberIdxs: []int{6, 7, 8},
		},
	}

	for _, teamData := range teams {
		if teamData.LeaderIdx >= len(studentProfiles) {
			continue
		}

		var existingTeam challenge.Team
		err := db.Where("team_name = ?", teamData.TeamName).First(&existingTeam).Error
		if err == nil {
			log.Printf("Team '%s' sudah ada, melewati...", teamData.TeamName)
			continue
		}

		newTeam := challenge.Team{
			TeamName:                  teamData.TeamName,
			About:                     &teamData.About,
			CreatedByStudentProfileID: studentProfiles[teamData.LeaderIdx].ID,
		}

		if err := db.Create(&newTeam).Error; err != nil {
			return fmt.Errorf("failed to create team %s: %v", teamData.TeamName, err)
		}

		for i, memberIdx := range teamData.MemberIdxs {
			if memberIdx >= len(studentProfiles) {
				continue
			}

			role := "Developer"
			if i == 0 {
				role = "Team Leader"
			}

			member := challenge.TeamMember{
				TeamID:           newTeam.ID,
				StudentProfileID: studentProfiles[memberIdx].ID,
				MemberRole:       &role,
				ProfilingRoleID:  &targetRoles[i%len(targetRoles)].ID,
				JoinedAt:         time.Now().AddDate(0, 0, -randomInt(30)),
			}

			if err := db.Create(&member).Error; err != nil {
				return fmt.Errorf("failed to create team member: %v", err)
			}
		}

		log.Printf("Team '%s' berhasil dibuat dengan %d members", teamData.TeamName, len(teamData.MemberIdxs))
	}

	return nil
}

func SeedChallengeSubmissions(db *gorm.DB) error {
	log.Println("Seeding challenge submissions...")

	var challenges []challenge.Challenge
	if err := db.Find(&challenges).Error; err != nil {
		return fmt.Errorf("challenges not found: %v", err)
	}

	var teams []challenge.Team
	if err := db.Find(&teams).Error; err != nil {
		return fmt.Errorf("teams not found: %v", err)
	}

	var studentProfiles []user.StudentProfile
	if err := db.Limit(3).Find(&studentProfiles).Error; err != nil {
		return fmt.Errorf("student profiles not found: %v", err)
	}

	submissions := []struct {
		Title    string
		ImageURL string
		RepoURL  string
		DocsURL  string
		Points   int
	}{
		{
			Title:    "AI-Powered Learning Platform",
			ImageURL: "https://example.com/images/ai-learning-platform.jpg",
			RepoURL:  "https://github.com/team/ai-learning-platform",
			DocsURL:  "https://docs.example.com/ai-learning-platform",
			Points:   95,
		},
		{
			Title:    "Smart Campus Management System",
			ImageURL: "https://example.com/images/campus-system.jpg",
			RepoURL:  "https://github.com/team/campus-management",
			DocsURL:  "https://docs.example.com/campus-management",
			Points:   88,
		},
		{
			Title:    "Student Collaboration Hub",
			ImageURL: "https://example.com/images/collaboration-hub.jpg",
			RepoURL:  "https://github.com/team/collaboration-hub",
			DocsURL:  "https://docs.example.com/collaboration-hub",
			Points:   82,
		},
	}

	for i, submissionData := range submissions {
		if i >= len(challenges) {
			break
		}

		var existingSubmission challenge.Submission
		err := db.Where("title = ? AND challenge_id = ?", submissionData.Title, challenges[i].ID).First(&existingSubmission).Error
		if err == nil {
			log.Printf("Submission '%s' sudah ada, melewati...", submissionData.Title)
			continue
		}

		newSubmission := challenge.Submission{
			ChallengeID: challenges[i].ID,
			Title:       submissionData.Title,
			ImageURL:    &submissionData.ImageURL,
			RepoURL:     &submissionData.RepoURL,
			DocsURL:     &submissionData.DocsURL,
			SubmittedAt: time.Now().AddDate(0, 0, -randomInt(10)),
			Points:      &submissionData.Points,
		}

		if i < len(teams) {
			newSubmission.TeamID = &teams[i].ID
		} else if i < len(studentProfiles) {
			newSubmission.StudentProfileID = &studentProfiles[i].ID
		}

		if err := db.Create(&newSubmission).Error; err != nil {
			return fmt.Errorf("failed to create submission %s: %v", submissionData.Title, err)
		}

		log.Printf("Submission '%s' berhasil dibuat", submissionData.Title)
	}

	return nil
}

func SeedChallengeJudges(db *gorm.DB) error {
	log.Println("Seeding challenge judges...")

	var challenges []challenge.Challenge
	if err := db.Find(&challenges).Error; err != nil {
		return fmt.Errorf("challenges not found: %v", err)
	}

	var teacherProfiles []user.TeacherProfile
	if err := db.Find(&teacherProfiles).Error; err != nil {
		return fmt.Errorf("teacher profiles not found: %v", err)
	}

	if len(teacherProfiles) == 0 {
		log.Println("No teacher profiles found, skipping challenge judges seeding")
		return nil
	}

	for i, challengeData := range challenges {
		teacherIdx := i % len(teacherProfiles)

		var existingJudge challenge.ChallengeJudge
		err := db.Where("challenge_id = ? AND teacher_profile_id = ?", challengeData.ID, teacherProfiles[teacherIdx].ID).First(&existingJudge).Error
		if err == nil {
			log.Printf("Judge untuk challenge '%s' sudah ada, melewati...", challengeData.Title)
			continue
		}

		newJudge := challenge.ChallengeJudge{
			ChallengeID:      challengeData.ID,
			TeacherProfileID: teacherProfiles[teacherIdx].ID,
		}

		if err := db.Create(&newJudge).Error; err != nil {
			return fmt.Errorf("failed to create challenge judge: %v", err)
		}

		log.Printf("Judge berhasil ditambahkan untuk challenge '%s'", challengeData.Title)
	}

	return nil
}

func SeedInternshipApplications(db *gorm.DB) error {
	log.Println("Seeding internship applications...")

	var internships []pkl.Internship
	if err := db.Limit(3).Find(&internships).Error; err != nil {
		return fmt.Errorf("internships not found: %v", err)
	}

	var studentProfiles []user.StudentProfile
	if err := db.Limit(5).Find(&studentProfiles).Error; err != nil {
		return fmt.Errorf("student profiles not found: %v", err)
	}

	var alumniProfiles []user.AlumniProfile
	if err := db.Limit(3).Find(&alumniProfiles).Error; err != nil {
		return fmt.Errorf("alumni profiles not found: %v", err)
	}

	var teacherProfiles []user.TeacherProfile
	if err := db.Find(&teacherProfiles).Error; err != nil {
		return fmt.Errorf("teacher profiles not found: %v", err)
	}

	applicationCount := 0
	for i, internship := range internships {
		for j := 0; j < 2 && j < len(studentProfiles); j++ {
			var existingApp pkl.InternshipApplication
			err := db.Where("internship_id = ? AND student_profile_id = ?", internship.ID, studentProfiles[j].ID).First(&existingApp).Error
			if err == nil {
				continue
			}

			status := pkl.ApplicationStatusPending
			var reviewedAt *time.Time
			var approvedByUserID *uuid.UUID
			var approvedByRole *string

			if applicationCount%3 == 0 {
				status = pkl.ApplicationStatusApproved
				reviewedAt = timePtr(time.Now().AddDate(0, 0, -randomInt(5)))
				if len(teacherProfiles) > 0 {
					approvedByUserID = &teacherProfiles[0].UserID
					role := "teacher"
					approvedByRole = &role
				}
			} else if applicationCount%3 == 1 {
				status = pkl.ApplicationStatusRejected
				reviewedAt = timePtr(time.Now().AddDate(0, 0, -randomInt(5)))
				if len(teacherProfiles) > 0 {
					approvedByUserID = &teacherProfiles[0].UserID
					role := "teacher"
					approvedByRole = &role
				}
			}

			newApp := pkl.InternshipApplication{
				InternshipID:     internship.ID,
				StudentProfileID: &studentProfiles[j].ID,
				Status:           status,
				AppliedAt:        time.Now().AddDate(0, 0, -randomInt(15)),
				ReviewedAt:       reviewedAt,
				ApprovedByUserID: approvedByUserID,
				ApprovedByRole:   approvedByRole,
			}

			if err := db.Create(&newApp).Error; err != nil {
				log.Printf("Warning: failed to create internship application: %v", err)
			} else {
				log.Printf("Internship application berhasil dibuat untuk student %s", studentProfiles[j].Fullname)
			}
			applicationCount++
		}

		if i < len(alumniProfiles) {
			var existingApp pkl.InternshipApplication
			err := db.Where("internship_id = ? AND alumni_profile_id = ?", internship.ID, alumniProfiles[i].ID).First(&existingApp).Error
			if err != nil {
				newApp := pkl.InternshipApplication{
					InternshipID:    internship.ID,
					AlumniProfileID: &alumniProfiles[i].ID,
					Status:          pkl.ApplicationStatusPending,
					AppliedAt:       time.Now().AddDate(0, 0, -randomInt(15)),
				}

				if err := db.Create(&newApp).Error; err != nil {
					log.Printf("Warning: failed to create alumni internship application: %v", err)
				} else {
					log.Printf("Internship application berhasil dibuat untuk alumni %s", alumniProfiles[i].Fullname)
				}
			}
		}
	}

	return nil
}

func SeedInternshipReviews(db *gorm.DB) error {
	log.Println("Seeding internship reviews...")

	var internships []pkl.Internship
	if err := db.Find(&internships).Error; err != nil {
		return fmt.Errorf("internships not found: %v", err)
	}

	var studentProfiles []user.StudentProfile
	if err := db.Limit(3).Find(&studentProfiles).Error; err != nil {
		return fmt.Errorf("student profiles not found: %v", err)
	}

	reviews := []struct {
		Rating      int
		Testimonial string
	}{
		{
			Rating:      5,
			Testimonial: "Pengalaman magang yang luar biasa! Mendapat bimbingan yang sangat baik dari mentor dan team. Banyak belajar tentang development process dan best practices dalam industri.",
		},
		{
			Rating:      4,
			Testimonial: "Magang yang sangat bermanfaat, environment kerja yang supportive dan project yang challenging. Sangat membantu untuk mempersiapkan karir di bidang teknologi.",
		},
		{
			Rating:      5,
			Testimonial: "Amazing internship experience! Tim yang solid, mentor yang berpengalaman, dan project yang real-world. Definitely recommended untuk yang ingin belajar lebih dalam tentang software development.",
		},
	}

	for i, reviewData := range reviews {
		if i >= len(internships) || i >= len(studentProfiles) {
			break
		}

		var existingReview pkl.InternshipReview
		err := db.Where("internship_id = ? AND student_profile_id = ?", internships[i].ID, studentProfiles[i].ID).First(&existingReview).Error
		if err == nil {
			log.Printf("Review sudah ada untuk internship dan student, melewati...")
			continue
		}

		newReview := pkl.InternshipReview{
			InternshipID:     internships[i].ID,
			StudentProfileID: studentProfiles[i].ID,
			Rating:           reviewData.Rating,
			Testimonial:      reviewData.Testimonial,
		}

		if err := db.Create(&newReview).Error; err != nil {
			return fmt.Errorf("failed to create internship review: %v", err)
		}

		log.Printf("Internship review berhasil dibuat untuk student %s", studentProfiles[i].Fullname)
	}

	return nil
}

// Helper functions for generating random content
func getRandomSubmissionNote() string {
	notes := []string{
		"Project berhasil diselesaikan sesuai dengan requirements yang diberikan. Semua fitur telah diimplementasikan dengan baik.",
		"Implementasi sudah sesuai dengan spesifikasi teknis. Testing telah dilakukan dan semua test case passed.",
		"Aplikasi berjalan dengan lancar dan memenuhi kriteria penilaian. User interface responsif dan user-friendly.",
		"Source code clean dan well-documented. Best practices sudah diterapkan dalam development process.",
		"Deployment berhasil dan aplikasi dapat diakses dengan baik. Performance optimization sudah dilakukan.",
		"Fitur tambahan berhasil diimplementasikan untuk meningkatkan user experience.",
		"Integration dengan external API berjalan dengan baik. Error handling sudah diterapkan dengan proper.",
		"Database design optimal dan normalized. Query performance sudah dioptimasi.",
		"Security measures telah diimplementasikan sesuai dengan standar keamanan.",
		"Code review internal sudah dilakukan dan feedback sudah diincorporate.",
	}
	return notes[randomInt(len(notes))]
}

func getRandomValidationNote(approved bool) string {
	if approved {
		notes := []string{
			"Excellent work! Project memenuhi semua kriteria penilaian dengan sangat baik.",
			"Great job! Implementation sangat solid dan code quality bagus.",
			"Outstanding! Creativity dan problem-solving skills terlihat jelas dalam project ini.",
			"Very good! Technical implementation sesuai dengan best practices.",
			"Impressive work! Attention to detail sangat baik dan hasil akhir memuaskan.",
			"Well done! Project menunjukkan pemahaman yang mendalam terhadap materi.",
			"Fantastic! Innovation dan technical skills yang ditunjukkan sangat baik.",
			"Superb execution! Code organization dan dokumentasi sangat rapi.",
			"Excellent problem-solving approach dan implementation yang efisien.",
			"Outstanding project! Menunjukkan kemampuan development yang matang.",
		}
		return notes[randomInt(len(notes))]
	} else {
		notes := []string{
			"Project needs improvement. Beberapa requirements belum terpenuhi dengan baik.",
			"Implementation perlu diperbaiki. Code quality masih bisa ditingkatkan.",
			"Submission tidak memenuhi kriteria minimum. Silakan revisi dan submit ulang.",
			"Functionality tidak sesuai dengan spesifikasi. Perlu review ulang requirements.",
			"Code structure perlu diperbaiki. Best practices belum diterapkan dengan konsisten.",
			"Testing tidak sufficient. Perlu menambahkan test cases yang lebih comprehensive.",
			"Documentation kurang lengkap. Silakan tambahkan penjelasan yang lebih detail.",
			"Performance issues ditemukan. Optimasi diperlukan sebelum submission final.",
			"Security vulnerabilities terdeteksi. Silakan perbaiki sebelum submit ulang.",
			"UI/UX perlu improvement. User experience masih bisa diperbaiki.",
		}
		return notes[randomInt(len(notes))]
	}
}
