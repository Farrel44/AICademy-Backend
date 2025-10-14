package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type PDFGenerator struct {
	pdf *gofpdf.Fpdf
}

type CVContent struct {
	PersonalInfo   PersonalInfo      `json:"personal_info"`
	Summary        string            `json:"summary"`
	Experiences    []CVExperience    `json:"experiences"`
	Projects       []CVProject       `json:"projects"`
	Skills         []CVSkill         `json:"skills"`
	Certifications []CVCertification `json:"certifications"`
	Education      CVEducation       `json:"education"`
	Languages      []CVLanguage      `json:"languages"`
	Keywords       []string          `json:"keywords"`
}

type PersonalInfo struct {
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
	PersonalEmail string `json:"personal_email"`
	Phone         string `json:"phone,omitempty"`
	Location      string `json:"location,omitempty"`
	LinkedIn      string `json:"linkedin,omitempty"`
	GitHub        string `json:"github,omitempty"`
	Portfolio     string `json:"portfolio,omitempty"`
}

type CVSkill struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Level    string `json:"level"`
}

type CVProject struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Role         string     `json:"role"`
	Description  string     `json:"description"`
	Technologies []string   `json:"technologies"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	URL          string     `json:"url,omitempty"`
	Highlights   []string   `json:"highlights"`
}

type CVCertification struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	IssuingOrganization string     `json:"issuing_organization"`
	IssueDate           time.Time  `json:"issue_date"`
	ExpirationDate      *time.Time `json:"expiration_date"`
	CredentialID        string     `json:"credential_id,omitempty"`
	CredentialURL       string     `json:"credential_url,omitempty"`
}

type CVEducation struct {
	School    string `json:"school"`
	Degree    string `json:"degree"`
	Major     string `json:"major"`
	StartYear int    `json:"start_year"`
	EndYear   int    `json:"end_year"`
	GPA       string `json:"gpa,omitempty"`
}

type CVExperience struct {
	ID               string     `json:"id"`
	CompanyName      string     `json:"company_name"`
	Position         string     `json:"position"`
	Department       string     `json:"department,omitempty"`
	EmploymentType   string     `json:"employment_type"`
	Location         string     `json:"location"`
	LocationType     string     `json:"location_type"`
	Description      string     `json:"description"`
	Responsibilities []string   `json:"responsibilities"`
	Achievements     []string   `json:"achievements"`
	Skills           []string   `json:"skills"`
	StartDate        time.Time  `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	IsCurrent        bool       `json:"is_current"`
}

type CVLanguage struct {
	Name      string `json:"name"`
	Level     string `json:"level"`
	Certified bool   `json:"certified"`
}

func NewPDFGenerator() *PDFGenerator {
	pdf := gofpdf.New("P", "mm", "A4", "")
	return &PDFGenerator{pdf: pdf}
}

func (pg *PDFGenerator) GenerateATSCV(content *CVContent) (string, error) {
	pg.pdf.AddPage()
	pg.pdf.SetFont("Arial", "", 10)

	// 1. Personal Info (Name + Contact/Location)
	pg.generateHeader(&content.PersonalInfo)

	// 2. Professional Summary
	if content.Summary != "" {
		pg.addSectionHeader("PROFESSIONAL SUMMARY")
		pg.pdf.SetFont("Arial", "", 10)
		pg.pdf.MultiCell(0, 5, content.Summary, "", "", false)
		pg.pdf.Ln(5)
	}

	// 3. Work Experiences (New section)
	if len(content.Experiences) > 0 {
		pg.generateWorkExperiences(content.Experiences)
	}

	// 4. Projects
	if len(content.Projects) > 0 {
		pg.generateProjects(content.Projects)
	}

	// 5. Skills
	if len(content.Skills) > 0 {
		pg.generateSkills(content.Skills)
	}

	// 6. Certifications
	if len(content.Certifications) > 0 {
		pg.generateCertifications(content.Certifications)
	}

	// 7. Education
	pg.generateEducation(&content.Education)

	// 8. Languages (New section)
	if len(content.Languages) > 0 {
		pg.generateLanguages(content.Languages)
	}

	filename := fmt.Sprintf("cv_%d.pdf", time.Now().Unix())
	outputPath := filepath.Join("temp", filename)

	if err := os.MkdirAll("temp", 0755); err != nil {
		return "", err
	}

	if err := pg.pdf.OutputFileAndClose(outputPath); err != nil {
		return "", err
	}

	return outputPath, nil
}

