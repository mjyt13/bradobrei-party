package geo

import "bradobrei/backend/internal/models"

// EnrichSalonLatLon заполняет вычисляемые поля latitude/longitude для JSON-ответов API.
func EnrichSalonLatLon(s *models.Salon) {
	if s == nil {
		return
	}
	s.Latitude = nil
	s.Longitude = nil
	if s.Location == nil || *s.Location == "" {
		return
	}
	lat, lon, ok := ParsePointWKT(*s.Location)
	if !ok {
		return
	}
	s.Latitude = &lat
	s.Longitude = &lon
}
