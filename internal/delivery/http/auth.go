package http

import (
	"encoding/json"
	"log/slog"
	"mz-monitoring/internal/usecase"
	"net/http"
)

type Register struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type UserHandler struct {
	Usecase *usecase.UserUsecase
}

func NewUserHandler(u *usecase.UserUsecase) *UserHandler {
	return &UserHandler{Usecase: u}
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type LoginStr struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input Register

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		slog.Error("failed to parse json", "err", err)
		http.Error(w, "failed to parse json", http.StatusBadRequest)
		return
	}

	if input.Email == "" || input.Password == "" {
		slog.Error("fields must not be empty")
		http.Error(w, "fields must not be empty", http.StatusBadRequest)
		return
	}

	_, err = h.Usecase.Create(ctx, input.Name, input.Password, input.Email)
	if err != nil {
		slog.Error("user creation error", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := Response{Status: "success", Message: "User created successfully"}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		slog.Error("json transfer error", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var inputLogin LoginStr

	err := json.NewDecoder(r.Body).Decode(&inputLogin)

	if err != nil {
		slog.Error("an error has occurred", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if inputLogin.Email == "" || inputLogin.Password == "" {
		slog.Error("Email or Password are empty")
		http.Error(w, "Email or Password are empty", http.StatusBadRequest)
		return
	}

	token, err := h.Usecase.Login(ctx, inputLogin.Email, inputLogin.Password)
	if err != nil {
		slog.Error("an error occurred", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokenJs := TokenResponse{token}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(tokenJs)

	if err != nil {
		slog.Error("an error occurred in answer json", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
