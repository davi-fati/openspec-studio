SHELL := /bin/bash

.PHONY: help install backend frontend dev build ui-embed sidecar desktop-dev desktop-build desktop-stop desktop-clean desktop-rebuild icons test sqlc swagger kill-ports clean

BACKEND_PORT ?= 4173
FRONTEND_PORT ?= 5173
SWAG := $(shell go env GOPATH)/bin/swag
# Must match src-tauri/tauri.conf.json: the desktop app only attaches to an
# already-running backend that reports the same version.
VERSION ?= 0.1.0
GO_LDFLAGS := -X main.version=$(VERSION)
WEBUI_DIST := backend/internal/webui/dist
TARGET_TRIPLE := $(shell rustc -vV 2>/dev/null | sed -n 's/^host: //p')
SIDECAR := src-tauri/binaries/openspec-studio-server-$(TARGET_TRIPLE)
TAURI := frontend/node_modules/.bin/tauri
APP_NAME := OpenSpec Studio
APP_ID := com.davidlima.openspec-studio
BUNDLE_DIR := src-tauri/target/release/bundle
LSREGISTER := /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister
# Desktop processes only: the shell (bundled .app or `tauri dev`) and the
# sidecar backend it spawned. A web-mode backend (backend/bin/...) is left alone.
DESKTOP_APP_PROC := (Contents/MacOS|target/(debug|release))/openspec-studio( |$$)
DESKTOP_SIDECAR_PROC := (Contents/MacOS|target/(debug|release))/openspec-studio-server

help:
	@echo "OpenSpec Studio (web + desktop)"
	@echo ""
	@echo "  make install   Install backend and frontend dependencies"
	@echo "  make backend   Run the Go backend alone (127.0.0.1:$(BACKEND_PORT))"
	@echo "  make frontend  Run the frontend dev server alone (localhost:$(FRONTEND_PORT))"
	@echo "  make dev       Run backend + frontend together (Ctrl-C stops both)"
	@echo "  make build     Build one self-contained web binary (backend + embedded UI)"
	@echo "  make desktop-dev    Run the desktop app (Tauri) against a fresh full build"
	@echo "  make desktop-build  Build the macOS desktop app (.app/.dmg)"
	@echo "  make desktop-stop   Quit the desktop app and stop the backend it started"
	@echo "  make desktop-clean  Stop the app and remove src-tauri/target, installed copy and webview data (keeps studio.db)"
	@echo "  make desktop-rebuild  desktop-clean + desktop-build"
	@echo "  make icons     Regenerate desktop/web icons from the SVG masters (needs rsvg-convert)"
	@echo "  make test      Run backend tests and frontend typecheck/build"
	@echo "  make sqlc      Regenerate sqlc code from backend/internal/repository/sql"
	@echo "  make swagger   Regenerate the OpenAPI spec from handler annotations"
	@echo "  make kill-ports  Force-free $(BACKEND_PORT) and $(FRONTEND_PORT) (stray processes from a previous run)"
	@echo "  make clean     Remove build artifacts"

install:
	cd backend && go mod download
	cd frontend && npm install

# Frees a port by killing whatever is bound to it, if anything. Used as a
# self-heal step before backend/frontend start and as a final sweep when dev
# exits, so a process that survives a hard kill (e.g. `go run`'s child
# process outliving a SIGKILL'd wrapper) doesn't block the next run.
kill-ports:
	@lsof -ti:$(BACKEND_PORT) 2>/dev/null | xargs kill -9 2>/dev/null || true
	@lsof -ti:$(FRONTEND_PORT) 2>/dev/null | xargs kill -9 2>/dev/null || true

backend:
	@lsof -ti:$(BACKEND_PORT) 2>/dev/null | xargs kill -9 2>/dev/null || true
	cd backend && OPENSPEC_STUDIO_PORT=$(BACKEND_PORT) go run ./cmd/server

frontend:
	@lsof -ti:$(FRONTEND_PORT) 2>/dev/null | xargs kill -9 2>/dev/null || true
	cd frontend && npm run dev -- --port $(FRONTEND_PORT)

# Runs backend and frontend together. If either one exits for any reason
# (Ctrl-C, crash, port already in use), the other is killed too instead of
# being left running - plain `wait` would keep waiting on the survivor, and
# `wait -n` isn't available on macOS's default bash 3.2, hence the poll loop.
# The trap's final kill-ports sweep covers grandchild processes that
# `kill 0` alone might miss (e.g. `go run`'s compiled child).
dev:
	@trap 'kill 0 2>/dev/null; $(MAKE) kill-ports' EXIT INT TERM; \
	$(MAKE) backend & pid1=$$!; \
	$(MAKE) frontend & pid2=$$!; \
	while kill -0 $$pid1 2>/dev/null && kill -0 $$pid2 2>/dev/null; do sleep 1; done

