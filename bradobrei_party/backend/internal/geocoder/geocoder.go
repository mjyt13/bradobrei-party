package geocoder

import (
	"context"
)

// Result — нормализованный ответ геокодера (серверный ключ, не для браузера).
type Result struct {
	Lat              float64
	Lon              float64
	FormattedAddress string
}

// Geocoder вызывает внешний API геокодирования (Yandex / Google) на backend.
type Geocoder interface {
	Provider() string
	Geocode(ctx context.Context, address string) (Result, error)
}
