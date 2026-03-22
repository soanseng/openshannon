BINARY := claude-channels
INSTALL_DIR := $(HOME)/go/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
WORKSPACE := $(HOME)/OpenShannon
CONFIG_DIR := $(HOME)/.config/claude-channels

.PHONY: build install setup test vet run clean start stop restart logs status workspace

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(INSTALL_DIR)/$(BINARY) ./cmd/claude-channels

test:
	go test -race ./...

vet:
	go vet ./...

cover:
	go test -race -cover ./...

run:
	go run ./cmd/claude-channels

# Full one-click setup: config + workspace + service
setup: build config workspace service
	@echo ""
	@echo "============================================"
	@echo "  claude-channels setup complete!"
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
		git add -A && git commit -m "init: workspace for Claude Channels"; \
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
	@cp claude-channels.service $(HOME)/.config/systemd/user/
	@systemctl --user daemon-reload
	@echo "Systemd service installed"

# Legacy install target (same as setup)
install: setup

start:
	systemctl --user enable --now claude-channels
	@loginctl enable-linger $(shell whoami) 2>/dev/null || true
	@echo "Service started + linger enabled"

stop:
	systemctl --user stop claude-channels

restart: build
	systemctl --user restart claude-channels

logs:
	journalctl --user -u claude-channels -f

status:
	systemctl --user status claude-channels

clean:
	rm -f $(INSTALL_DIR)/$(BINARY)

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
