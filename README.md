# QR Generator

A small CLI tool to generate designer-friendly QR codes as SVG. It focuses on flexible styling so your designer friend can pick a variant and tweak colors without touching the QR logic.

## Variant Gallery (example outputs)
Each QR below encodes `https://example.com` using the default settings for that variant.

<table>
  <tr>
    <td align="center"><strong>classic</strong><br/><img src="docs/variants/classic.svg" width="140"/></td>
    <td align="center"><strong>square</strong><br/><img src="docs/variants/square.svg" width="140"/></td>
    <td align="center"><strong>rounded</strong><br/><img src="docs/variants/rounded.svg" width="140"/></td>
  </tr>
  <tr>
    <td align="center"><strong>dot</strong><br/><img src="docs/variants/dot.svg" width="140"/></td>
    <td align="center"><strong>clear</strong><br/><img src="docs/variants/clear.svg" width="140"/></td>
    <td align="center"><strong>clear-rounded</strong><br/><img src="docs/variants/clear-rounded.svg" width="140"/></td>
  </tr>
  <tr>
    <td align="center"><strong>clear-dot</strong><br/><img src="docs/variants/clear-dot.svg" width="140"/></td>
    <td align="center"><strong>inverted</strong><br/><img src="docs/variants/inverted.svg" width="140"/></td>
    <td align="center"><strong>midnight</strong><br/><img src="docs/variants/midnight.svg" width="140"/></td>
  </tr>
  <tr>
    <td align="center"><strong>sunset</strong><br/><img src="docs/variants/sunset.svg" width="140"/></td>
    <td align="center"><strong>neon</strong><br/><img src="docs/variants/neon.svg" width="140"/></td>
    <td align="center"></td>
  </tr>
</table>

## Why
- Fast CLI workflow today, room for a GUI later.
- A handful of standard styles plus playful, colorful variants.
- SVG output for easy editing in design tools.

## Requirements
- Python 3.8+
- `segno` (`pip install segno`)
- Optional for PNG/PDF/PS export: `cairosvg` (`pip install cairosvg`)

## Usage
```bash
python3 qr_generator.py "https://example.com" -o out/qr.svg
python3 qr_generator.py "hello" --variant rounded -o rounded.svg
python3 qr_generator.py --variant neon --scale 12 --border 3 "Designer ready"
python3 qr_generator.py --png "Preview me"
python3 qr_generator.py --png --png-scale 4 "Print-ready preview"
python3 qr_generator.py --pdf "Photoshop friendly"
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
- `clear`: black modules on transparent background
- `clear-rounded`: rounded modules on transparent background
- `clear-dot`: dot modules on transparent background

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

Increase raster resolution with `--png-scale` (multiplies the SVG pixel size):
```bash
python3 qr_generator.py "https://example.com" --png --png-scale 4
```

You can also choose the PNG path:
```bash
python3 qr_generator.py "https://example.com" --png --png-output preview.png
```

## PDF/PS Export
Export vector formats that Photoshop can open (requires `cairosvg`):
```bash
python3 qr_generator.py "https://example.com" --pdf
python3 qr_generator.py "https://example.com" --ps
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
- `--png-scale`: scale multiplier for PNG export

## Notes
- Primary output is SVG; add `--png` for a bitmap preview.
- `--pdf` and `--ps` are available for Photoshop-friendly vector output.
- Default output goes into `out/` (auto-created).
- If you override `--dark`, gradient variants fall back to the flat color.
- Catalog output uses each variant's default styling.
- Keep contrast high for scan reliability.
