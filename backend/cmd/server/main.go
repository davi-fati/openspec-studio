// Package main is the OpenSpec Studio backend entrypoint.
//
//	@title			OpenSpec Studio API
//	@version		1.0
//	@description	REST + SSE API for the OpenSpec Studio backend.
//	@BasePath		/api
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/handler"
	"github.com/davidlima/openspec-studio/backend/internal/instancelock"
	"github.com/davidlima/openspec-studio/backend/internal/repository"
	"github.com/davidlima/openspec-studio/backend/internal/service"
	"github.com/davidlima/openspec-studio/backend/internal/webui"
)

const defaultPort = "4173"

// version is stamped at build time (-ldflags "-X main.version=...").
var version = "dev"

// readyPrefix starts the single stdout line a supervisor (the desktop
// shell) waits for before loading the UI.
const readyPrefix = "OPENSPEC_STUDIO_READY"

func main() {
	os.Exit(run())
}

func run() int {
	port := flag.String("port", "", "port to listen on; 0 picks a free port (default: $OPENSPEC_STUDIO_PORT or "+defaultPort+")")
	inheritLoginPath := flag.Bool("inherit-login-path", false, "resolve PATH from the user's login shell (set by the desktop app, which does not inherit the terminal environment)")
	flag.Parse()

	resolvedPort := *port
	if resolvedPort == "" {
		resolvedPort = os.Getenv("OPENSPEC_STUDIO_PORT")
	}
	if resolvedPort == "" {
		resolvedPort = defaultPort
	}

	handler.Version = version
	ctx := context.Background()

	if *inheritLoginPath {
		if p, err := loginShellPath(3 * time.Second); err != nil {
			log.Printf("login shell PATH unavailable, keeping inherited PATH: %v", err)
		} else {
			os.Setenv("PATH", p)
		}
	}

	dbPath, err := appDataDBPath()
	if err != nil {
		log.Printf("resolve app-data dir: %v", err)
		return 1
	}
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Printf("create app-data dir: %v", err)
		return 1
	}

	// Taken before opening the database: a second backend must not touch
	// studio.db or start a scheduler that would run the same flows again.
	lock, err := instancelock.Acquire(dataDir)
	if err != nil {
		log.Print(err)
		return 1
	}
	defer lock.Release()

	db, err := repository.Open(ctx, dbPath)
	if err != nil {
		log.Printf("open database: %v", err)
		return 1
	}
	defer db.Close()

	projectRepo := repository.NewProjectRepository(db)
	projectService := service.NewProjectService(projectRepo)
	specService := service.NewSpecService(projectService, currentAuthor())
	changeService := service.NewChangeService(projectService)
	contextService := service.NewContextService(projectService)
	overviewService := service.NewOverviewService(projectService, specService, changeService)

	credentialStore := repository.NewCredentialStore(dataDir)
	providerRepo := repository.NewProviderRepository(db, credentialStore)
	providerService := service.NewProviderService(providerRepo)
	specService.SetProviderInvoker(providerService)

	events := service.NewEventBroadcaster()

	specflowRepo := repository.NewSpecflowRepository(db)
	specflowService := service.NewSpecflowService(specflowRepo, projectService, providerService, events)

	// Flows overdue at launch are flagged missed, never silently auto-run -
	// this must happen before the scheduler starts polling.
	if err := specflowService.MarkMissedOverdueFlows(ctx); err != nil {
		log.Printf("specflow: mark missed flows: %v", err)
	}
	schedCtx, stopScheduler := context.WithCancel(ctx)
	defer stopScheduler()
	go specflowService.RunScheduler(schedCtx)

	watcher, err := service.NewProjectWatcher(events)
	if err != nil {
		log.Printf("start file watcher: %v", err)
		return 1
	}
	defer watcher.Close()
	go watcher.Run()

	// Re-arm watches for already-registered projects on startup.
	if existing, err := projectService.ListProjects(ctx); err == nil {
		for _, p := range existing {
			if p.Available {
				_ = watcher.WatchProject(p.Path)
			}
		}
	}

	router := handler.New(projectService, specService, changeService, contextService, overviewService, providerService, specflowService, watcher, events)
	ui, hasUI := webui.FS()
	if !hasUI {
		ui = nil
		log.Printf("no embedded frontend in this build; serving the API only")
	}
	handler.ServeUI(router, ui)

	ln, err := net.Listen("tcp", "127.0.0.1:"+resolvedPort)
	if err != nil {
		log.Printf("listen: %v", err)
		return 1
	}
	addr := ln.Addr().String()
	if err := lock.SetAddr(addr); err != nil {
		log.Printf("record address in instance lock: %v", err)
	}

	srv := &http.Server{Handler: router}
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()

	log.Printf("openspec-studio backend %s listening on %s", version, addr)
	fmt.Printf("%s addr=%s\n", readyPrefix, addr)

	sigCtx, stopSignals := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
			return 1
		}
		return 0
	case <-sigCtx.Done():
	}

	log.Printf("shutting down")
	stopScheduler()
	flowCtx, cancelFlows := context.WithTimeout(ctx, 10*time.Second)
	defer cancelFlows()
	if err := specflowService.Shutdown(flowCtx); err != nil {
		log.Printf("specflow: running flows did not stop in time: %v", err)
	}
	// SSE connections never go idle, so graceful shutdown is bounded and
	// then forced.
	httpCtx, cancelHTTP := context.WithTimeout(ctx, 2*time.Second)
	defer cancelHTTP()
	if err := srv.Shutdown(httpCtx); err != nil {
		srv.Close()
	}
	return 0
}

// appDataDBPath returns the path to the Studio's SQLite database file,
// stored in the OS's local app-data directory, outside any project's own
// openspec/ folder.
func appDataDBPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "openspec-studio", "studio.db"), nil
}

// currentAuthor resolves the identity recorded on provenance metadata for
// specs/changes created through the Studio. No user accounts exist yet, so
// this is simply the OS user running the backend.
func currentAuthor() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	if name := os.Getenv("USER"); name != "" {
		return name
	}
	return "unknown"
}