# Builds the frontend and copies it where the backend embeds it from.
ui-embed:
	cd frontend && npm run build
	rm -rf $(WEBUI_DIST)
	mkdir -p $(WEBUI_DIST)
	cp -R frontend/dist/. $(WEBUI_DIST)/
	touch $(WEBUI_DIST)/.gitkeep

build: swagger ui-embed
	cd backend && go build -ldflags "$(GO_LDFLAGS)" -o bin/openspec-studio-server ./cmd/server

# The backend binary under the name Tauri's externalBin expects.
sidecar: build
	@test -n "$(TARGET_TRIPLE)" || { echo "rustc not found: install Rust to build the desktop app"; exit 1; }
	mkdir -p src-tauri/binaries
	cp backend/bin/openspec-studio-server $(SIDECAR)

desktop-dev: sidecar
	$(TAURI) dev

desktop-build: sidecar
	$(TAURI) build

# The app is killed first so it can't report the backend's exit or respawn
# it; the backend then gets SIGTERM (cancels running flows, closes the DB)
# and SIGKILL only after the same 13s grace the app itself allows.
desktop-stop:
	@pids=$$(pgrep -f '$(DESKTOP_APP_PROC)'); \
	if [ -n "$$pids" ]; then echo "stopping desktop app (pid $$pids)"; kill $$pids 2>/dev/null; fi; \
	pids=$$(pgrep -f '$(DESKTOP_SIDECAR_PROC)'); \
	if [ -n "$$pids" ]; then \
		echo "stopping desktop backend (pid $$pids)"; \
		kill -TERM $$pids 2>/dev/null; \
		for i in $$(seq 1 130); do pgrep -f '$(DESKTOP_SIDECAR_PROC)' >/dev/null || break; sleep 0.1; done; \
		pgrep -f '$(DESKTOP_SIDECAR_PROC)' | xargs kill -9 2>/dev/null || true; \
	fi

# Leaves the machine as if the app was never built or installed, except the
# Studio data dir (~/Library/Application Support/openspec-studio: projects,
# flows), which web mode shares. The whole Rust target dir goes too, so the
# next desktop-build recompiles the shell from scratch.
desktop-clean: desktop-stop
	@for vol in "/Volumes/$(APP_NAME)"* /Volumes/dmg.*; do \
		[ -d "$$vol/$(APP_NAME).app" ] && { echo "detaching $$vol"; hdiutil detach -force "$$vol" >/dev/null; }; \
	done; true
	@for app in "$(BUNDLE_DIR)/macos/$(APP_NAME).app" "/Applications/$(APP_NAME).app"; do \
		[ -d "$$app" ] && { $(LSREGISTER) -u "$$app" 2>/dev/null; echo "removing $$app"; rm -rf "$$app"; }; \
	done; true
	rm -rf src-tauri/target src-tauri/binaries
	rm -rf "$(HOME)/Library/WebKit/$(APP_ID)" "$(HOME)/Library/Caches/$(APP_ID)" \
		"$(HOME)/Library/Application Support/$(APP_ID)" \
		"$(HOME)/Library/Saved Application State/$(APP_ID).savedState"

desktop-rebuild:
	$(MAKE) desktop-clean
	$(MAKE) desktop-build

# SVG masters: src-tauri/icons/source.svg (app), src-tauri/icons/tray.svg
# (menu bar), frontend/public/favicon.svg (web). `tauri icon` also emits
# mobile/Windows Store sizes; only the files tauri.conf.json uses are kept.
ICON_TMP := $(shell mktemp -d -u)
icons:
	@command -v rsvg-convert >/dev/null || { echo "rsvg-convert not found: brew install librsvg"; exit 1; }
	mkdir -p $(ICON_TMP)
	rsvg-convert -w 1024 -h 1024 src-tauri/icons/source.svg -o $(ICON_TMP)/source.png
	$(TAURI) icon $(ICON_TMP)/source.png -o $(ICON_TMP)/out
	cd $(ICON_TMP)/out && cp 32x32.png 64x64.png 128x128.png 128x128@2x.png icon.icns icon.ico icon.png $(CURDIR)/src-tauri/icons/
	rsvg-convert -w 44 -h 44 src-tauri/icons/tray.svg -o src-tauri/icons/tray.png
	rsvg-convert -w 180 -h 180 frontend/public/favicon.svg -o frontend/public/apple-touch-icon.png
	rm -rf $(ICON_TMP)

test: swagger
	cd backend && go vet ./... && go test ./...
	cd frontend && npm run build

sqlc:
	cd backend && sqlc generate

# Regenerates backend/docs/ (OpenAPI spec) from swag annotations on the handlers.
# Installs the swag CLI into $GOPATH/bin on first use if it isn't there yet.
swagger:
	@if [ ! -x "$(SWAG)" ]; then \
		echo "installing swag CLI..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	cd backend && $(SWAG) init -g cmd/server/main.go -o docs

clean:
	rm -rf backend/bin backend/docs frontend/dist src-tauri/binaries src-tauri/target
	find $(WEBUI_DIST) -mindepth 1 ! -name .gitkeep -delete 2>/dev/null || true
