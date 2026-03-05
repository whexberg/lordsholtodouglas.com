#!/usr/bin/env bash
set -euo pipefail

# ------------------------------------------------------------------
# Deploy LSD3 to production via rsync + SSH
# Usage: ./deploy.sh [--first-run]
# ------------------------------------------------------------------

SSH_HOST="lsd3"
REMOTE_DIR="/srv/lordsholtodouglas"
LOCAL_DIR="$(cd "$(dirname "$0")" && pwd)"

# Files/dirs to sync (everything Docker needs to build and run)
INCLUDES=(
  cmd/
  content/
  deploy/
  internal/
  static/
  templates/
  go.mod
  go.sum
  Taskfile.yml
)

# Bold text helpers
bold() { printf '\033[1m%s\033[0m\n' "$1"; }
green() { printf '\033[1;32m%s\033[0m\n' "$1"; }
red() { printf '\033[1;31m%s\033[0m\n' "$1"; }

# ------------------------------------------------------------------
# First-run setup (creates dirs, copies .env, syncs images)
# ------------------------------------------------------------------
first_run() {
  bold "=== First-run setup ==="

  echo "Creating remote directory..."
  ssh "$SSH_HOST" "mkdir -p $REMOTE_DIR"

  # Copy .env if it exists locally
  if [ -f "$LOCAL_DIR/.env" ]; then
    echo "Copying .env to server..."
    rsync -av "$LOCAL_DIR/.env" "$SSH_HOST:$REMOTE_DIR/.env"
    echo "IMPORTANT: Edit $REMOTE_DIR/.env on the server with production values"
    echo "  - CLOVER_BASE_URL=https://api.clover.com"
    echo "  - CLOVER_ECOMMERCE_URL=https://scl.clover.com"
    echo "  - DOMAIN=yourdomain.com"
    echo "  - ACME_EMAIL=you@example.com"
  else
    echo "No local .env found. Copy .env.example to the server and configure it:"
    rsync -av "$LOCAL_DIR/.env.example" "$SSH_HOST:$REMOTE_DIR/.env.example"
  fi

  # Sync images (large, only on first run or when changed)
  if [ -d "$LOCAL_DIR/static/images" ]; then
    echo "Syncing images..."
    ssh "$SSH_HOST" "mkdir -p $REMOTE_DIR/static/images"
    rsync -avz --delete "$LOCAL_DIR/static/images/" "$SSH_HOST:$REMOTE_DIR/static/images/"
  fi

  green "First-run setup complete."
  echo "Next steps:"
  echo "  1. SSH in and edit $REMOTE_DIR/.env with production credentials"
  echo "  2. Run ./deploy.sh again (without --first-run) to deploy"
}

# ------------------------------------------------------------------
# Build rsync include/exclude args
# ------------------------------------------------------------------
build_rsync_args() {
  local args=()
  for item in "${INCLUDES[@]}"; do
    args+=(--include="$item" --include="$item**")
  done
  # Exclude everything else at the root level
  args+=(--exclude="*")
  echo "${args[@]}"
}

# ------------------------------------------------------------------
# Main deploy
# ------------------------------------------------------------------
deploy() {
  bold "=== Deploying LSD3 ==="

  # Verify .env exists on server
  if ! ssh "$SSH_HOST" "test -f $REMOTE_DIR/.env"; then
    red "ERROR: No .env file on server. Run './deploy.sh --first-run' first."
    exit 1
  fi

  # Sync source files
  echo "Syncing files..."
  rsync -avz --delete \
    $(build_rsync_args) \
    "$LOCAL_DIR/" "$SSH_HOST:$REMOTE_DIR/"

  # Build and restart on server
  echo "Building and restarting containers..."
  ssh "$SSH_HOST" "cd $REMOTE_DIR/deploy && docker compose -f docker-compose.prod.yml up --build -d"

  # Wait for health check
  echo "Waiting for health check..."
  sleep 5
  if ssh "$SSH_HOST" "curl -sf http://localhost:8080/health > /dev/null 2>&1"; then
    green "Deploy successful! Site is healthy."
  else
    echo "Health check didn't pass yet. Checking container status..."
    ssh "$SSH_HOST" "cd $REMOTE_DIR/deploy && docker compose -f docker-compose.prod.yml ps"
    echo ""
    echo "Check logs with: ssh $SSH_HOST 'cd $REMOTE_DIR/deploy && docker compose -f docker-compose.prod.yml logs'"
  fi
}

# ------------------------------------------------------------------
# Helpers
# ------------------------------------------------------------------
sync_images() {
  bold "=== Syncing images ==="
  if [ -d "$LOCAL_DIR/static/images" ]; then
    rsync -avz --delete "$LOCAL_DIR/static/images/" "$SSH_HOST:$REMOTE_DIR/static/images/"
    green "Images synced."
  else
    red "No static/images/ directory found locally."
  fi
}

show_logs() {
  ssh "$SSH_HOST" "cd $REMOTE_DIR/deploy && docker compose -f docker-compose.prod.yml logs -f"
}

show_status() {
  ssh "$SSH_HOST" "cd $REMOTE_DIR/deploy && docker compose -f docker-compose.prod.yml ps"
}

# ------------------------------------------------------------------
# CLI
# ------------------------------------------------------------------
case "${1:-}" in
  --first-run)  first_run ;;
  --images)     sync_images ;;
  --logs)       show_logs ;;
  --status)     show_status ;;
  --help|-h)
    echo "Usage: ./deploy.sh [OPTION]"
    echo ""
    echo "Options:"
    echo "  (none)        Deploy code and restart containers"
    echo "  --first-run   Initial server setup (create dirs, copy .env, sync images)"
    echo "  --images      Sync static/images/ to server"
    echo "  --logs        Tail production logs"
    echo "  --status      Show container status"
    echo "  --help        Show this help"
    ;;
  *)            deploy ;;
esac
