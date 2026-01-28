#!/usr/bin/env bash
set -euo pipefail

DATA="${QR_DATA:-https://example.com}"
OUT_DIR="${OUT_DIR:-docs/variants}"

mkdir -p "$OUT_DIR"

QR_GENERATOR="${QR_GENERATOR:-}"
if [[ -n "$QR_GENERATOR" ]]; then
  # Allow overriding with a full command like "go run ./cmd/qr_generator".
  # shellcheck disable=SC2206
  QR_GENERATOR_CMD=($QR_GENERATOR)
else
  # Prefer go run so config + code stay in sync; set QR_GENERATOR to use a local binary.
  QR_GENERATOR_CMD=(go run ./cmd/qr_generator)
fi

readarray -t VARIANT_LINES < <("${QR_GENERATOR_CMD[@]}" --list-variants)
VARIANTS=()
ANIMATIONS=()
MODE="variants"
for line in "${VARIANT_LINES[@]}"; do
  if [[ "$line" == "Animations:" ]]; then
    MODE="animations"
    continue
  fi
  if [[ -z "$line" ]]; then
    continue
  fi
  if [[ "$MODE" == "variants" ]]; then
    VARIANTS+=("$line")
  else
    ANIMATIONS+=("$line")
  fi
done

for variant in "${VARIANTS[@]}"; do
  "${QR_GENERATOR_CMD[@]}" --variant "$variant" -o "$OUT_DIR/${variant}.svg" "$DATA"
done

for animation in "${ANIMATIONS[@]}"; do
  "${QR_GENERATOR_CMD[@]}" --variant classic \
    -o "$OUT_DIR/animation-${animation}.svg" \
    --animation --animation-variant "$animation" \
    "$DATA"
done
