#!/usr/bin/env bash
# Claude Channels — Interactive Setup Script
# Usage: bash install.sh
set -euo pipefail

# ── Colors ────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

CONFIG_DIR="$HOME/.config/claude-channels"
WORKSPACE="$HOME/OpenShannon"
INSTALL_DIR="$HOME/go/bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

info()  { echo -e "${BLUE}→${NC} $*"; }
ok()    { echo -e "${GREEN}✓${NC} $*"; }
warn()  { echo -e "${YELLOW}!${NC} $*"; }
err()   { echo -e "${RED}✗${NC} $*"; }
ask()   { echo -en "${CYAN}?${NC} $* "; }

# ── Header ────────────────────────────────────────────
echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║        Claude Channels — Setup Wizard        ║${NC}"
echo -e "${BOLD}║  Telegram → Claude Code Bridge               ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════════╝${NC}"
echo ""

# ── Step 0: Prerequisites ────────────────────────────
echo -e "${BOLD}[0/6] Checking prerequisites...${NC}"

# Go
if ! command -v go &>/dev/null; then
    err "Go not found. Install Go 1.22+ first: https://go.dev/dl/"
    exit 1
fi
ok "Go $(go version | awk '{print $3}')"

# Claude Code
if ! command -v claude &>/dev/null; then
    err "Claude Code CLI not found. Install: npm install -g @anthropic-ai/claude-code"
    exit 1
fi
ok "Claude Code $(claude --version 2>/dev/null || echo 'installed')"

# ── Step 1: Build ─────────────────────────────────────
echo ""
echo -e "${BOLD}[1/6] Building binary...${NC}"
cd "$SCRIPT_DIR"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
mkdir -p "$INSTALL_DIR"
go build -ldflags "-X main.version=$VERSION" -o "$INSTALL_DIR/claude-channels" ./cmd/claude-channels
ok "Built claude-channels ($VERSION) → $INSTALL_DIR/"

# Check PATH
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    warn "$INSTALL_DIR is not in your PATH"
    warn "Add this to your shell config: export PATH=\"$INSTALL_DIR:\$PATH\""
fi

# ── Step 2: Telegram Bot ─────────────────────────────
echo ""
echo -e "${BOLD}[2/6] Telegram Bot Setup${NC}"
echo ""
mkdir -p "$CONFIG_DIR"
chmod 700 "$CONFIG_DIR"

TELEGRAM_TOKEN=""
TELEGRAM_USER_ID=""

if [ -f "$CONFIG_DIR/env" ]; then
    source "$CONFIG_DIR/env" 2>/dev/null || true
    TELEGRAM_TOKEN="${TELEGRAM_BOT_TOKEN:-}"
fi

if [ -z "$TELEGRAM_TOKEN" ]; then
    echo -e "  ${BOLD}Create a Telegram bot:${NC}"
    echo "  1. Open @BotFather in Telegram"
    echo "  2. Send /newbot and follow prompts"
    echo "  3. Copy the bot token"
    echo "  4. Send /setprivacy → select your bot → Disable"
    echo ""
    ask "Bot token (7123456789:AAH...):"
    read -r TELEGRAM_TOKEN
    if [ -z "$TELEGRAM_TOKEN" ]; then
        err "Bot token is required"
        exit 1
    fi
else
    ok "Bot token found in existing config"
fi

if [ -z "$TELEGRAM_USER_ID" ]; then
    echo ""
    echo -e "  ${BOLD}Get your Telegram user ID:${NC}"
    echo "  Message @userinfobot in Telegram — it replies with your ID"
    echo ""
    ask "Your Telegram user ID (numeric):"
    read -r TELEGRAM_USER_ID
    if [ -z "$TELEGRAM_USER_ID" ]; then
        err "User ID is required"
        exit 1
    fi
else
    ok "User ID found"
fi

ok "Telegram configured"

# ── Step 3: Gemini (optional) ────────────────────────
echo ""
echo -e "${BOLD}[3/6] Gemini Image Generation (optional)${NC}"
echo ""

GEMINI_KEY="${GEMINI_API_KEY:-}"
if [ -z "$GEMINI_KEY" ]; then
    echo "  Gemini enables /imagine command for AI image generation."
    echo "  Get a free API key: https://aistudio.google.com/apikey"
    echo ""
    ask "Gemini API key (or press Enter to skip):"
    read -r GEMINI_KEY
    if [ -n "$GEMINI_KEY" ]; then
        ok "Gemini configured"
    else
        info "Skipped (you can add GEMINI_API_KEY to env later)"
    fi
else
    ok "Gemini API key found in existing config"
fi

# ── Step 4: Google Services / gog (optional) ─────────
echo ""
echo -e "${BOLD}[4/6] Google Services — gog CLI (optional)${NC}"
echo ""

GOG_PASSWORD="${GOG_KEYRING_PASSWORD:-}"
GOG_ACCT="${GOG_ACCOUNT:-}"

