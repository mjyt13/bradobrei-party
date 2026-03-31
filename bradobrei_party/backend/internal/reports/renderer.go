package reports

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"strings"
	"time"

	"bradobrei/backend/internal/models"
)

//go:embed templates/*.html templates/*.css
var templatesFS embed.FS

type Renderer struct {
	client  *GotenbergClient
	cssData []byte
}

func NewRenderer(client *GotenbergClient) (*Renderer, error) {
	cssData, err := templatesFS.ReadFile("templates/report.css")
	if err != nil {
		return nil, err
	}

	return &Renderer{
		client:  client,
		cssData: cssData,
	}, nil
}

func (r *Renderer) RenderEmployeesHTML(doc models.EmployeeRegistryReportDocument) ([]byte, error) {
	return r.render("employees.html", doc)
}

func (r *Renderer) RenderSalonActivityHTML(doc models.SalonActivityReportDocument) ([]byte, error) {
	return r.render("salon_activity.html", doc)
}

func (r *Renderer) RenderMasterActivityHTML(doc models.MasterActivityReportDocument) ([]byte, error) {
	return r.render("master_activity.html", doc)
}

func (r *Renderer) RenderReviewsHTML(doc models.ReviewsReportDocument) ([]byte, error) {
	return r.render("reviews.html", doc)
}

func (r *Renderer) RenderEmployeesPDF(ctx context.Context, doc models.EmployeeRegistryReportDocument) ([]byte, error) {
	htmlBytes, err := r.RenderEmployeesHTML(doc)
	if err != nil {
		return nil, err
	}
	return r.convertHTML(ctx, "employees-report", htmlBytes)
}

func (r *Renderer) RenderSalonActivityPDF(ctx context.Context, doc models.SalonActivityReportDocument) ([]byte, error) {
	htmlBytes, err := r.RenderSalonActivityHTML(doc)
	if err != nil {
		return nil, err
	}
	return r.convertHTML(ctx, "salon-activity-report", htmlBytes)
}

func (r *Renderer) RenderMasterActivityPDF(ctx context.Context, doc models.MasterActivityReportDocument) ([]byte, error) {
	htmlBytes, err := r.RenderMasterActivityHTML(doc)
	if err != nil {
		return nil, err
	}
	return r.convertHTML(ctx, "master-activity-report", htmlBytes)
}

func (r *Renderer) RenderReviewsPDF(ctx context.Context, doc models.ReviewsReportDocument) ([]byte, error) {
	htmlBytes, err := r.RenderReviewsHTML(doc)
	if err != nil {
		return nil, err
	}
	return r.convertHTML(ctx, "reviews-report", htmlBytes)
}

func (r *Renderer) render(templateName string, data any) ([]byte, error) {
	tpl, err := template.New("base.html").Funcs(template.FuncMap{
		"formatDate":     formatDate,
		"formatDateTime": formatDateTime,
		"formatMoney":    formatMoney,
		"join":           strings.Join,
		"safePeriod":     safePeriod,
	}).ParseFS(templatesFS, "templates/base.html", "templates/"+templateName)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "base", data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *Renderer) convertHTML(ctx context.Context, outputFilename string, htmlBytes []byte) ([]byte, error) {
	if r.client == nil {
		return nil, fmt.Errorf("gotenberg client is not configured")
	}

	return r.client.ConvertHTML(ctx, HTMLRequest{
		IndexHTML:      htmlBytes,
		OutputFilename: outputFilename,
		Assets: []HTMLAsset{
			{
				Filename:    "report.css",
				Content:     r.cssData,
				ContentType: "text/css; charset=utf-8",
			},
		},
		Options: HTMLConvertOptions{
			MarginTop:         "0.5in",
			MarginBottom:      "0.5in",
			MarginLeft:        "0.45in",
			MarginRight:       "0.45in",
			PreferCSSPageSize: true,
			WaitDelay:         "500ms",
		},
	})
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("02.01.2006")
}

func formatDateTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("02.01.2006 15:04")
}

func formatMoney(value float64) string {
	return fmt.Sprintf("%.2f руб.", value)
}

func safePeriod(from, to *time.Time) string {
	if from == nil && to == nil {
		return "За всё время"
	}
	if from != nil && to != nil {
		return fmt.Sprintf("%s - %s", formatDate(*from), formatDate(*to))
	}
	if from != nil {
		return fmt.Sprintf("С %s", formatDate(*from))
	}
	return fmt.Sprintf("По %s", formatDate(*to))
}
