# Agent Notes

## Project intent
- Build a flexible CLI QR code generator for a designer-friendly workflow.
- Keep the core generation in `qr_generator.py` and output SVG.
- GUI is out of scope for now (but leave room to add later).
- `qr_generator.py` should stay focused on CLI + orchestration; rendering lives in `qr_render.py`.

## How to work here
- Prefer updating `qr_generator.py` directly; keep the CLI stable.
- When you add or change variants, update `VARIANTS` in `qr_generator.py` and the variants list in `README.md`.
- Keep defaults scan-safe: high contrast, sensible quiet zone, and error correction.
- The catalog output should include every variant with a readable label.
- PNG export depends on `cairosvg`; keep it optional and fail fast with a clear message.
- GIF export depends on `cairosvg` + `Pillow`; keep it optional and fail fast with a clear message.
- Default output goes to `out/` and that directory is gitignored.
- `.env` is supported in the project root; command-line arguments always take precedence.

## .env keys
- `QR_DATA` text or URL to encode (used if no CLI arg and no stdin)
- `QR_OUTPUT` output SVG path
- `QR_PNG` set `true`/`false` for PNG export
- `QR_PNG_OUTPUT` output PNG path
- `QR_ANIMATION` set `true`/`false` to enable animation output
- `QR_ANIMATION_FORMAT` animation format (currently `gif`)
- `QR_ANIMATION_VARIANT` animation variant (currently `wave`)
- `QR_GIF` set `true`/`false` for GIF export
- `QR_GIF_OUTPUT` output GIF path
- `QR_GIF_VARIANT` GIF animation variant (currently `wave`)
- `QR_GIF_FPS` integer frames per second
- `QR_GIF_FRAMES` integer wave animation frames
- `QR_GIF_HOLD` integer still frames before/after wave
- `QR_WAVE_AMP` float wave amplitude in modules
- `QR_WAVE_PERIOD` float wave period in columns
- `QR_READABLE_GIF` set `true` to prefer scan-safer wave defaults
- `QR_VARIANT` variant name
- `QR_SCALE` integer module size
- `QR_BORDER` integer quiet zone
- `QR_ERROR` one of `l`, `m`, `q`, `h`
- `QR_DARK` foreground color
- `QR_LIGHT` background color
- `QR_NO_BACKGROUND` set `true` for transparent background
- `QR_RADIUS` float for rounded modules
- `QR_CATALOG` set `true` to generate the catalog grid
- `QR_CATALOG_COLUMNS` integer column count
- `QR_CATALOG_BACKGROUND` catalog canvas background
- `QR_CATALOG_LABEL_SIZE` label font size (0 = auto)

## Quick commands
```bash
python3 qr_generator.py --help
python3 qr_generator.py --list-variants
python3 qr_generator.py --catalog --png
python3 qr_generator.py --gif "Wave me"
python3 qr_generator.py --animation --animation-variant wave "Wave me"
```

## Style goals
- Provide at least a few standard variants (black/white, square, rounded).
- Add more playful variants (color, gradients, or different shapes) but keep scanability in mind.
