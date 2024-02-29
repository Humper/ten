package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/humper/tor_exit_nodes/pkg/auth"
)

func (s *Server) AddTorRoutes(ctx context.Context, mux *http.ServeMux) {
	mux.HandleFunc("GET /tor", func(w http.ResponseWriter, r *http.Request) {
		s.HandleGetTorExitNodes(ctx, w, r)
	})
}

func (s *Server) HandleGetTorExitNodes(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	pagination, err := getPagination(w, r)
	if err != nil {
		return
	}

	user := auth.GetUser(r.Context())
	allowed_ips := []string{}
	if user != nil {
		allowed_ips = user.AllowedIPs
	}

	pagination, err = s.db.TorExitNodes.GetAll(ctx, allowed_ips, pagination)
	if err != nil {
		HttpError(w, "Failed to get tor exit nodes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pagination)
}
