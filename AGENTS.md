# Agent Notes

## Project intent
- Build a flexible CLI QR code generator for a designer-friendly workflow.
- Legacy Python lives in the `legacy/python/` submodule but is now obsolete; no need to keep changes in sync unless explicitly requested.
- GUI is out of scope for now (but leave room to add later).
- When a GUI is added, it should be a thin wrapper that calls the same core logic as the CLI.

## How to work here
- Avoid touching `legacy/python/` unless explicitly asked; focus on the Go workflow at repo root.
- When you add or change variants, update `variants.json` and the variants list in `README.md`.
- Always add or update unit tests to cover new or changed code.
- Keep defaults scan-safe: high contrast, sensible quiet zone, and error correction.
- The catalog output should include every variant with a readable label.
- PNG export depends on `cairosvg`; keep it optional and fail fast with a clear message.
- GIF export depends on `cairosvg` + `Pillow`; keep it optional and fail fast with a clear message.
- Default output goes to `out/` and that directory is gitignored.
- `.env` is supported in the project root; command-line arguments always take precedence.
- Run relevant tests while developing changes; for Go use `go test ./...`.

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
- `QR_FLOAT_ANGLE` float float drift angle in degrees
- `QR_READABLE_GIF` set `true` to prefer scan-safer wave defaults
- `QR_VARIANT` variant name
- `QR_SCALE` integer module size
- `QR_BORDER` integer quiet zone
- `QR_ERROR` one of `l`, `m`, `q`, `h`
- `QR_DARK` foreground color
- `QR_LIGHT` background color
- `QR_NO_BACKGROUND` set `true` for transparent background
- `QR_CUTOUT` set `true` for background cutout mode
- `QR_RADIUS` float for rounded modules
- `QR_GRADIENT` set `true`/`false` to enable foreground gradients
- `QR_NO_GRADIENT` set `true` to disable foreground gradients
- `QR_GRADIENT_FROM` gradient start color
- `QR_GRADIENT_TO` gradient end color
- `QR_GRADIENT_ANGLE` gradient direction in degrees
- `QR_GRADIENT_FROM_STOP` gradient start stop (0-1)
- `QR_GRADIENT_TO_STOP` gradient end stop (0-1)
- `QR_GRADIENT_SCOPE` gradient scope (`module` or `global`)
- `QR_BG_GRADIENT` set `true`/`false` to enable background gradients
- `QR_NO_BG_GRADIENT` set `true` to disable background gradients
- `QR_BG_GRADIENT_FROM` background gradient start color
- `QR_BG_GRADIENT_TO` background gradient end color
- `QR_BG_GRADIENT_ANGLE` background gradient direction in degrees
- `QR_BG_GRADIENT_FROM_STOP` background gradient start stop (0-1)
- `QR_BG_GRADIENT_TO_STOP` background gradient end stop (0-1)
- `QR_CATALOG` set `true` to generate the catalog grid
- `QR_CATALOG_COLUMNS` integer column count
- `QR_CATALOG_BACKGROUND` catalog canvas background
- `QR_CATALOG_LABEL_SIZE` label font size (0 = auto)

## Quick commands
```bash
go run ./cmd/qr_generator --help
go run ./cmd/qr_generator --list-variants
go run ./cmd/qr_generator --catalog --png
go run ./cmd/qr_generator --gif "Wave me"
go run ./cmd/qr_generator --animation --animation-variant wave "Wave me"
```

## Style goals
- Provide at least a few standard variants (black/white, square, rounded).
- Add more playful variants (color, gradients, or different shapes) but keep scanability in mind.

## Go migration plan (draft)
- Keep `variants.json` as the shared source of truth for variants + animation defaults.
- Python remains the reference CLI; Go CLI should be flag- and env-compatible before replacing anything.
- Extract shared logic into config + tests first; only then port rendering/animation routines.
- Plan for a thin GUI wrapper (Qt or similar) that calls the CLI/library without duplicating logic.
