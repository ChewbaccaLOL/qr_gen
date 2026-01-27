# QR Generator

A small CLI tool to generate designer-friendly QR codes as SVG. It focuses on flexible styling so your designer friend can pick a variant and tweak colors without touching the QR logic.

## Why
- Fast CLI workflow today, room for a GUI later.
- A handful of standard styles plus playful, colorful variants.
- SVG output for easy editing in design tools.

## Requirements
- Python 3.8+
- `segno` (`pip install segno`)
- Optional for PNG export: `cairosvg` (`pip install cairosvg`)

## Usage
```bash
python3 qr_generator.py "https://example.com" -o qr.svg
python3 qr_generator.py "hello" --variant rounded -o rounded.svg
python3 qr_generator.py --variant neon --scale 12 --border 3 "Designer ready"
python3 qr_generator.py --png "Preview me"
```

You can also pipe data:
```bash
echo "https://example.com" | python3 qr_generator.py -o piped.svg
```

## Variants
Standard:
- `classic` (also `square`): black modules on white background
- `rounded`: softened corners
- `dot`: circular modules

More playful:
- `inverted`: white on black
- `midnight`: pale modules on a deep blue background
- `sunset`: rounded modules with a warm gradient
- `neon`: dot modules with a high-contrast gradient

List all variants:
```bash
python3 qr_generator.py --list-variants
```

## PNG Export
Add `--png` to write a PNG alongside the SVG (requires `cairosvg`):
```bash
python3 qr_generator.py "https://example.com" --png -o qr.svg
```

You can also choose the PNG path:
```bash
python3 qr_generator.py "https://example.com" --png --png-output preview.png
```

## Catalog Grid
Generate a single labeled grid showing all variants:
```bash
python3 qr_generator.py "https://example.com" --catalog
python3 qr_generator.py "https://example.com" --catalog --catalog-columns 4 --png
```

## Common Options
- `--scale`: module size in pixels (default 10)
- `--border`: quiet zone in modules (default 4)
- `--error`: `l`, `m`, `q`, `h` (default `m`)
- `--dark`: override foreground color
- `--light`: override background color
- `--no-background`: transparent background
- `--radius`: rounded corner radius (0-0.5)

## Notes
- Primary output is SVG; add `--png` for a bitmap preview.
- If you override `--dark`, gradient variants fall back to the flat color.
- Catalog output uses each variant's default styling.
- Keep contrast high for scan reliability.
