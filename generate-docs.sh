#!/usr/bin/env bash
set -euo pipefail

PYTHON="${PYTHON:-}"
if [[ -z "$PYTHON" ]]; then
  if [[ -x ".venv/bin/python" ]]; then
    PYTHON=".venv/bin/python"
  else
    PYTHON="python3"
  fi
fi

DATA="${QR_DATA:-https://example.com}"
OUT_DIR="${OUT_DIR:-docs/variants}"

mkdir -p "$OUT_DIR"

VARIANTS=$("$PYTHON" - <<'PY'
from qr_generator import VARIANTS
print(" ".join(sorted(VARIANTS.keys())))
PY
)

ANIMATIONS=$("$PYTHON" - <<'PY'
from qr_generator import ANIMATION_VARIANTS
print(" ".join(ANIMATION_VARIANTS))
PY
)

for variant in $VARIANTS; do
  "$PYTHON" qr_generator.py "$DATA" --variant "$variant" -o "$OUT_DIR/${variant}.svg"
done

for animation in $ANIMATIONS; do
  "$PYTHON" qr_generator.py "$DATA" --variant classic \
    -o "$OUT_DIR/animation-${animation}.svg" \
    --animation --animation-variant "$animation"
done