func (pg *PDFGenerator) generateHeader(info *PersonalInfo) {
	pg.pdf.SetFont("Arial", "B", 16)
	pg.pdf.CellFormat(0, 10, info.FullName, "", 1, "C", false, 0, "")

	pg.pdf.SetFont("Arial", "", 10)

	// Primary contact info: Email, Phone, Location
	contactInfo := info.Email
	if info.PersonalEmail != "" && info.PersonalEmail != info.Email {
		contactInfo += " | " + info.PersonalEmail
	}
	if info.Phone != "" {
		contactInfo += " | " + info.Phone
	}
	if info.Location != "" {
		contactInfo += " | " + info.Location
	}
	pg.pdf.CellFormat(0, 8, contactInfo, "", 1, "C", false, 0, "")

	// Professional links
	links := []string{}
	if info.LinkedIn != "" {
		links = append(links, info.LinkedIn)
	}
	if info.GitHub != "" {
		links = append(links, info.GitHub)
	}
	if info.Portfolio != "" {
		links = append(links, info.Portfolio)
	}

	if len(links) > 0 {
		pg.pdf.CellFormat(0, 6, strings.Join(links, " | "), "", 1, "C", false, 0, "")
	}

	pg.pdf.Ln(5)
}

func (pg *PDFGenerator) generateSkills(skills []CVSkill) {
	pg.addSectionHeader("TECHNICAL SKILLS")

	skillsByCategory := make(map[string][]string)
	for _, skill := range skills {
		category := skill.Category
		if category == "" {
			category = "Technical"
		}
		skillsByCategory[category] = append(skillsByCategory[category], skill.Name)
	}

	for category, skillList := range skillsByCategory {
		pg.pdf.SetFont("Arial", "B", 10)
		pg.pdf.Cell(40, 5, category+": ")
		pg.pdf.SetFont("Arial", "", 10)
		pg.pdf.MultiCell(0, 5, strings.Join(skillList, ", "), "", "", false)
		pg.pdf.Ln(2)
	}
	pg.pdf.Ln(3)
}

func (pg *PDFGenerator) generateProjects(projects []CVProject) {
	pg.addSectionHeader("PROJECTS")

	for _, project := range projects {
		pg.pdf.SetFont("Arial", "B", 11)
		projectTitle := project.Name
		if project.Role != "" {
			projectTitle += " - " + project.Role
		}
		pg.pdf.CellFormat(0, 6, projectTitle, "", 1, "", false, 0, "")

		pg.pdf.SetFont("Arial", "", 9)
		dateRange := project.StartDate.Format("Jan 2006")
		if project.EndDate != nil {
			dateRange += " - " + project.EndDate.Format("Jan 2006")
		} else {
			dateRange += " - Present"
		}
		pg.pdf.CellFormat(0, 4, dateRange, "", 1, "", false, 0, "")

		if project.Description != "" {
			pg.pdf.SetFont("Arial", "", 10)
			pg.pdf.MultiCell(0, 5, project.Description, "", "", false)
		}

		if len(project.Technologies) > 0 {
			pg.pdf.SetFont("Arial", "I", 9)
			techText := "Technologies: " + strings.Join(project.Technologies, ", ")
			pg.pdf.MultiCell(0, 4, techText, "", "", false)
		}

		if len(project.Highlights) > 0 {
			pg.pdf.SetFont("Arial", "", 10)
			for _, highlight := range project.Highlights {
				pg.pdf.CellFormat(5, 5, "-", "", 0, "", false, 0, "")
				pg.pdf.MultiCell(0, 5, highlight, "", "", false)
			}
		}

		pg.pdf.Ln(4)
	}
}

