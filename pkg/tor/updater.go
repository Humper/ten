package tor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/humper/tor_exit_nodes/models"
	"github.com/humper/tor_exit_nodes/pkg/database"
	"gorm.io/gorm"
)

type TORUpdater struct {
	DB           *database.Database
	SourceURLs   []string
	GeoURL       string
	GeoBatchSize int
	Client       *http.Client
}

type NewTorUpdaterParams struct {
	DB           *database.Database
	SourceURLs   []string
	GeoURL       string
	GeoBatchSize int
	Client       *http.Client
}

type GeoQuery struct {
	Query  string `json:"query"`
	Fields string `json:"fields"`
}

type GeoResponse struct {
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Query       string `json:"query"`
}

func NewTORUpdater(ctx context.Context, params *NewTorUpdaterParams) *TORUpdater {
	tu := &TORUpdater{
		DB:           params.DB,
		SourceURLs:   params.SourceURLs,
		GeoURL:       params.GeoURL,
		GeoBatchSize: params.GeoBatchSize,
		Client:       params.Client,
	}

	return tu
}

func (tu *TORUpdater) UpdateTorExitNodes(ctx context.Context) {
	// get the lists of tor exit nodes from all known sources and take their union

	tu.DoUpdateTorExitNodes(ctx) // initial update

	updateTicker := time.NewTicker(1 * time.Hour)
	geoTicker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-updateTicker.C:
			tu.DoUpdateTorExitNodes(ctx)
		case <-geoTicker.C:
			tu.DoUpdateGeoData(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (tu *TORUpdater) DoUpdateGeoData(ctx context.Context) {

	// Look for nodes that don't have country information

	missing_countries, err := tu.DB.TorExitNodes.GetMissingCountries(ctx, tu.GeoBatchSize)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get tor exit nodes with missing countries", "error", err)
		return
	}
	if len(missing_countries) == 0 {
		return
	}

	slog.InfoContext(ctx, "Updating tor exit countries", "num_missing_countries", len(missing_countries))

	err = tu.addGeoData(ctx, missing_countries)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get geo data", "error", err)
		return
	}

	if err := tu.DB.TorExitNodes.Update(ctx, missing_countries); err != nil {
		slog.ErrorContext(ctx, "Failed to update countries", "error", err)
		return
	}
}

func (tu *TORUpdater) addGeoData(ctx context.Context, nodes []*models.TorExitNode) error {

	queries := []GeoQuery{}
	for _, node := range nodes {
		queries = append(queries, GeoQuery{
			Query:  node.IP,
			Fields: "country,countryCode,query",
		})
	}

	bodyBytes, err := json.Marshal(queries)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, tu.GeoURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := tu.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var geoResponses []GeoResponse
	if err := json.Unmarshal(respBody, &geoResponses); err != nil {
		return err
	}

	for i, node := range nodes {
		node.CountryName = geoResponses[i].Country
		node.CountryCode = geoResponses[i].CountryCode
	}

	return nil
}

func (tu *TORUpdater) DoUpdateTorExitNodes(ctx context.Context) {
	found_ips := mapset.NewSet[string]()

	for _, source := range tu.SourceURLs {
		req, err := http.NewRequest(http.MethodGet, source, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create request", "error", err, "source", source)
			continue
		}

		resp, err := tu.Client.Do(req)

		if err != nil {
			slog.ErrorContext(ctx, "Failed to get tor exit nodes", "error", err, "source", source)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			slog.ErrorContext(ctx, "Failed to get tor exit nodes", "status", resp.StatusCode, "source", source)
			continue
		}

		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to read response body", "error", err, "source", source)
			continue
		}

		ips := strings.Split(string(respBody), "\n")
		found_ips.Append(ips...)
	}

	existing_ip_set := mapset.NewSet[string]()
	existing_exit_nodes_by_ip := map[string]*models.TorExitNode{}

	page_count := 1
	pagination := &models.Pagination{
		Page:  1,
		Limit: 100,
	}

	total_found := 0
	for {
		pagination, err := tu.DB.TorExitNodes.GetAll(ctx, []string{}, pagination)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get tor exit nodes", "error", err)
			return
		}

		total_found += len(pagination.Rows.([]*models.TorExitNode))

		for _, node := range pagination.Rows.([]*models.TorExitNode) {
			existing_ip_set.Add(node.IP)
			existing_exit_nodes_by_ip[node.IP] = node
		}

		page_count++
		pagination.Page = page_count
		if page_count > pagination.TotalPages {
			break
		}
	}

	ips_to_delete := existing_ip_set.Difference(found_ips)
	ips_to_add := found_ips.Difference(existing_ip_set)

	nodes_to_delete := []models.TorExitNode{}
	for ip := range ips_to_delete.Iter() {
		nodes_to_delete = append(nodes_to_delete, models.TorExitNode{
			Model: gorm.Model{
				ID: existing_exit_nodes_by_ip[ip].ID,
			},
			IP: ip,
		})
	}
	nodes_to_add := []*models.TorExitNode{}
	for ip := range ips_to_add.Iter() {
		if ip != "" {
			nodes_to_add = append(nodes_to_add, &models.TorExitNode{IP: ip})
		}
	}

	if err := tu.DB.TorExitNodes.DeleteAndAdd(ctx, nodes_to_delete, nodes_to_add); err != nil {
		slog.ErrorContext(ctx, "Failed to update tor exit nodes", "error", err)
		return
	}
}
