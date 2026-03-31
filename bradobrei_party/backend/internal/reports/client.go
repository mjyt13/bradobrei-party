package reports

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const defaultGotenbergURL = "http://localhost:3000"

type HTMLAsset struct {
	Filename    string
	Content     []byte
	ContentType string
}

type HTMLConvertOptions struct {
	PaperWidth        string
	PaperHeight       string
	MarginTop         string
	MarginBottom      string
	MarginLeft        string
	MarginRight       string
	Landscape         bool
	PreferCSSPageSize bool
	WaitDelay         string
}

type HTMLRequest struct {
	IndexHTML      []byte
	Assets         []HTMLAsset
	OutputFilename string
	TraceID        string
	Options        HTMLConvertOptions
}

type GotenbergClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewGotenbergClient(baseURL string) *GotenbergClient {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = defaultGotenbergURL
	}

	return &GotenbergClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func NewGotenbergClientFromEnv() *GotenbergClient {
	return NewGotenbergClient(os.Getenv("GOTENBERG_URL"))
}

func (c *GotenbergClient) ConvertHTML(ctx context.Context, req HTMLRequest) ([]byte, error) {
	if len(req.IndexHTML) == 0 {
		return nil, fmt.Errorf("index.html is required")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writeMultipartFile(writer, "files", HTMLAsset{
		Filename:    "index.html",
		Content:     req.IndexHTML,
		ContentType: "text/html; charset=utf-8",
	}); err != nil {
		return nil, err
	}

	for _, asset := range req.Assets {
		if err := writeMultipartFile(writer, "files", asset); err != nil {
			return nil, err
		}
	}

	writeFormField(writer, "paperWidth", req.Options.PaperWidth)
	writeFormField(writer, "paperHeight", req.Options.PaperHeight)
	writeFormField(writer, "marginTop", req.Options.MarginTop)
	writeFormField(writer, "marginBottom", req.Options.MarginBottom)
	writeFormField(writer, "marginLeft", req.Options.MarginLeft)
	writeFormField(writer, "marginRight", req.Options.MarginRight)
	writeFormField(writer, "waitDelay", req.Options.WaitDelay)
	if req.Options.Landscape {
		writeFormField(writer, "landscape", "true")
	}
	if req.Options.PreferCSSPageSize {
		writeFormField(writer, "preferCssPageSize", "true")
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	endpoint := c.baseURL + "/forms/chromium/convert/html"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	if req.OutputFilename != "" {
		httpReq.Header.Set("Gotenberg-Output-Filename", sanitizeOutputFilename(req.OutputFilename))
	}
	if req.TraceID != "" {
		httpReq.Header.Set("Gotenberg-Trace", req.TraceID)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	pdfBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gotenberg returned %d: %s", resp.StatusCode, strings.TrimSpace(string(pdfBytes)))
	}

	return pdfBytes, nil
}

func writeMultipartFile(writer *multipart.Writer, fieldName string, asset HTMLAsset) error {
	filename := path.Base(strings.TrimSpace(asset.Filename))
	if filename == "" {
		return fmt.Errorf("asset filename is required")
	}

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return err
	}

	_, err = part.Write(asset.Content)
	return err
}

func writeFormField(writer *multipart.Writer, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	_ = writer.WriteField(key, value)
}

func sanitizeOutputFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	name = strings.TrimSuffix(name, path.Ext(name))
	name = strings.ReplaceAll(name, " ", "_")
	return path.Base(name)
}
