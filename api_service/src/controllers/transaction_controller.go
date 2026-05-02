package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"wex/api_service/src/core/domain"
	"wex/api_service/src/core/ports"
	"wex/api_service/src/core/services"

	"github.com/google/uuid"
)

type TransactionController struct {
	txProducerService *services.TransactionProducerService
}

func NewTransactionController(txProducerService *services.TransactionProducerService) *TransactionController {
	return &TransactionController{
		txProducerService: txProducerService,
	}
}

// HandleCreateTransaction godoc
// @Summary Create a new purchase transaction
// @Description Validates and stores a transaction payload, then publishes it for asynchronous processing
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body ports.TransactionRequestDTO true "Transaction Payload"
// @Success 202 {object} map[string]string "Accepted"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /transactions [post]
func (c *TransactionController) HandleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req ports.TransactionRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := c.txProducerService.CreateTransaction(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"id": id.String()})
}

// HandleGetStatus godoc
// @Summary Check transaction processing status
// @Description Retrieves the current status from Valkey
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} map[string]string "Status"
// @Failure 400 {string} string "Invalid UUID"
// @Failure 404 {string} string "Status not found"
// @Router /transactions/{id}/status [get]
func (c *TransactionController) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	status, err := c.txProducerService.GetTransactionStatus(r.Context(), id)
	if err != nil {
		http.Error(w, "Status not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"id": id.String(), "status": string(status)})
}

func (c *TransactionController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /transactions", c.HandleCreateTransaction)
	mux.HandleFunc("GET /transactions/{id}/status", c.HandleGetStatus)
}
