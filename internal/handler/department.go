package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"org-api/internal/model"
	"org-api/internal/service"
	"org-api/pkg/logger"
)

type DepartmentHandler struct {
	svc service.DepartmentService
	log *logger.Logger
}

func NewDepartmentHandler(svc service.DepartmentService, log *logger.Logger) *DepartmentHandler {
	return &DepartmentHandler{svc: svc, log: log}
}

func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	dept, err := h.svc.Create(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, dept)
}

func (h *DepartmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := extractID(r, w)
	if !ok {
		return
	}

	depth := 1
	if d := r.URL.Query().Get("depth"); d != "" {
		if v, err := strconv.Atoi(d); err == nil {
			depth = v
		}
	}
	includeEmployees := true
	if ie := r.URL.Query().Get("include_employees"); ie == "false" {
		includeEmployees = false
	}

	resp, err := h.svc.GetTree(r.Context(), id, depth, includeEmployees)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := extractID(r, w)
	if !ok {
		return
	}
	var req model.UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	dept, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dept)
}

func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := extractID(r, w)
	if !ok {
		return
	}
	mode := r.URL.Query().Get("mode")
	var reassignTo *uint
	if rt := r.URL.Query().Get("reassign_to_department_id"); rt != "" {
		v, err := strconv.ParseUint(rt, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid reassign_to_department_id")
			return
		}
		uid := uint(v)
		reassignTo = &uid
	}
	if err := h.svc.Delete(r.Context(), id, mode, reassignTo); err != nil {
		h.handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DepartmentHandler) handleServiceError(w http.ResponseWriter, err error) {
	h.log.Errorf("service error: %v", err)
	switch {
	case errors.Is(err, service.ErrNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrConflict):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrBadRequest):
		respondError(w, http.StatusBadRequest, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}
