package geocoder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Yandex struct {
	apiKey string
	client *http.Client
}

func NewYandex(apiKey string) *Yandex {
	return &Yandex{
		apiKey: strings.TrimSpace(apiKey),
		client: &http.Client{Timeout: 12 * time.Second},
	}
}

func (y *Yandex) Provider() string { return "yandex" }

func (y *Yandex) Geocode(ctx context.Context, address string) (Result, error) {
	if y.apiKey == "" {
		return Result{}, fmt.Errorf("пустой API-ключ Yandex Geocoder")
	}
	addr := strings.TrimSpace(address)
	if addr == "" {
		return Result{}, fmt.Errorf("пустой адрес для геокодирования")
	}

	// Сначала v1 (актуальный контракт), затем 1.x — часть ключей/тарифов отдаёт пустую выдачу только на одном из endpoint.
	bases := []string{
		"https://geocode-maps.yandex.ru/v1/",
		"https://geocode-maps.yandex.ru/1.x/",
	}
	var lastErr error
	for _, base := range bases {
		for _, withLang := range []bool{true, false} {
			q := yandexQueryParams(y.apiKey, addr, withLang)
			u := base + "?" + q.Encode()
			body, status, err := y.httpGet(ctx, u)
			if err != nil {
				lastErr = err
				continue
			}
			if status != http.StatusOK {
				lastErr = yandexHTTPError(status, body)
				continue
			}
			r, err := yandexParseGeocodeJSON(body)
			if err == nil {
				return r, nil
			}
			lastErr = err
		}
	}
	if lastErr != nil {
		return Result{}, lastErr
	}
	return Result{}, fmt.Errorf("yandex geocoder: не удалось выполнить запрос")
}

func yandexQueryParams(apiKey, addr string, withLang bool) url.Values {
	v := url.Values{}
	v.Set("apikey", apiKey)
	v.Set("geocode", addr)
	v.Set("format", "json")
	v.Set("results", "10")
	if withLang {
		v.Set("lang", "ru_RU")
	}
	return v
}

func (y *Yandex) httpGet(ctx context.Context, u string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "BradobreiParty/1.0 (server geocoder)")
	req.Header.Set("Accept", "application/json")

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func yandexHTTPError(status int, body []byte) error {
	msg := strings.TrimSpace(string(body))
	var apiErr struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
		msg = apiErr.Message
	}
	if status == http.StatusForbidden {
		return fmt.Errorf("yandex geocoder: HTTP 403 — неверный или неподходящий apikey (в кабинете developer.tech.yandex.ru включите «Геокодер» для этого ключа; 403 также бывает у ключа только для JS API). Детали: %s", msg)
	}
	if msg != "" {
		return fmt.Errorf("yandex geocoder: HTTP %d: %s", status, msg)
	}
	return fmt.Errorf("yandex geocoder: HTTP %d", status)
}

func yandexParseGeocodeJSON(body []byte) (Result, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return Result{}, fmt.Errorf("yandex geocoder: разбор JSON: %w", err)
	}

	if err := yandexRootAPIError(root); err != nil {
		return Result{}, err
	}

	resp, ok := mapFromCI(root, "response")
	if !ok {
		return Result{}, fmt.Errorf("yandex geocoder: в ответе нет поля response (первые байты: %s)", yandexSnippet(body))
	}

	goc, ok := mapFromCI(resp, "GeoObjectCollection")
	if !ok {
		return Result{}, fmt.Errorf("yandex geocoder: в ответе нет GeoObjectCollection (%s)", yandexSnippet(body))
	}

	foundStr := yandexCollectionFound(goc)
	fmRaw, _ := getCI(goc, "featureMember")
	members := yandexNormalizeFeatureMembers(fmRaw)
	if len(members) == 0 {
		if foundStr == "0" {
			return Result{}, fmt.Errorf("геокодер не нашёл объектов по этому адресу; уточните город, регион или попробуйте короче формулировку")
		}
		if foundStr != "" {
			return Result{}, fmt.Errorf("yandex geocoder: нет результатов (found=%s)", foundStr)
		}
		return Result{}, fmt.Errorf("yandex geocoder: пустой featureMember (%s)", yandexSnippet(body))
	}

	fm0, ok := members[0].(map[string]interface{})
	if !ok {
		return Result{}, fmt.Errorf("yandex geocoder: некорректный элемент featureMember[0]")
	}
	geoObj, ok := mapFromCI(fm0, "GeoObject")
	if !ok {
		return Result{}, fmt.Errorf("yandex geocoder: нет GeoObject в первом результате")
	}
	point, ok := mapFromCI(geoObj, "Point")
	if !ok {
		return Result{}, fmt.Errorf("yandex geocoder: нет Point в GeoObject")
	}
	posStr, ok := pointPosString(point)
	if !ok || strings.TrimSpace(posStr) == "" {
		return Result{}, fmt.Errorf("yandex geocoder: нет или пустой Point.pos")
	}

	text := ""
	if md, ok := mapFromCI(geoObj, "metaDataProperty"); ok {
		if gmd, ok := mapFromCI(md, "GeocoderMetaData"); ok {
			if t, ok := gmd["text"].(string); ok {
				text = t
			}
		}
	}

	lat, lon, err := yandexParsePos(posStr)
	if err != nil {
		return Result{}, err
	}
	return Result{Lat: lat, Lon: lon, FormattedAddress: strings.TrimSpace(text)}, nil
}

