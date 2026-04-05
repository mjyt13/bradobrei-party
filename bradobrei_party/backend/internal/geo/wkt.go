package geo

import (
	"strconv"
	"strings"
)

// ParsePointWKT разбирает WKT точки PostGIS в формате POINT(долгота широта) для SRID 4326.
func ParsePointWKT(raw string) (lat, lon float64, ok bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, 0, false
	}
	upper := strings.ToUpper(s)
	if strings.HasPrefix(upper, "SRID=") {
		if idx := strings.Index(s, ";"); idx >= 0 {
			s = strings.TrimSpace(s[idx+1:])
			upper = strings.ToUpper(s)
		}
	}
	if !strings.HasPrefix(upper, "POINT(") || !strings.HasSuffix(s, ")") {
		return 0, 0, false
	}
	inner := strings.TrimSpace(s[len("POINT(") : len(s)-1])
	inner = strings.NewReplacer(",", " ").Replace(inner)
	parts := strings.Fields(inner)
	if len(parts) != 2 {
		return 0, 0, false
	}
	a, err1 := strconv.ParseFloat(parts[0], 64)
	b, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	// В проекте и в PostGIS для 4326 хранится порядок долгота, широта.
	lon, lat = a, b
	return lat, lon, true
}
