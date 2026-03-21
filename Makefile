BINARY := claude-channels
INSTALL_DIR := $(HOME)/go/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build install test vet run clean restart logs status

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

install: build
	mkdir -p ~/.config/claude-channels
	@if [ ! -f ~/.config/claude-channels/config.yaml ]; then \
		cp config.example.yaml ~/.config/claude-channels/config.yaml; \
		echo "Created config.yaml — edit it with your Telegram user ID"; \
	fi
	@if [ ! -f ~/.config/claude-channels/env ]; then \
		echo "TELEGRAM_BOT_TOKEN=" > ~/.config/claude-channels/env; \
		chmod 600 ~/.config/claude-channels/env; \
		echo "Created env file — add your bot token"; \
	fi
	cp claude-channels.service ~/.config/systemd/user/
	systemctl --user daemon-reload
	@echo "Run 'make start' to start the service"

start:
	systemctl --user enable --now claude-channels

stop:
	systemctl --user stop claude-channels

restart:
	systemctl --user restart claude-channels

logs:
	journalctl --user -u claude-channels -f

status:
	systemctl --user status claude-channels

clean:
	rm -f $(INSTALL_DIR)/$(BINARY)
