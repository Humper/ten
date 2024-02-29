package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/humper/tor_exit_nodes/models"
	"github.com/humper/tor_exit_nodes/pkg/auth"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) HandleLogin(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		HttpError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := s.db.Users.GetByEmail(ctx, loginReq.Email)
	if err != nil {
		HttpError(w, "Unknown user", http.StatusUnauthorized)
		return
	}

	if !auth.ComparePassword(loginReq.Password, user.Password) {
		HttpError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	signedToken, expirationTime, err := auth.CreateJWT(user)
	if err != nil {
		HttpError(w, "Failed to create JWT", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    signedToken,
		Expires:  expirationTime,
		Domain:   "localhost",
		Path:     "/",
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *Server) HandleRegister(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		HttpError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if _, err := s.db.Users.GetByEmail(ctx, u.Email); err == nil {
		HttpError(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := auth.HashPassword(u.Password)
	if err != nil {
		HttpError(w, "Failed to hash password", http.StatusBadRequest)
		return
	}

	u.Password = hashedPassword

	if err := s.db.Users.Create(ctx, &u); err != nil {
		HttpError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(u)
}

func (s *Server) HandleLogout(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		MaxAge:   0,
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) HandleGetUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	pagination, err := getPagination(w, r)
	if err != nil {
		return
	}

	pagination, err = s.db.Users.GetAll(ctx, pagination)
	if err != nil {
		HttpError(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pagination)
}

func (s *Server) HandleGetUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		HttpError(w, "Invalid user id", http.StatusBadRequest)
		return
	}
	user, err := s.db.Users.GetByID(ctx, uint(id))
	if err != nil {
		HttpError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *Server) HandleUpdateUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		HttpError(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		HttpError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	existingUser, err := s.db.Users.GetByID(ctx, uint(id))
	if err != nil {
		HttpError(w, "Unknown user", http.StatusInternalServerError)
		return
	}

	loggedInUser := auth.GetUser(r.Context())
	if loggedInUser == nil || !(loggedInUser.Role == "admin" || loggedInUser.ID == u.ID) {
		HttpError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	u.ID = existingUser.ID

	if err := s.db.Users.Update(ctx, &u); err != nil {
		HttpError(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func (s *Server) HandleDeleteUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil || user.Role != "admin" {
		HttpError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		HttpError(w, "Invalid user id", http.StatusBadRequest)
		return
	}
	u, err := s.db.Users.Delete(ctx, uint(id))
	if err != nil {
		HttpError(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func (s *Server) AddAuthRoutes(ctx context.Context, mux *http.ServeMux) {
	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		s.HandleLogin(ctx, w, r)
	})
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		s.HandleRegister(ctx, w, r)
	})
	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		s.HandleGetUsers(ctx, w, r)
	})
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		s.HandleGetUser(ctx, w, r)
	})
	mux.HandleFunc("PUT /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		s.HandleUpdateUser(ctx, w, r)
	})
	mux.HandleFunc("DELETE /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		s.HandleDeleteUser(ctx, w, r)
	})
	mux.HandleFunc("POST /logout", func(w http.ResponseWriter, r *http.Request) {
		s.HandleLogout(ctx, w, r)
	})
}
