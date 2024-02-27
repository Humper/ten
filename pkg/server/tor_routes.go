package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/humper/tor_exit_nodes/models"
	"github.com/humper/tor_exit_nodes/pkg/auth"
)

func (s *Server) AddTorRoutes(ctx context.Context, mux *http.ServeMux) {
	mux.HandleFunc("GET /tor", func(w http.ResponseWriter, r *http.Request) {
		s.HandleGetTorExitNodes(ctx, w, r)
	})
	mux.HandleFunc("GET /country_code", func(w http.ResponseWriter, r *http.Request) {
		s.HandleGetTorCountryCodes(ctx, w, r)
	})
}

func (s *Server) HandleGetTorExitNodes(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	pagination, err := getPagination(w, r)
	if err != nil {
		return
	}

	countryFilters, ok := r.URL.Query()["countries"]
	if !ok {
		countryFilters = []string{}
	}

	user := auth.GetUser(r.Context())
	allowed_ips := []string{}
	if user != nil {
		allowed_ips = user.AllowedIPs
	}

	pagination, err = s.db.TorExitNodes.GetAll(ctx, countryFilters, allowed_ips, pagination)
	if err != nil {
		HttpError(w, "Failed to get tor exit nodes", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pagination)
}

func (s *Server) HandleGetTorCountryCodes(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	pagestr := r.URL.Query().Get("page")
	if pagestr == "" {
		pagestr = "1"
	}
	limitstr := r.URL.Query().Get("limit")
	if limitstr == "" {
		limitstr = "50"
	}

	page, err := strconv.Atoi(pagestr)
	if err != nil {
		HttpError(w, "Invalid page", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(limitstr)
	if err != nil {
		HttpError(w, "Invalid limit", http.StatusBadRequest)
		return
	}
	sortColumn := r.URL.Query().Get("sort")

	pagination := &models.Pagination{
		Page:  page,
		Limit: limit,
		Sort:  sortColumn,
	}

	pagination, err = s.db.TorExitNodes.GetUniqueCountryCodes(ctx, pagination)
	if err != nil {
		HttpError(w, "Failed to get tor exit nodes", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pagination)
}
