BINARY := openshannon
INSTALL_DIR := $(HOME)/go/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
WORKSPACE := $(HOME)/OpenShannon
CONFIG_DIR := $(HOME)/.config/openshannon

.PHONY: build install setup test vet run clean start stop restart logs status workspace

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(INSTALL_DIR)/$(BINARY) ./cmd/openshannon

test:
	go test -race ./...

vet:
	go vet ./...

cover:
	go test -race -cover ./...

run:
	go run ./cmd/openshannon

# Full one-click setup: config + workspace + service
setup: build config workspace service
	@echo ""
	@echo "============================================"
	@echo "  openshannon setup complete!"
	@echo "============================================"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit $(CONFIG_DIR)/config.yaml"
	@echo "     - Set your Telegram user ID in allowed_users"
	@echo ""
	@echo "  2. Edit $(CONFIG_DIR)/env"
	@echo "     - TELEGRAM_BOT_TOKEN=your_bot_token"
	@echo "     - (optional) GEMINI_API_KEY=your_key"
	@echo "     - (optional) GOG_KEYRING_PASSWORD=your_password"
	@echo "     - (optional) GOG_ACCOUNT=your@gmail.com"
	@echo ""
	@echo "  3. Start the bot:"
	@echo "     make start"
	@echo ""

# Create config files (idempotent)
config:
	@mkdir -p $(CONFIG_DIR)
	@chmod 700 $(CONFIG_DIR)
	@if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		cp config.example.yaml $(CONFIG_DIR)/config.yaml; \
		chmod 600 $(CONFIG_DIR)/config.yaml; \
		echo "Created $(CONFIG_DIR)/config.yaml"; \
	else \
		echo "Config exists: $(CONFIG_DIR)/config.yaml (skipped)"; \
	fi
	@if [ ! -f $(CONFIG_DIR)/env ]; then \
		printf "TELEGRAM_BOT_TOKEN=\nGEMINI_API_KEY=\nGOG_KEYRING_PASSWORD=\nGOG_ACCOUNT=\n" > $(CONFIG_DIR)/env; \
		chmod 600 $(CONFIG_DIR)/env; \
		echo "Created $(CONFIG_DIR)/env — fill in your secrets"; \
	else \
		echo "Env exists: $(CONFIG_DIR)/env (skipped)"; \
	fi

# Create workspace directory with CLAUDE.md
workspace:
	@mkdir -p $(WORKSPACE)
	@if [ ! -d $(WORKSPACE)/.git ]; then \
		cd $(WORKSPACE) && git init && \
		echo "# OpenShannon Workspace" > README.md && \
		git add -A && git commit -m "init: workspace for OpenShannon"; \
		echo "Initialized git repo: $(WORKSPACE)"; \
	fi
	@if [ ! -f $(WORKSPACE)/CLAUDE.md ]; then \
		if [ -f $(CONFIG_DIR)/env ]; then \
			. $(CONFIG_DIR)/env 2>/dev/null; \
			GOG_ACCOUNT=$${GOG_ACCOUNT:-your@gmail.com}; \
			sed "s|\$$GOG_ACCOUNT|$$GOG_ACCOUNT|g" workspace-claude.md.template > $(WORKSPACE)/CLAUDE.md; \
		else \
			sed "s|\$$GOG_ACCOUNT|your@gmail.com|g" workspace-claude.md.template > $(WORKSPACE)/CLAUDE.md; \
		fi; \
		echo "Created $(WORKSPACE)/CLAUDE.md"; \
	else \
		echo "CLAUDE.md exists: $(WORKSPACE)/CLAUDE.md (skipped)"; \
	fi

# Install systemd service
service:
	@mkdir -p $(HOME)/.config/systemd/user
	@cp openshannon.service $(HOME)/.config/systemd/user/
	@systemctl --user daemon-reload
	@echo "Systemd service installed"

# Legacy install target (same as setup)
install: setup

start:
	systemctl --user enable --now openshannon
	@loginctl enable-linger $(shell whoami) 2>/dev/null || true
	@echo "Service started + linger enabled"

stop:
	systemctl --user stop openshannon

restart: build
	systemctl --user restart openshannon

logs:
	journalctl --user -u openshannon -f

status:
	systemctl --user status openshannon

clean:
	rm -f $(INSTALL_DIR)/$(BINARY)

# Interactive setup wizard (recommended for first-time users)
wizard:
	@bash install.sh

# Setup gog (Google services) separately
setup-gog:
	@echo "Setting up gog (Google services)..."
	@echo "You need gog CLI installed: go install github.com/AarynSmith/gog@latest"
	@command -v gog >/dev/null 2>&1 || { echo "gog not found. Install it first."; exit 1; }
	@echo ""
	@echo "Choose a keyring password (encrypts your Google OAuth token):"
	@read -p "Password: " GOG_PW; \
	echo "Google account email:"; \
	read -p "Email: " GOG_EMAIL; \
	GOG_KEYRING_PASSWORD="$$GOG_PW" gog auth add "$$GOG_EMAIL"; \
	echo ""; \
	echo "Testing..."; \
	GOG_KEYRING_PASSWORD="$$GOG_PW" gog gmail search "newer_than:1d" --account "$$GOG_EMAIL" --plain --no-input 2>/dev/null | head -3; \
	echo ""; \
	echo "Add to $(CONFIG_DIR)/env:"; \
	echo "  GOG_KEYRING_PASSWORD=$$GOG_PW"; \
	echo "  GOG_ACCOUNT=$$GOG_EMAIL"; \
	echo "Then: make restart"

# Update workspace CLAUDE.md from template (overwrites existing)
update-workspace:
	@if [ -f $(CONFIG_DIR)/env ]; then \
		. $(CONFIG_DIR)/env 2>/dev/null; \
		GOG_ACCOUNT=$${GOG_ACCOUNT:-your@gmail.com}; \
		sed "s|\$$GOG_ACCOUNT|$$GOG_ACCOUNT|g" workspace-claude.md.template > $(WORKSPACE)/CLAUDE.md; \
	else \
		sed "s|\$$GOG_ACCOUNT|your@gmail.com|g" workspace-claude.md.template > $(WORKSPACE)/CLAUDE.md; \
	fi
	@echo "Updated $(WORKSPACE)/CLAUDE.md"
