package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/geocoder"
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"
)

type SalonService struct {
	salonRepo *repository.SalonRepository
	geocoder  geocoder.Geocoder
}

func NewSalonService(salonRepo *repository.SalonRepository, g geocoder.Geocoder) *SalonService {
	return &SalonService{salonRepo: salonRepo, geocoder: g}
}

func (s *SalonService) GeocoderEnabled() bool {
	return s.geocoder != nil
}

func (s *SalonService) GeocodeAddress(ctx context.Context, address string) (dto.GeocodeAddressResponse, error) {
	var out dto.GeocodeAddressResponse
	if s.geocoder == nil {
		return out, fmt.Errorf("геокодер не настроен")
	}

	res, err := s.geocoder.Geocode(ctx, strings.TrimSpace(address))
	if err != nil {
		return out, err
	}

	out.Latitude = res.Lat
	out.Longitude = res.Lon
	out.FormattedAddress = res.FormattedAddress
	out.Provider = s.geocoder.Provider()
	return out, nil
}

func (s *SalonService) GetAll() ([]models.Salon, error) {
	return s.salonRepo.GetAll()
}

func (s *SalonService) GetByID(id uint) (*models.Salon, error) {
	return s.salonRepo.GetByID(id)
}

func (s *SalonService) Create(salon *models.Salon) error {
	ctx := context.Background()
	if err := s.applyServerGeocoding(ctx, salon); err != nil {
		return err
	}
	if err := normalizeSalonWorkingHours(salon); err != nil {
		return err
	}
	if s.geocoder == nil {
		if err := normalizeSalonLocation(salon); err != nil {
			return err
		}
	}
	return s.salonRepo.Create(salon)
}

func (s *SalonService) Update(salon *models.Salon) error {
	ctx := context.Background()
	if err := s.applyServerGeocoding(ctx, salon); err != nil {
		return err
	}
	if err := normalizeSalonWorkingHours(salon); err != nil {
		return err
	}
	if s.geocoder == nil {
		if err := normalizeSalonLocation(salon); err != nil {
			return err
		}
	}
	return s.salonRepo.Update(salon)
}

func (s *SalonService) applyServerGeocoding(ctx context.Context, salon *models.Salon) error {
	if s.geocoder == nil {
		return nil
	}

	addr := strings.TrimSpace(salon.Address)
	if addr == "" {
		return fmt.Errorf("укажите адрес салона для проверки геокодером")
	}

	res, err := s.geocoder.Geocode(ctx, addr)
	if err != nil {
		return fmt.Errorf("геокодирование адреса: %w", err)
	}

	wkt := fmt.Sprintf("POINT(%g %g)", res.Lon, res.Lat)
	salon.Location = &wkt
	return nil
}

func (s *SalonService) Delete(id uint) error {
	return s.salonRepo.Delete(id)
}

func (s *SalonService) GetMasters(salonID uint) ([]models.EmployeeProfile, error) {
	return s.salonRepo.GetMasters(salonID)
}

func normalizeSalonLocation(salon *models.Salon) error {
	if salon.Location == nil {
		return nil
	}

	raw := strings.TrimSpace(*salon.Location)
	if raw == "" {
		salon.Location = nil
		return nil
	}

	normalized, err := normalizePoint(raw)
	if err != nil {
		return err
	}

	salon.Location = &normalized
	return nil
}

func normalizeSalonWorkingHours(salon *models.Salon) error {
	if salon.WorkingHours == nil {
		return nil
	}

	normalized, err := normalizeScheduleJSON(*salon.WorkingHours, "working_hours")
	if err != nil {
		return err
	}

	salon.WorkingHours = normalized
	return nil
}

func normalizePoint(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	upper := strings.ToUpper(trimmed)

	if strings.HasPrefix(upper, "POINT(") && strings.HasSuffix(trimmed, ")") {
		inside := strings.TrimSpace(trimmed[len("POINT(") : len(trimmed)-1])
		if strings.Contains(inside, ",") {
			return normalizeLatLonPair(inside)
		}

		parts := strings.Fields(inside)
		if len(parts) != 2 {
			return "", fmt.Errorf("координаты должны быть в формате \"широта, долгота\" или \"POINT(долгота широта)\"")
		}

		lon, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return "", fmt.Errorf("не удалось разобрать долготу")
		}
		lat, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return "", fmt.Errorf("не удалось разобрать широту")
		}
		if err := validateLatLon(lat, lon); err != nil {
			return "", err
		}

		return fmt.Sprintf("POINT(%g %g)", lon, lat), nil
	}

	return normalizeLatLonPair(trimmed)
}

func normalizeLatLonPair(raw string) (string, error) {
	cleaned := strings.NewReplacer(",", " ", ";", " ").Replace(raw)
	parts := strings.Fields(cleaned)
	if len(parts) != 2 {
		return "", fmt.Errorf("координаты должны содержать широту и долготу")
	}

	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return "", fmt.Errorf("не удалось разобрать широту")
	}
	lon, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return "", fmt.Errorf("не удалось разобрать долготу")
	}

	if err := validateLatLon(lat, lon); err != nil {
		return "", err
	}

	return fmt.Sprintf("POINT(%g %g)", lon, lat), nil
}

func validateLatLon(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("широта должна быть в диапазоне от -90 до 90")
	}
	if lon < -180 || lon > 180 {
		return fmt.Errorf("долгота должна быть в диапазоне от -180 до 180")
	}
	return nil
}
