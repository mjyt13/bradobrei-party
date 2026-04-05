package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/geo"
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type SalonHandler struct {
	salonService *services.SalonService
}

func NewSalonHandler(salonService *services.SalonService) *SalonHandler {
	return &SalonHandler{salonService: salonService}
}

// GetAll godoc
// @Summary РЎРїРёСЃРѕРє СЃР°Р»РѕРЅРѕРІ
// @Description Р‘Р°Р·РѕРІС‹Р№ СЃРїСЂР°РІРѕС‡РЅС‹Р№ СЌРЅРґРїРѕРёРЅС‚ РґР»СЏ РїСЂРѕСЃРјРѕС‚СЂР° РґРѕСЃС‚СѓРїРЅС‹С… С„РёР»РёР°Р»РѕРІ Рё Р°РЅР°Р»РёС‚РёРєРё 2.2.2.
// @Tags salons
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Salon
// @Failure 500 {object} dto.ErrorResponse
// @Router /salons [get]
func (h *SalonHandler) GetAll(c *gin.Context) {
	salons, err := h.salonService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	for i := range salons {
		geo.EnrichSalonLatLon(&salons[i])
	}
	c.JSON(http.StatusOK, salons)
}

// GetByID godoc
// @Summary РЎР°Р»РѕРЅ РїРѕ ID
// @Description Р’РѕР·РІСЂР°С‰Р°РµС‚ РѕРґРёРЅ СЃР°Р»РѕРЅ СЃ Р±Р°Р·РѕРІРѕР№ РёРЅС„РѕСЂРјР°С†РёРµР№.
// @Tags salons
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID СЃР°Р»РѕРЅР°"
// @Success 200 {object} models.Salon
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /salons/{id} [get]
func (h *SalonHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	salon, err := h.salonService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "РЎР°Р»РѕРЅ РЅРµ РЅР°Р№РґРµРЅ",
		})
		return
	}
	geo.EnrichSalonLatLon(salon)
	c.JSON(http.StatusOK, salon)
}

// GeocodeAddress godoc
// @Summary РџСЂРѕРІРµСЂРєР° Р°РґСЂРµСЃР° РіРµРѕРєРѕРґРµСЂРѕРј (СЃРµСЂРІРµСЂ)
// @Description Р’С‹Р·С‹РІР°РµС‚ РІРЅРµС€РЅРёР№ Geocoder API СЃ СЃРµРєСЂРµС‚РЅС‹Рј РєР»СЋС‡РѕРј РЅР° backend. Р”Р»СЏ РїСЂРµРґРїСЂРѕСЃРјРѕС‚СЂР° РєРѕРѕСЂРґРёРЅР°С‚ РїРµСЂРµРґ СЃРѕС…СЂР°РЅРµРЅРёРµРј СЃР°Р»РѕРЅР°.
// @Tags salons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.GeocodeAddressRequest true "РђРґСЂРµСЃ"
// @Success 200 {object} dto.GeocodeAddressResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /salons/geocode [post]
func (h *SalonHandler) GeocodeAddress(c *gin.Context) {
	if !h.salonService.GeocoderEnabled() {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "geocoder_unavailable",
			Code:    503,
			Message: "РЎРµСЂРІРµСЂРЅС‹Р№ РіРµРѕРєРѕРґРµСЂ РЅРµ РЅР°СЃС‚СЂРѕРµРЅ (СЃРј. GEOCODER_PROVIDER Рё РєР»СЋС‡Рё РІ .env)",
		})
		return
	}
	var req dto.GeocodeAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	out, err := h.salonService.GeocodeAddress(c.Request.Context(), req.Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "geocode_failed",
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, out)
}

// GetMasters godoc
// @Summary РњР°СЃС‚РµСЂР° СЃР°Р»РѕРЅР°
// @Description Р’РѕР·РІСЂР°С‰Р°РµС‚ РјР°СЃС‚РµСЂРѕРІ, Р·Р°РєСЂРµРїР»С‘РЅРЅС‹С… Р·Р° РІС‹Р±СЂР°РЅРЅС‹Рј СЃР°Р»РѕРЅРѕРј.
// @Tags salons
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID СЃР°Р»РѕРЅР°"
// @Success 200 {array} models.EmployeeProfile
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /salons/{id}/masters [get]
func (h *SalonHandler) GetMasters(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	masters, err := h.salonService.GetMasters(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, masters)
}

// Create godoc
// @Summary РЎРѕР·РґР°С‚СЊ СЃР°Р»РѕРЅ
// @Description РЎРѕР·РґР°С‘С‚ РЅРѕРІС‹Р№ СЃР°Р»РѕРЅ.
// @Tags salons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Salon true "Р”Р°РЅРЅС‹Рµ СЃР°Р»РѕРЅР°"
// @Success 201 {object} models.Salon
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /salons [post]
func (h *SalonHandler) Create(c *gin.Context) {
	var salon models.Salon
	if err := c.ShouldBindJSON(&salon); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	if err := h.salonService.Create(&salon); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	geo.EnrichSalonLatLon(&salon)
	c.JSON(http.StatusCreated, salon)
}

// Update godoc
// @Summary РћР±РЅРѕРІРёС‚СЊ СЃР°Р»РѕРЅ
// @Description РћР±РЅРѕРІР»СЏРµС‚ РґР°РЅРЅС‹Рµ СЃСѓС‰РµСЃС‚РІСѓСЋС‰РµРіРѕ СЃР°Р»РѕРЅР°.
// @Tags salons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID СЃР°Р»РѕРЅР°"
// @Param request body models.Salon true "РћР±РЅРѕРІР»С‘РЅРЅС‹Рµ РґР°РЅРЅС‹Рµ СЃР°Р»РѕРЅР°"
// @Success 200 {object} models.Salon
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /salons/{id} [put]
func (h *SalonHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	existing, err := h.salonService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "РЎР°Р»РѕРЅ РЅРµ РЅР°Р№РґРµРЅ",
		})
		return
	}

	if err := c.ShouldBindJSON(existing); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	if err := h.salonService.Update(existing); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	geo.EnrichSalonLatLon(existing)
	c.JSON(http.StatusOK, existing)
}

// Delete godoc
// @Summary РЈРґР°Р»РёС‚СЊ СЃР°Р»РѕРЅ
// @Description РЈРґР°Р»СЏРµС‚ СЃР°Р»РѕРЅ РёР· СЃРёСЃС‚РµРјС‹.
// @Tags salons
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID СЃР°Р»РѕРЅР°"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /salons/{id} [delete]
func (h *SalonHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	if err := h.salonService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "РЎР°Р»РѕРЅ СѓРґР°Р»С‘РЅ"})
}
