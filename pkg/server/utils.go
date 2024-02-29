package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/humper/tor_exit_nodes/models"
)

func getPagination(w http.ResponseWriter, r *http.Request) (*models.Pagination, error) {

	pagestr := r.URL.Query().Get("page")
	if pagestr == "" {
		pagestr = "1"
	}
	limitstr := r.URL.Query().Get("limit")
	if limitstr == "" {
		limitstr = "10"
	}

	page, err := strconv.Atoi(pagestr)
	if err != nil {
		HttpError(w, "Invalid page", http.StatusBadRequest)
		return nil, err
	}

	limit, err := strconv.Atoi(limitstr)
	if err != nil {
		HttpError(w, "Invalid limit", http.StatusBadRequest)
		return nil, err
	}
	sortColumn := r.URL.Query().Get("sort")

	filterStr := r.URL.Query().Get("filter")
	var filterDict map[string][]string
	if filterStr != "" {
		err := json.Unmarshal([]byte(filterStr), &filterDict)
		if err != nil {
			HttpError(w, "Invalid filter", http.StatusBadRequest)
			return nil, err
		}
	}

	return &models.Pagination{
		Page:   page,
		Limit:  limit,
		Sort:   sortColumn,
		Filter: filterDict,
	}, nil
}