func (pg *PDFGenerator) generateWorkExperiences(experiences []CVExperience) {
	pg.addSectionHeader("PROFESSIONAL EXPERIENCE")

	for _, exp := range experiences {
		pg.pdf.SetFont("Arial", "B", 11)
		positionTitle := exp.Position + " - " + exp.CompanyName
		pg.pdf.CellFormat(0, 6, positionTitle, "", 1, "", false, 0, "")

		pg.pdf.SetFont("Arial", "", 9)
		locationInfo := exp.Location + " (" + exp.LocationType + ")"
		pg.pdf.CellFormat(0, 4, locationInfo, "", 1, "", false, 0, "")

		dateRange := exp.StartDate.Format("Jan 2006")
		if exp.IsCurrent {
			dateRange += " - Present"
		} else if exp.EndDate != nil {
			dateRange += " - " + exp.EndDate.Format("Jan 2006")
		}
		pg.pdf.CellFormat(0, 4, dateRange, "", 1, "", false, 0, "")

		if exp.Description != "" {
			pg.pdf.SetFont("Arial", "", 10)
			pg.pdf.MultiCell(0, 5, exp.Description, "", "", false)
		}

		if len(exp.Responsibilities) > 0 {
			pg.pdf.SetFont("Arial", "B", 10)
			pg.pdf.CellFormat(0, 5, "Key Responsibilities:", "", 1, "", false, 0, "")
			pg.pdf.SetFont("Arial", "", 10)
			for _, resp := range exp.Responsibilities {
				pg.pdf.CellFormat(5, 5, "-", "", 0, "", false, 0, "")
				pg.pdf.MultiCell(0, 5, resp, "", "", false)
			}
		}

		if len(exp.Achievements) > 0 {
			pg.pdf.SetFont("Arial", "B", 10)
			pg.pdf.CellFormat(0, 5, "Key Achievements:", "", 1, "", false, 0, "")
			pg.pdf.SetFont("Arial", "", 10)
			for _, achievement := range exp.Achievements {
				pg.pdf.CellFormat(5, 5, "-", "", 0, "", false, 0, "")
				pg.pdf.MultiCell(0, 5, achievement, "", "", false)
			}
		}

		if len(exp.Skills) > 0 {
			pg.pdf.SetFont("Arial", "I", 9)
			skillsText := "Skills Used: " + strings.Join(exp.Skills, ", ")
			pg.pdf.MultiCell(0, 4, skillsText, "", "", false)
		}

		pg.pdf.Ln(4)
	}
}

func (pg *PDFGenerator) generateLanguages(languages []CVLanguage) {
	pg.addSectionHeader("LANGUAGES")

	pg.pdf.SetFont("Arial", "", 10)
	for i, lang := range languages {
		langText := lang.Name + " - " + lang.Level
		if lang.Certified {
			langText += " (Certified)"
		}

		if i > 0 && i%2 == 0 {
			pg.pdf.Ln(5)
		}

		if i%2 == 0 {
			pg.pdf.CellFormat(95, 5, langText, "", 0, "", false, 0, "")
		} else {
			pg.pdf.CellFormat(95, 5, langText, "", 1, "", false, 0, "")
		}
	}

	// Add line break if last row wasn't complete
	if len(languages)%2 != 0 {
		pg.pdf.Ln(5)
	}

	pg.pdf.Ln(3)
}

func (pg *PDFGenerator) generateCertifications(certifications []CVCertification) {
	pg.addSectionHeader("CERTIFICATIONS")

	for _, cert := range certifications {
		pg.pdf.SetFont("Arial", "B", 10)
		certText := cert.Name + " - " + cert.IssuingOrganization
		pg.pdf.MultiCell(0, 5, certText, "", "", false)

		pg.pdf.SetFont("Arial", "", 9)
		dateText := cert.IssueDate.Format("Jan 2006")
		if cert.ExpirationDate != nil {
			dateText += " - " + cert.ExpirationDate.Format("Jan 2006")
		}

		pg.pdf.CellFormat(5, 5, "-", "", 0, "", false, 0, "")
		pg.pdf.MultiCell(0, 4, "Issued: "+dateText, "", "", false)
		pg.pdf.Ln(2)
	}
	pg.pdf.Ln(3)
}

func (pg *PDFGenerator) generateEducation(education *CVEducation) {
	pg.addSectionHeader("EDUCATION")

	pg.pdf.SetFont("Arial", "B", 11)
	pg.pdf.CellFormat(0, 6, education.School, "", 1, "", false, 0, "")

	pg.pdf.SetFont("Arial", "", 10)
	degreeText := education.Degree
	if education.Major != "" {
		degreeText += " in " + education.Major
	}

	dateText := fmt.Sprintf("%d - %d", education.StartYear, education.EndYear)
	pg.pdf.CellFormat(0, 5, degreeText+" ("+dateText+")", "", 1, "", false, 0, "")

	if education.GPA != "" {
		pg.pdf.CellFormat(0, 5, "GPA: "+education.GPA, "", 1, "", false, 0, "")
	}

	pg.pdf.Ln(5)
}

func (pg *PDFGenerator) addSectionHeader(title string) {
	pg.pdf.SetFont("Arial", "B", 12)
	pg.pdf.CellFormat(0, 8, title, "", 1, "", false, 0, "")
	pg.pdf.SetDrawColor(0, 0, 0)
	pg.pdf.Line(10, pg.pdf.GetY(), 200, pg.pdf.GetY())
	pg.pdf.Ln(5)
}
