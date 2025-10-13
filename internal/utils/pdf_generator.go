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
	Skills         []CVSkill         `json:"skills"`
	Projects       []CVProject       `json:"projects"`
	Certifications []CVCertification `json:"certifications"`
	Education      CVEducation       `json:"education"`
	Keywords       []string          `json:"keywords"`
}

type PersonalInfo struct {
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	Location  string `json:"location,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	GitHub    string `json:"github,omitempty"`
	Portfolio string `json:"portfolio,omitempty"`
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

func NewPDFGenerator() *PDFGenerator {
	pdf := gofpdf.New("P", "mm", "A4", "")
	return &PDFGenerator{pdf: pdf}
}

func (pg *PDFGenerator) GenerateATSCV(content *CVContent) (string, error) {
	pg.pdf.AddPage()
	pg.pdf.SetFont("Arial", "", 10)

	pg.generateHeader(&content.PersonalInfo)

	if content.Summary != "" {
		pg.addSectionHeader("PROFESSIONAL SUMMARY")
		pg.pdf.SetFont("Arial", "", 10)
		pg.pdf.MultiCell(0, 5, content.Summary, "", "", false)
		pg.pdf.Ln(5)
	}

	if len(content.Skills) > 0 {
		pg.generateSkills(content.Skills)
	}

	if len(content.Projects) > 0 {
		pg.generateExperience(content.Projects)
	}

	if len(content.Certifications) > 0 {
		pg.generateCertifications(content.Certifications)
	}

	pg.generateEducation(&content.Education)

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
	contactInfo := info.Email
	if info.Phone != "" {
		contactInfo += " | " + info.Phone
	}
	if info.Location != "" {
		contactInfo += " | " + info.Location
	}
	pg.pdf.CellFormat(0, 8, contactInfo, "", 1, "C", false, 0, "")

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

func (pg *PDFGenerator) generateExperience(projects []CVProject) {
	pg.addSectionHeader("PROFESSIONAL EXPERIENCE")

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
				pg.pdf.CellFormat(5, 5, "•", "", 0, "", false, 0, "")
				pg.pdf.MultiCell(0, 5, highlight, "", "", false)
			}
		}

		pg.pdf.Ln(4)
	}
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

		pg.pdf.CellFormat(5, 5, "•", "", 0, "", false, 0, "")
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