if command -v gog &>/dev/null; then
    ok "gog CLI found: $(which gog)"
    echo ""
    echo "  gog provides Gmail, Calendar, Drive, Tasks, Contacts, Sheets, Docs..."
    echo "  It needs a Google account authenticated with OAuth."
    echo ""
    ask "Set up gog? (y/N):"
    read -r SETUP_GOG

    if [[ "$SETUP_GOG" =~ ^[Yy] ]]; then
        echo ""
        echo -e "  ${BOLD}Step 1: Choose a keyring password${NC}"
        echo "  This encrypts your Google OAuth token on disk."
        echo "  Pick something memorable (or press Enter for a random one)."
        echo ""
        ask "Keyring password:"
        read -r GOG_PASSWORD
        if [ -z "$GOG_PASSWORD" ]; then
            GOG_PASSWORD=$(openssl rand -base64 32 2>/dev/null || head -c 32 /dev/urandom | base64)
            info "Generated random password (saved in env file)"
        fi

        echo ""
        echo -e "  ${BOLD}Step 2: Authenticate your Google account${NC}"
        ask "Google account email:"
        read -r GOG_ACCT
        if [ -z "$GOG_ACCT" ]; then
            warn "Skipped gog setup"
        else
            echo ""
            info "Starting gog OAuth flow..."
            info "A URL will appear — open it in your browser to authorize."
            echo ""
            GOG_KEYRING_PASSWORD="$GOG_PASSWORD" gog auth add "$GOG_ACCT" || true
            echo ""

            # Test
            info "Testing Gmail access..."
            if GOG_KEYRING_PASSWORD="$GOG_PASSWORD" gog gmail search "newer_than:1d" --account "$GOG_ACCT" --plain --no-input 2>/dev/null | head -3; then
                ok "Gmail access working!"
            else
                warn "Gmail test failed — you may need to re-authenticate later"
            fi
        fi
    else
        info "Skipped (you can run 'make setup-gog' later)"
    fi
else
    info "gog CLI not installed. Google services won't be available."
    echo "  Install: go install github.com/AarynSmith/gog@latest"
    echo "  Or: https://github.com/AarynSmith/gog"
fi

# ── Step 5: Write config files ───────────────────────
echo ""
echo -e "${BOLD}[5/6] Writing configuration...${NC}"

# env file
cat > "$CONFIG_DIR/env" << EOF
TELEGRAM_BOT_TOKEN=$TELEGRAM_TOKEN
GEMINI_API_KEY=$GEMINI_KEY
GOG_KEYRING_PASSWORD=$GOG_PASSWORD
GOG_ACCOUNT=$GOG_ACCT
EOF
chmod 600 "$CONFIG_DIR/env"
ok "Secrets → $CONFIG_DIR/env (mode 600)"

# config.yaml — set the user ID
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    sed "s/123456789/$TELEGRAM_USER_ID/" "$SCRIPT_DIR/config.example.yaml" > "$CONFIG_DIR/config.yaml"
    chmod 600 "$CONFIG_DIR/config.yaml"
    ok "Config → $CONFIG_DIR/config.yaml (mode 600)"
else
    # Update user ID in existing config
    sed -i "s/- 123456789/- $TELEGRAM_USER_ID/" "$CONFIG_DIR/config.yaml" 2>/dev/null || true
    ok "Config updated with your user ID"
fi

# ── Step 6: Workspace + systemd ──────────────────────
echo ""
echo -e "${BOLD}[6/6] Setting up workspace and service...${NC}"

# Workspace
mkdir -p "$WORKSPACE"
if [ ! -d "$WORKSPACE/.git" ]; then
    cd "$WORKSPACE" && git init -q
    echo "# OpenShannon Workspace" > README.md
    git add -A && git commit -q -m "init: workspace for Claude Channels"
    ok "Workspace git repo → $WORKSPACE"
fi

# CLAUDE.md
cd "$SCRIPT_DIR"
GOG_ACCOUNT="${GOG_ACCT:-your@gmail.com}"
sed "s|\$GOG_ACCOUNT|$GOG_ACCOUNT|g" workspace-claude.md.template > "$WORKSPACE/CLAUDE.md"
ok "CLAUDE.md → $WORKSPACE/CLAUDE.md"

# systemd
mkdir -p "$HOME/.config/systemd/user"
cp "$SCRIPT_DIR/claude-channels.service" "$HOME/.config/systemd/user/"
systemctl --user daemon-reload
loginctl enable-linger "$(whoami)" 2>/dev/null || true
ok "Systemd service installed (linger enabled)"

# ── Done ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║          Setup Complete!                      ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${GREEN}Start the bot:${NC}"
echo "    make start"
echo ""
echo -e "  ${GREEN}Check status:${NC}"
echo "    make status"
echo "    make logs"
echo ""
echo -e "  ${GREEN}In Telegram:${NC}"
echo "    1. Add the bot to a group (make it admin)"
echo "    2. Enable Topics in group settings"
echo "    3. Send /help to see all commands"
echo ""
if [ -n "$GOG_ACCT" ]; then
    echo -e "  ${GREEN}Google services ready:${NC}"
    echo "    /gog gmail search newer_than:1d"
    echo "    /gog calendar events"
    echo "    /gog drive ls"
    echo ""
fi
if [ -n "$GEMINI_KEY" ]; then
    echo -e "  ${GREEN}Image generation ready:${NC}"
    echo "    /imagine a sunset over the ocean"
    echo ""
fi
