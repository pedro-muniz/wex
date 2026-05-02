package controllers

import "net/http"

type APIControllers struct {
	TxController   *TransactionController
	ConvController *ConversionController
}

func NewAPIControllers(tx *TransactionController, conv *ConversionController) *APIControllers {
	return &APIControllers{
		TxController:   tx,
		ConvController: conv,
	}
}

func (c *APIControllers) RegisterAll(mux *http.ServeMux) {
	c.TxController.RegisterRoutes(mux)
	c.ConvController.RegisterRoutes(mux)
}
