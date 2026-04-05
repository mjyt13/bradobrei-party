package geocoder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Google struct {
	apiKey string
	client *http.Client
}

func NewGoogle(apiKey string) *Google {
	return &Google{
		apiKey: strings.TrimSpace(apiKey),
		client: &http.Client{Timeout: 12 * time.Second},
	}
}

func (g *Google) Provider() string { return "google" }

func (g *Google) Geocode(ctx context.Context, address string) (Result, error) {
	if g.apiKey == "" {
		return Result{}, fmt.Errorf("пустой API-ключ Google Geocoding")
	}
	u := "https://maps.googleapis.com/maps/api/geocode/json?" + url.Values{
		"address": {strings.TrimSpace(address)},
		"key":     {g.apiKey},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Result{}, err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("google geocoder: HTTP %d", resp.StatusCode)
	}

	var payload struct {
		Status  string `json:"status"`
		Results []struct {
			FormattedAddress string `json:"formatted_address"`
			Geometry         struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"results"`
		ErrorMessage string `json:"error_message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Result{}, err
	}
	if payload.Status != "OK" || len(payload.Results) == 0 {
		msg := payload.Status
		if payload.ErrorMessage != "" {
			msg = payload.ErrorMessage
		}
		return Result{}, fmt.Errorf("google geocoder: %s", msg)
	}
	r0 := payload.Results[0]
	return Result{
		Lat:              r0.Geometry.Location.Lat,
		Lon:              r0.Geometry.Location.Lng,
		FormattedAddress: strings.TrimSpace(r0.FormattedAddress),
	}, nil
}
