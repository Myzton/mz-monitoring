package http

import (
	"encoding/json"
	"log/slog"
	"mz-monitoring/internal/delivery/http/middleware"
	"mz-monitoring/internal/usecase"
	"net/http"
	"strconv"
)

type Target struct {
	URL         string `json:"url"`
	IntervalSec int    `json:"interval_sec"`
	IsActive    bool   `json:"is_active"`
}
type TargetHandler struct {
	Usecase *usecase.TargetUsecase
}

func NewTargetHandler(u *usecase.TargetUsecase) *TargetHandler {
	return &TargetHandler{u}
}

type ResponseTarget struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	ID      int    `json:"id"`
}

type TargetResponse struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	URL         string `json:"url"`
	IntervalSec int    `json:"interval_sec"`
	IsActive    bool   `json:"is_active"`
	IsOnline    bool   `json:"is_online"`
}

func (h *TargetHandler) CreateTargetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		slog.Error("error get  user id", "error", ok)
		http.Error(w, "error get  user id", http.StatusUnauthorized)
		return
	}

	var input Target

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		slog.Error("failed to parse json", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	targetCreated, err := h.Usecase.Create(ctx, userId, input.URL, input.IntervalSec)
	if err != nil {
		slog.Error("Error create target", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := ResponseTarget{Status: "success", Message: "Target created successfully", ID: targetCreated.ID}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		slog.Error("Error encode json", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TargetHandler) DeleteTargetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		slog.Error("error get  user id", "error", ok)
		http.Error(w, "error get  user id", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")

	targetId, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("invalid target id", "error", err)
		http.Error(w, "invalid target id", http.StatusBadRequest)
		return
	}

	err = h.Usecase.Delete(ctx, targetId, userId)
	if err != nil {
		slog.Error("error delete target", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := ResponseTarget{Status: "success", Message: "Target deleted"}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		slog.Error("Error encode json", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TargetHandler) GetListTargetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		slog.Error("error get  user id", "error", ok)
		http.Error(w, "error get  user id", http.StatusUnauthorized)
		return
	}

	domainTarget, err := h.Usecase.GetList(ctx, userId)
	if err != nil {
		slog.Error("error get target", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var responseTarget []TargetResponse

	for i := 0; i < len(domainTarget); i++ {
		responseTarget = append(responseTarget, TargetResponse{
			ID:          domainTarget[i].ID,
			UserID:      domainTarget[i].UserID,
			URL:         domainTarget[i].URL,
			IntervalSec: domainTarget[i].IntervalSec,
			IsActive:    domainTarget[i].IsActive,
			IsOnline:    domainTarget[i].IsOnline,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(responseTarget)
	if err != nil {
		slog.Error("Error encode json", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
