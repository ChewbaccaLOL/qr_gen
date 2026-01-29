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
    --gif-output "$OUT_DIR/animation-${animation}.gif" \
    "$DATA"
done

COMBO_NAMES=(
  "combo-transparent-wave"
  "combo-gradient-float"
  "combo-dot-wave-loop"
  "combo-rounded-tilt"
  "combo-bg-gradient-wave"
)

COMBO_ARGS=(
  "--variant clear-dot --no-background --dark #00ff7f --animation --animation-variant wave"
  "--variant sunset --animation --animation-variant float"
  "--variant neon --animation --animation-variant wave-loop"
  "--variant sunset --animation --animation-variant float-tilt-still"
  "--variant classic --bg-gradient --bg-gradient-from #ffb347 --bg-gradient-to #ff00cc --bg-gradient-angle 135 --dark #0b1020 --animation --animation-variant wave"
)

for i in "${!COMBO_NAMES[@]}"; do
  name="${COMBO_NAMES[$i]}"
  read -r -a combo_args <<< "${COMBO_ARGS[$i]}"
  "${QR_GENERATOR_CMD[@]}" \
    -o "$OUT_DIR/${name}.svg" \
    --gif-output "$OUT_DIR/${name}.gif" \
    "${combo_args[@]}" \
    "$DATA"
done

CUTOUT_NAMES=(
  "cutout-classic"
  "cutout-dot"
  "cutout-rounded"
  "cutout-prism"
  "cutout-aurora"
  "cutout-neon"
)

CUTOUT_ARGS=(
  "--variant classic --cutout --bg-gradient --bg-gradient-from #0b1020 --bg-gradient-to #142a46 --bg-gradient-angle 120"
  "--variant dot --cutout --bg-gradient --bg-gradient-from #003b2f --bg-gradient-to #15b089 --bg-gradient-angle 225"
  "--variant sunset --cutout --bg-gradient --bg-gradient-from #2a0a44 --bg-gradient-to #ff7a59 --bg-gradient-angle 140"
  "--variant prism --cutout --bg-gradient --bg-gradient-from #1a0f3b --bg-gradient-to #5b4bff --bg-gradient-angle 200"
  "--variant aurora --cutout --bg-gradient --bg-gradient-from #0b1020 --bg-gradient-to #7de2ff --bg-gradient-angle 320"
  "--variant neon --cutout --bg-gradient --bg-gradient-from #06060a --bg-gradient-to #6bff2e --bg-gradient-angle 60"
)

for i in "${!CUTOUT_NAMES[@]}"; do
  name="${CUTOUT_NAMES[$i]}"
  read -r -a cutout_args <<< "${CUTOUT_ARGS[$i]}"
  "${QR_GENERATOR_CMD[@]}" \
    -o "$OUT_DIR/${name}.svg" \
    "${cutout_args[@]}" \
    "$DATA"
done
