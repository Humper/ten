package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/humper/tor_exit_nodes/pkg/auth"
	"github.com/humper/tor_exit_nodes/pkg/database"
	"github.com/humper/tor_exit_nodes/pkg/tor"

	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type NewServerParams struct {
	DB         *database.Database
	TorUpdater *tor.TORUpdater
	ETCD       *etcd.Client
}

type Server struct {
	db         *database.Database
	mux        *http.ServeMux
	torUpdater *tor.TORUpdater
	etcd       *etcd.Client
}

func New(ctx context.Context, params *NewServerParams) *Server {
	s := &Server{
		db:         params.DB,
		torUpdater: params.TorUpdater,
		etcd:       params.ETCD,
	}

	mux := http.NewServeMux()

	s.AddAuthRoutes(ctx, mux)
	s.AddTorRoutes(ctx, mux)

	// mux.Handle("GET /", http.FileServer(http.Dir("static")))

	s.mux = mux
	go s.process(ctx)

	return s
}

func (s *Server) process(ctx context.Context) {
	for {
		slog.InfoContext(ctx, "Starting leader election")
		session, err := concurrency.NewSession(s.etcd, concurrency.WithTTL(10))
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create etcd session", "error", err)
			panic(err)
		}
		defer session.Close()

		slog.InfoContext(ctx, "Trying to become leader")
		election := concurrency.NewElection(session, "ten")
		if err := election.Campaign(ctx, "ten"); err != nil {
			slog.ErrorContext(ctx, "Failed to campaign for leadership", "error", err)
			panic(err)
		}

		slog.InfoContext(ctx, "Became leader")

		go s.torUpdater.UpdateTorExitNodes(ctx)
		select {
		case <-ctx.Done():
			if err := election.Resign(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to resign", "error", err)
			}
			slog.InfoContext(ctx, "Resigned Leadership")
			return
		}
	}
}

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) GetLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}
		claims, err := auth.ParseJWT(cookie.Value)
		if err != nil {
			slog.ErrorContext(r.Context(), "Couldn't parse JWT", "error", err, "cookie", cookie.Value)
			HttpError(w, "Couldn't parse JWT", http.StatusUnauthorized)
			return
		}

		userIdStr := claims.StandardClaims.Id
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			slog.ErrorContext(r.Context(), "Bad User ID in JWT", "error", err, "id", userIdStr)
			HttpError(w, "Bad User ID in JWT", http.StatusBadRequest)
			return
		}
		user, err := s.db.Users.GetByID(r.Context(), uint(userId))
		if err == nil {
			r = r.WithContext(auth.NewContext(r.Context(), user))
		} else {
			// this is ok - the user may have been deleted
			slog.InfoContext(r.Context(), "User not found", "id", userId, "error", err.Error())
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) Serve(ctx context.Context, port int) error {
	return http.ListenAndServe(fmt.Sprintf(":%d", port), Cors(s.GetLogin(s.mux)))
}
