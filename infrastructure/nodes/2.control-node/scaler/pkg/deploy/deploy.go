package deploy

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"scaler/pkg/config"
	"time"
)

const remoteScript = `
#!/bin/bash
set -euo pipefail

if [ "${DEBUG:-0}" = "1" ]; then
  set -x
fi

# processing args
pluggable_api_repo="$1"
relative_compose_root="$2"
target_sub_dir="$3"
cors_b64="$4"
postgres_b64="$5"
redis_b64="$6"

decode_b64() {
  local input="$1"
  local out
  if out=$(printf "%s" "$input" | base64 --decode 2>/dev/null); then
    printf '%s' "$out"
    return 0
  fi
  if out=$(printf "%s" "$input" | base64 -d 2>/dev/null); then
    printf '%s' "$out"
    return 0
  fi
  if out=$(printf "%s" "$input" | base64 -D 2>/dev/null); then
    printf '%s' "$out"
    return 0
  fi
  printf ''
  return 0
}

cors_origins="$(decode_b64 "${cors_b64}")"
postgres_url="$(decode_b64 "${postgres_b64}")"
redis_url="$(decode_b64 "${redis_b64}")"

case "$relative_compose_root" in
  /*) ;;
  *) relative_compose_root="/${relative_compose_root}" ;;
esac

case "$target_sub_dir" in
  /*) ;;
  *) target_sub_dir="/${target_sub_dir}" ;;
esac

if ! readlink -f / >/dev/null 2>&1; then
	echo "[ERROR] readlink -f not available on remote host." >&2
	exit 1
fi

user_home=$(readlink -f "$HOME")
deploy_dir=$(readlink -f "$user_home$target_sub_dir")

# never operate outside home directory
for d in "$deploy_dir"; do
  case "$d" in
    "$user_home"/*) ;;
    *)
      echo "[ERROR] Directory outside user home: $d" >&2
      exit 2
      ;;
  esac
done

# refuse dangerous or empty deploy_dir
if [ -z "$deploy_dir" ] || [ "$deploy_dir" = "/" ] || [ "$deploy_dir" = "$user_home" ]; then
    echo "[ERROR] Refusing to operate on unsafe deploy_dir: '$deploy_dir'" >&2
    exit 2
fi

# git setup
if ! command -v git >/dev/null 2>&1; then
    echo "[INFO] git not found, attempting installation..."
    if command -v apt-get >/dev/null 2>&1; then
        sudo apt-get update -y || { echo "[ERROR] apt-get update failed"; exit 3; }
		    sudo apt-get install -y git || { echo "[ERROR] git installation failed"; exit 3; }
    else
        echo "[WARN] No supported package manager to install git. Proceeding only if git is available." >&2
        if ! command -v git >/dev/null 2>&1; then
            echo "[ERROR] git not found and cannot be installed."
		    exit 3
        fi
    fi
fi

# tmpdir setup
tmpdir=""
if tmpdir=$(mktemp -d /tmp/deploy.XXXXXX 2>/dev/null); then
    :
elif tmpdir=$(mktemp -d 2>/dev/null); then
    :
else
    tmpdir="/tmp/deploy.$RANDOM.$$"
    mkdir -p "$tmpdir"
fi

cleanup() {
  if [ -n "$tmpdir" ] && [[ "$tmpdir" == /tmp/deploy.* ]]; then
    rm -rf "$tmpdir"
  fi
}

trap cleanup EXIT

# clone repository
echo "[INFO] Cloning repository $pluggable_api_repo into $tmpdir/repo"
if ! git clone --depth=1 "$pluggable_api_repo" "$tmpdir/repo"; then
    echo "[ERROR] Repository clone failed."
    exit 4
fi

# prepare safe backup/replacement
echo "[INFO] Preparing $deploy_dir directory (safe replace with backup)"
mkdir -p "$(dirname "$deploy_dir")"

backup_dir="${deploy_dir}.bak.$(date +%s)"

if [ -d "$deploy_dir" ]; then
    echo "[INFO] Backing up existing deploy dir to $backup_dir"
    if mv "$deploy_dir" "$backup_dir" 2>/dev/null; then
        echo "[INFO] Backup created."
    else
        echo "[WARN] mv failed; attempting copy+remove fallback"
        rm -rf "$backup_dir"
        mkdir -p "$backup_dir"
        cp -a "$deploy_dir"/. "$backup_dir"/ || { echo "[WARN] copy fallback failed"; }
        rm -rf "$deploy_dir"
    fi
fi

# move new repo into place
if mv "$tmpdir/repo" "$deploy_dir" 2>/dev/null; then
    echo "[INFO] Repository moved into place."
else
    echo "[INFO] Fallback safe move"
    rm -rf "$deploy_dir"
    mv "$tmpdir/repo" "$deploy_dir"
fi

compose_dir=$(readlink -f "$deploy_dir$relative_compose_root")

# never operate outside home directory
for d in "$compose_dir"; do
  case "$d" in
    "$user_home"/*) ;;
    *)
      echo "[ERROR] Directory outside user home: $d" >&2
      exit 2
      ;;
  esac
done

# validate compose file location
compose_file="$compose_dir/docker-compose.yml"
if [ ! -f "$compose_file" ]; then
    echo "[ERROR] docker-compose.yml not found at $compose_file"
    if [ -d "$backup_dir" ]; then
        echo "[INFO] Restoring previous version from backup..."
        rm -rf "$deploy_dir"
        mv "$backup_dir" "$deploy_dir"
    fi
    exit 6
fi

# prepare .env file safely (treat values strictly as data)
env_file="$compose_dir/.env"
mkdir -p "$(dirname "$env_file")"

# sanitize values
sanitize_value() {
  printf "%s" "$1" | tr -d '\r\n'
}

cors_line=$(sanitize_value "$cors_origins")
pg_line=$(sanitize_value "$postgres_url")
redis_line=$(sanitize_value "$redis_url")

{
  printf '%s\n' "CORS_ORIGINS=${cors_line}"
  printf '%s\n' "POSTGRES_URL=${pg_line}"
  printf '%s\n' "REDIS_URL=${redis_line}"
} > "$env_file"

chmod 600 "$env_file" || true
echo "[INFO] Wrote $env_file"

# docker setup
if ! command -v docker >/dev/null 2>&1; then
    echo "[INFO] Docker not found, attempting installation..."
    if command -v apt-get >/dev/null 2>&1; then
        export DEBIAN_FRONTEND=noninteractive

        # Add Docker's official repository
        echo "[INFO] Adding Docker official repository..."
        sudo mkdir -p /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg 2>/dev/null || { echo "[ERROR] Failed to add Docker GPG key"; exit 5; }

        ARCH=$(dpkg --print-architecture)
        echo "deb [arch=${ARCH} signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | \
          sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

        sudo -E apt-get update -y || { echo "[ERROR] apt-get update failed"; exit 5; }

        # Install Docker CE (Community Edition) and Docker Compose plugin
        sudo -E apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin || { echo "[ERROR] Docker installation failed"; exit 5; }
        sudo systemctl enable --now docker || { echo "[ERROR] Docker service activation failed"; exit 5; }
    else
        echo "[WARN] Unsupported OS for automated Docker installation." >&2
        if ! command -v docker >/dev/null 2>&1; then
            echo "[ERROR] Docker not available."
            exit 5
        fi
    fi
else
    echo "[INFO] Docker already installed."
fi

# docker compose setup
if ! docker compose version >/dev/null 2>&1; then
    echo "[INFO] Docker Compose not found, attempting installation..."
    if command -v apt-get >/dev/null 2>&1; then
        export DEBIAN_FRONTEND=noninteractive

        # Ensure Docker repository is added (in case Docker was already installed via docker.io)
        sudo mkdir -p /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg 2>/dev/null || true

        ARCH=$(dpkg --print-architecture)
        echo "deb [arch=${ARCH} signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | \
          sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

        sudo -E apt-get update -y || { echo "[ERROR] apt-get update failed"; exit 5; }
        sudo -E apt-get install -y docker-compose-plugin || { echo "[ERROR] Docker Compose installation failed"; exit 5; }
    else
        echo "[WARN] Unsupported OS for automated Docker Compose installation." >&2
        exit 5
    fi
else
    echo "[INFO] Docker Compose already installed."
fi

# start pluggable api
echo "[INFO] Starting services using docker compose..."
if ! sudo docker compose -f "$compose_file" up -d --remove-orphans; then
    echo "[ERROR] docker compose failed to start services." >&2
    if [ -d "$backup_dir" ]; then
        echo "[INFO] Rolling back to previous version..."
        rm -rf "$deploy_dir"
        mv "$backup_dir" "$deploy_dir"
        echo "[INFO] Previous version restored."
    fi
    exit 7
fi

# deployment succeeded: remove backup and show status/logs
if [ -d "$backup_dir" ]; then
    rm -rf "$backup_dir"
fi

echo "[SUCCESS] Deployment completed. Services should be running."

# provide quick health output
echo "[INFO] docker compose ps (services):"
sudo docker compose -f "$compose_file" ps || true

echo "[INFO] recent logs (tail 100):"
sudo docker compose -f "$compose_file" logs --tail=100 || true

exit 0
`

