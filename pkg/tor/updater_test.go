package tor_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/humper/tor_exit_nodes/models"
	"github.com/humper/tor_exit_nodes/pkg/database/memory"
	"github.com/humper/tor_exit_nodes/pkg/tor"
	"github.com/humper/tor_exit_nodes/testing/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	exitnodeServer *httptest.Server
	geoServer      *httptest.Server
)

func TestMain(m *testing.M) {
	exitnodeServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		results, ok := fixtures.MockEndpoints[url]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(results))
		w.WriteHeader(http.StatusOK)
	}))
	geoServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries := []tor.GeoQuery{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&queries)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		results := []tor.GeoResponse{}
		for _, q := range queries {
			result, ok := fixtures.GeoData[q.Query]
			if ok {
				result.Query = q.Query
			}
			results = append(results, result)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
		w.WriteHeader(http.StatusOK)
	}))

	code := m.Run()
	os.Exit(code)
}

func TestUpdateTorNodesCleanEmptyDatabase(t *testing.T) {
	ctx := context.Background()
	db, err := memory.New(ctx)
	require.NoError(t, err, "Failed to create memory database")

	tu := tor.NewTORUpdater(ctx, &tor.NewTorUpdaterParams{
		DB:           db,
		SourceURLs:   []string{exitnodeServer.URL + "/tor/clean"},
		GeoURL:       geoServer.URL,
		GeoBatchSize: 100,
		Client:       http.DefaultClient,
	})

	tu.DoUpdateTorExitNodes(ctx)

	pagination, err := db.TorExitNodes.GetAll(ctx, []string{}, &models.Pagination{})
	require.NoError(t, err, "Failed to get tor exit nodes")
	assert.Equal(t, int64(1744), pagination.TotalRows)
	require.Len(t, pagination.Rows, 10, "Unexpected number of tor exit nodes")

}

func TestUpdateTorNodesCleanFullDatabase(t *testing.T) {
	ctx := context.Background()
	db, err := memory.New(ctx)
	require.NoError(t, err, "Failed to create memory database")

	tu := tor.NewTORUpdater(ctx, &tor.NewTorUpdaterParams{
		DB:           db,
		SourceURLs:   []string{exitnodeServer.URL + "/tor/clean"},
		GeoURL:       geoServer.URL,
		GeoBatchSize: 100,
		Client:       http.DefaultClient,
	})

	tu.DoUpdateTorExitNodes(ctx)
	// calling this twice should not add any new nodes
	tu.DoUpdateTorExitNodes(ctx)

	pagination, err := db.TorExitNodes.GetAll(ctx, []string{}, &models.Pagination{})
	require.NoError(t, err, "Failed to get tor exit nodes")
	assert.Equal(t, int64(1744), pagination.TotalRows)
	require.Len(t, pagination.Rows, 10, "Unexpected number of tor exit nodes")
}

func TestUpdateTorNodesSmallOverlap(t *testing.T) {
	ctx := context.Background()
	db, err := memory.New(ctx)
	require.NoError(t, err, "Failed to create memory database")

	tu := tor.NewTORUpdater(ctx, &tor.NewTorUpdaterParams{
		DB: db,
		SourceURLs: []string{
			exitnodeServer.URL + "/tor/small",
			exitnodeServer.URL + "/tor/small_overlap",
		},
		GeoURL:       geoServer.URL,
		GeoBatchSize: 100,
		Client:       http.DefaultClient,
	})

	tu.DoUpdateTorExitNodes(ctx)

	pagination, err := db.TorExitNodes.GetAll(ctx, []string{}, &models.Pagination{})
	require.NoError(t, err, "Failed to get tor exit nodes")
	assert.Equal(t, int64(21), pagination.TotalRows)
	require.Len(t, pagination.Rows, 10, "Unexpected number of tor exit nodes")
}

func TestUpdateTorNodesSmallOverlapDelete(t *testing.T) {
	ctx := context.Background()
	db, err := memory.New(ctx)
	require.NoError(t, err, "Failed to create memory database")

	tu := tor.NewTORUpdater(ctx, &tor.NewTorUpdaterParams{
		DB: db,
		SourceURLs: []string{
			exitnodeServer.URL + "/tor/small",
			exitnodeServer.URL + "/tor/small_overlap",
		},
		GeoURL:       geoServer.URL,
		GeoBatchSize: 100,
		Client:       http.DefaultClient,
	})

	tu.DoUpdateTorExitNodes(ctx)

	tu.SourceURLs = []string{exitnodeServer.URL + "/tor/small"}
	tu.DoUpdateTorExitNodes(ctx)

	pagination, err := db.TorExitNodes.GetAll(ctx, []string{}, &models.Pagination{})
	require.NoError(t, err, "Failed to get tor exit nodes")
	assert.Equal(t, int64(17), pagination.TotalRows)
	require.Len(t, pagination.Rows, 10, "Unexpected number of tor exit nodes")
}

func TestGeoUpdates(t *testing.T) {
	ctx := context.Background()
	db, err := memory.New(ctx)
	require.NoError(t, err, "Failed to create memory database")

	tu := tor.NewTORUpdater(ctx, &tor.NewTorUpdaterParams{
		DB:           db,
		SourceURLs:   []string{exitnodeServer.URL + "/tor/tiny"},
		GeoURL:       geoServer.URL,
		GeoBatchSize: 100,
		Client:       http.DefaultClient,
	})

	tu.DoUpdateTorExitNodes(ctx)

	tu.DoUpdateGeoData(ctx)

	pagination, err := db.TorExitNodes.GetAll(ctx, []string{}, &models.Pagination{})
	require.NoError(t, err, "Failed to get tor exit nodes")
	assert.Len(t, pagination.Rows, 3, "Unexpected number of tor exit nodes")

	missing_countries, err := db.TorExitNodes.GetMissingCountries(ctx, 100)
	require.NoError(t, err, "Failed to get tor exit nodes with missing countries")
	assert.Len(t, missing_countries, 0, "Unexpected number of tor exit nodes with missing countries")

	for _, node := range pagination.Rows.([]*models.TorExitNode) {
		assert.NotEmpty(t, node.CountryName, "Country should not be empty")
		assert.NotEmpty(t, node.CountryCode, "CountryCode should not be empty")
	}

}
