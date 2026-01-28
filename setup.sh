#!/usr/bin/env bash
set -euo pipefail

python3 -m pip install --upgrade pip
python3 -m pip install -r requirements.txt

for arg in "$@"; do
  case "$arg" in
    --optional)
      python3 -m pip install -r requirements-optional.txt
      ;;
    --dev)
      python3 -m pip install -r requirements-dev.txt
      ;;
    --all)
      python3 -m pip install -r requirements-optional.txt
      python3 -m pip install -r requirements-dev.txt
      ;;
    *)
      echo "Unknown option: $arg" >&2
      echo "Usage: ./setup.sh [--optional] [--dev] [--all]" >&2
      exit 1
      ;;
  esac
 done
