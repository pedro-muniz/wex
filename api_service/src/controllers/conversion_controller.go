package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"wex/api_service/src/core/services"

	"github.com/google/uuid"
)

type ConversionController struct {
	convProducerService *services.ConversionProducerService
}

func NewConversionController(convProducerService *services.ConversionProducerService) *ConversionController {
	return &ConversionController{
		convProducerService: convProducerService,
	}
}

// HandleRequestConversion godoc
// @Summary Request a currency conversion
// @Description Triggers an asynchronous conversion job for the specified currency
// @Tags transactions
// @Param id path string true "Transaction ID"
// @Param currency query string true "Target Currency Code (e.g. Brazil-Real)"
// @Success 202 {string} string "Accepted"
// @Failure 400 {string} string "Invalid Request"
// @Router /transactions/{id}/convert [post]
func (c *ConversionController) HandleRequestConversion(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	currency := r.URL.Query().Get("currency")
	if currency == "" {
		http.Error(w, "Currency parameter is required", http.StatusBadRequest)
		return
	}

	if err := c.convProducerService.RequestConversion(r.Context(), id, currency); err != nil {
		http.Error(w, "Failed to request conversion", http.StatusInternalServerError)
		return
	}

	resultKey := fmt.Sprintf("conversion:%s:%s", id.String(), currency)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"key": resultKey})
}

// HandleGetConversionResult godoc
// @Summary Retrieve a conversion result
// @Description Fetches the result of a conversion job from Valkey
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID"
// @Param currency query string true "Target Currency Code (e.g. Brazil-Real)"
// @Success 200 {object} ports.TransactionResponseDTO
// @Failure 404 {string} string "Result not found yet"
// @Router /transactions/{id}/convert [get]
func (c *ConversionController) HandleGetConversionResult(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	currency := r.URL.Query().Get("currency")
	if currency == "" {
		http.Error(w, "Currency parameter is required", http.StatusBadRequest)
		return
	}

	resultKey := fmt.Sprintf("conversion:%s:%s", id.String(), currency)
	respData, err := c.convProducerService.GetConversionResult(r.Context(), resultKey)
	if err != nil || respData == "" {
		http.Error(w, "Conversion result not found or still processing", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(respData))
}

func (c *ConversionController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /transactions/{id}/convert", c.HandleRequestConversion)
	mux.HandleFunc("GET /transactions/{id}/convert", c.HandleGetConversionResult)
}