func DeployPluggableAPI(agent config.AgentConfig) error {
	timeout := 60 * time.Minute

	pluggableApiRepo := "https://github.com/xscotophilic/cluster-ops-playground"
	relativeComposePath := "distributed-pluggable-api/compose"
	pluggableApiDir := "pluggable-api"

	cors := os.Getenv("CORS_ORIGINS")
	pg := os.Getenv("POSTGRES_URL")
	redis := os.Getenv("REDIS_URL")

	corsB64 := base64.StdEncoding.EncodeToString([]byte(cors))
	pgB64 := base64.StdEncoding.EncodeToString([]byte(pg))
	redisB64 := base64.StdEncoding.EncodeToString([]byte(redis))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %v", err)
	}

	commandArgs := []string{
		"-p", agent.SSH.Port,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i", fmt.Sprintf("%s/.ssh/id_ed25519", homeDir),
		fmt.Sprintf("%s@%s", agent.SSH.User, agent.SSH.IP),
		"bash", "-s", "--",
		pluggableApiRepo,
		relativeComposePath,
		pluggableApiDir,
		corsB64,
		pgB64,
		redisB64,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ssh", commandArgs...)
	cmd.Stdin = bytes.NewBufferString(remoteScript)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("remote deploy failed: %w\noutput:\n%s", err, string(output))
	}

	fmt.Fprintln(os.Stdout, string(output))
	return nil
}
