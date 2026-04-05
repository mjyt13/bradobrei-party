package geocoder

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// NewFromEnv создаёт геокодер для Salon Service. Секретные ключи только на сервере.
// GEOCODER_PROVIDER: none | yandex | google (пусто = none).
func NewFromEnv() (Geocoder, error) {
	p := strings.TrimSpace(strings.ToLower(os.Getenv("GEOCODER_PROVIDER")))
	if p == "" || p == "none" {
		return nil, nil
	}
	switch p {
	case "yandex":
		k := strings.TrimSpace(os.Getenv("YANDEX_GEOCODER_API_KEY"))
		if k == "" {
			log.Println("предупреждение: GEOCODER_PROVIDER=yandex, но YANDEX_GEOCODER_API_KEY пуст — геокодер отключён")
			return nil, nil
		}
		log.Printf("геокодер: провайдер %s", "yandex")
		return NewYandex(k), nil
	case "google":
		k := strings.TrimSpace(os.Getenv("GOOGLE_GEOCODER_API_KEY"))
		if k == "" {
			log.Println("предупреждение: GEOCODER_PROVIDER=google, но GOOGLE_GEOCODER_API_KEY пуст — геокодер отключён")
			return nil, nil
		}
		log.Printf("геокодер: провайдер %s", "google")
		return NewGoogle(k), nil
	default:
		return nil, fmt.Errorf("неизвестный GEOCODER_PROVIDER=%q (ожидается none|yandex|google)", p)
	}
}
