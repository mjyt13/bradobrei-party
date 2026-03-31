package reports

import (
	"strings"
	"testing"
	"time"

	"bradobrei/backend/internal/models"
)

func TestRenderEmployeesHTML(t *testing.T) {
	renderer, err := NewRenderer(nil)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	doc := models.EmployeeRegistryReportDocument{
		Meta: models.ReportMeta{
			ReportCode:  "2.2.1",
			Title:       "Реестр персонала",
			GeneratedAt: now,
		},
		Rows: []models.EmployeeRegistryReportRow{
			{
				FullName:       "Иван Барбер",
				Role:           models.RoleAdvancedMaster,
				Phone:          "+79990001122",
				Specialization: "Fade",
				Salons:         []string{"Пермь, Ленина 10"},
				ExpectedSalary: 85000,
			},
		},
	}

	htmlBytes, err := renderer.RenderEmployeesHTML(doc)
	if err != nil {
		t.Fatalf("failed to render html: %v", err)
	}

	html := string(htmlBytes)
	if !strings.Contains(html, "Реестр персонала") {
		t.Fatal("rendered html does not contain report title")
	}
	if !strings.Contains(html, "Иван Барбер") {
		t.Fatal("rendered html does not contain employee row")
	}
	if !strings.Contains(html, "report.css") {
		t.Fatal("rendered html does not link stylesheet asset")
	}
}
