#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

SYSTEM=0
INSTALL_DIR=""
BINARY_NAME="glagent"

while [ "$#" -gt 0 ]; do
    case "$1" in
        --system)
            SYSTEM=1
            shift
            ;;
        --install-dir)
            INSTALL_DIR="${2:-}"
            shift 2
            ;;
        --binary-name)
            BINARY_NAME="${2:-}"
            shift 2
            ;;
        *)
            echo "Unknown argument: $1" >&2
            exit 1
            ;;
    esac
done

set -- go run . setup
if [ "$SYSTEM" -eq 1 ]; then
    set -- "$@" --system
fi
if [ -n "$INSTALL_DIR" ]; then
    set -- "$@" --install-dir "$INSTALL_DIR"
fi
if [ -n "$BINARY_NAME" ]; then
    set -- "$@" --binary-name "$BINARY_NAME"
fi

"$@"