// yandexRootAPIError — иногда тело JSON с ошибкой приходит при HTTP 200 или без обёртки response.
func yandexRootAPIError(root map[string]interface{}) error {
	sc, ok := root["statusCode"]
	if !ok {
		return nil
	}
	var code float64
	switch v := sc.(type) {
	case float64:
		code = v
	case int:
		code = float64(v)
	default:
		return nil
	}
	if code == 0 || code == 200 {
		return nil
	}
	msg, _ := root["message"].(string)
	if msg == "" {
		msg, _ = root["error"].(string)
	}
	return fmt.Errorf("yandex geocoder: API statusCode=%v: %s", code, strings.TrimSpace(msg))
}

func yandexSnippet(body []byte) string {
	s := strings.TrimSpace(string(body))
	if len(s) > 180 {
		return s[:180] + "…"
	}
	return s
}

func yandexCollectionFound(goc map[string]interface{}) string {
	md, ok := mapFromCI(goc, "metaDataProperty")
	if !ok {
		return ""
	}
	grm, ok := mapFromCI(md, "GeocoderResponseMetaData")
	if !ok {
		return ""
	}
	switch v := grm["found"].(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case int:
		return strconv.Itoa(v)
	default:
		return ""
	}
}

func yandexNormalizeFeatureMembers(raw interface{}) []interface{} {
	if raw == nil {
		return nil
	}
	if arr, ok := raw.([]interface{}); ok {
		return arr
	}
	if obj, ok := raw.(map[string]interface{}); ok {
		return []interface{}{obj}
	}
	return nil
}

func pointPosString(point map[string]interface{}) (string, bool) {
	if s, ok := point["pos"].(string); ok {
		return s, true
	}
	// редко: числа в JSON
	if f, ok := point["pos"].(float64); ok {
		return fmt.Sprintf("%g", f), true
	}
	return "", false
}

func yandexParsePos(posStr string) (lat, lon float64, err error) {
	posStr = strings.TrimSpace(posStr)
	parts := strings.Fields(strings.NewReplacer(",", " ", ";", " ").Replace(posStr))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("некорректный формат координат Yandex: %q", posStr)
	}
	// API Яндекса: «долгота широта» (два числа через пробел).
	lon, err1 := strconv.ParseFloat(parts[0], 64)
	lat, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("не удалось разобрать координаты Yandex")
	}
	return lat, lon, nil
}

func getCI(m map[string]interface{}, want string) (interface{}, bool) {
	if m == nil {
		return nil, false
	}
	for k, v := range m {
		if strings.EqualFold(k, want) {
			return v, true
		}
	}
	return nil, false
}

func mapAsMap(v interface{}) (map[string]interface{}, bool) {
	m, ok := v.(map[string]interface{})
	return m, ok
}

func mapFromCI(m map[string]interface{}, key string) (map[string]interface{}, bool) {
	v, ok := getCI(m, key)
	if !ok {
		return nil, false
	}
	return mapAsMap(v)
}
