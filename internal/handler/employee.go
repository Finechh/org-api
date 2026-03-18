package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"org-api/internal/model"
	"org-api/internal/service"
	"org-api/pkg/logger"
)

type EmployeeHandler struct {
	svc service.EmployeeService
	log *logger.Logger
}

func NewEmployeeHandler(svc service.EmployeeService, log *logger.Logger) *EmployeeHandler {
	return &EmployeeHandler{svc: svc, log: log}
}

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	deptID, ok := extractID(r, w)
	if !ok {
		return
	}
	var req model.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	emp, err := h.svc.Create(r.Context(), deptID, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, emp)
}

func (h *EmployeeHandler) handleServiceError(w http.ResponseWriter, err error) {
	h.log.Errorf("service error: %v", err)
	switch {
	case errors.Is(err, service.ErrNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrBadRequest):
		respondError(w, http.StatusBadRequest, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}
