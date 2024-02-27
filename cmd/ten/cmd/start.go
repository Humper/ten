package cmd

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/humper/tor_exit_nodes/pkg/database/psql"
	"github.com/humper/tor_exit_nodes/pkg/server"
	"github.com/humper/tor_exit_nodes/pkg/tor"
	"github.com/humper/tor_exit_nodes/pkg/util"
	"github.com/spf13/cobra"

	etcd "go.etcd.io/etcd/client/v3"
)

var (
	instance     string
	port         int
	configPath   string
	dbConfigPath string
)

type config struct {
	GeolocationUrl       string   `yaml:"geolocation_url"`
	GeolocationBatchSize int      `yaml:"geolocation_batch_size"`
	TorSourceURLs        []string `yaml:"tor_source_urls"`
	EtcdHost             string   `yaml:"etcd_host"`
}

func makeStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start ten server",
		Args:  cobra.NoArgs,
	}

	port := cmd.PersistentFlags().Int("port", 8080, "server port")
	configPath := cmd.PersistentFlags().String("config_path", "/app_config/ten.yaml", "config file path")
	dbConfigPath := cmd.PersistentFlags().String("db_config_path", "/app_config/db.yaml", "DB config file path")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		cfg, err := util.ReadYamlFile[config](*configPath)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to load configuration", "error", err)
			os.Exit(-1)
		}
		slog.InfoContext(ctx, "Configuration loaded")

		db, err := psql.Load(ctx, *dbConfigPath)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to load DB configuration", "error", err)
			os.Exit(-1)
		}
		slog.InfoContext(ctx, "Database connection successful")

		tuParams := &tor.NewTorUpdaterParams{
			DB:           db,
			SourceURLs:   cfg.TorSourceURLs,
			GeoURL:       cfg.GeolocationUrl,
			GeoBatchSize: cfg.GeolocationBatchSize,
			Client:       http.DefaultClient,
		}

		torUpdater := tor.NewTORUpdater(ctx, tuParams)

		etcdClient, err := etcd.New(etcd.Config{
			Endpoints: []string{cfg.EtcdHost},
		})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create etcd client", "error", err, "etcd_host", cfg.EtcdHost)
			os.Exit(-1)
		}
		defer etcdClient.Close()

		params := &server.NewServerParams{
			DB:         db,
			TorUpdater: torUpdater,
			ETCD:       etcdClient,
		}

		wctx, cancel := context.WithCancel(ctx)

		s := server.New(ctx, params)

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer cancel()

			slog.InfoContext(ctx, "Starting server", "port", *port)
			if err := s.Serve(wctx, *port); err != nil {
				slog.ErrorContext(ctx, "Failed to start server", "error", err)
				os.Exit(-1)
			}
		}()

		go func() {
			defer cancel()
			sig := <-util.NewInterruptChan()
			slog.InfoContext(ctx, "Received signal", "signal", sig)
		}()

		cancel()
		wg.Wait()
	}

	return cmd
}
