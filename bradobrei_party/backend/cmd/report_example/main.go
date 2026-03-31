package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/reports"
)

func main() {
	htmlOut := flag.String("html-out", filepath.Join("test_artifacts", "employees_report_example.html"), "path to rendered HTML output")
	pdfOut := flag.String("pdf-out", filepath.Join("test_artifacts", "employees_report_example.pdf"), "path to rendered PDF output")
	skipPDF := flag.Bool("skip-pdf", false, "render only HTML without sending it to Gotenberg")
	flag.Parse()

	client := reports.NewGotenbergClientFromEnv()
	renderer, err := reports.NewRenderer(client)
	if err != nil {
		log.Fatal(err)
	}

	doc := sampleEmployeesReport()

	htmlBytes, err := renderer.RenderEmployeesHTML(doc)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(*htmlOut), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*htmlOut, htmlBytes, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("HTML example saved to %s", *htmlOut)

	if *skipPDF {
		log.Println("Skipping PDF generation by request.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	pdfBytes, err := renderer.RenderEmployeesPDF(ctx, doc)
	if err != nil {
		log.Fatalf("failed to render PDF via Gotenberg: %v", err)
	}
	if err := os.WriteFile(*pdfOut, pdfBytes, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("PDF example saved to %s", *pdfOut)
}

func sampleEmployeesReport() models.EmployeeRegistryReportDocument {
	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)

	return models.EmployeeRegistryReportDocument{
		Meta: models.ReportMeta{
			ReportCode:   "2.2.1",
			Title:        "Реестр персонала",
			Subtitle:     "Пример печатной формы для разработки HTML/PDF-шаблона.",
			GeneratedAt:  time.Now(),
			PeriodFrom:   &from,
			PeriodTo:     &to,
			GeneratedBy:  "Локальная demo-команда",
			Organization: "Bradobrei Party",
		},
		Rows: []models.EmployeeRegistryReportRow{
			{
				FullName:       "Иван Барбер",
				Role:           models.RoleAdvancedMaster,
				Phone:          "+79990001122",
				Email:          "ivan.barber@example.com",
				Specialization: "Fade, beard styling",
				Salons:         []string{"Пермь, Ленина 10", "Пермь, Куйбышева 45"},
				ExpectedSalary: 85000,
			},
			{
				FullName:       "Мария HR",
				Role:           models.RoleHR,
				Phone:          "+79995554433",
				Email:          "maria.hr@example.com",
				Specialization: "Подбор и адаптация персонала",
				Salons:         []string{"Екатеринбург, Малышева 15"},
				ExpectedSalary: 72000,
			},
		},
	}
}
