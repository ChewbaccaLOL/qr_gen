# Python CLI (legacy)

This project originally shipped a Python CLI + GUI. The Python implementation lives in `legacy/python/` for reference and compatibility, but the mainline workflow is now Go.

## Requirements
- Python 3.8+
- `segno` (`pip install segno`)
- Optional for PNG/PDF/PS export: `cairosvg` (`pip install cairosvg`)
- Optional for GIF export: `cairosvg` + `Pillow` (`pip install cairosvg pillow`)
- Optional for the Qt GUI: `PySide6` (`pip install pyside6`)

## Install from source
Core CLI only:
```bash
python3 -m pip install --upgrade pip
python3 -m pip install -r requirements.txt
```

Optional extras (PNG/PDF/PS, GIF, Qt GUI):
```bash
python3 -m pip install -r requirements-optional.txt
```

Dev/test tooling:
```bash
python3 -m pip install -r requirements-dev.txt
```

You can also use the minimal Makefile or the setup script:
```bash
make install
make install-optional
make install-dev
./setup.sh --optional
./setup.sh --dev
./setup.sh --all
```

## Binaries (Releases)
Prebuilt CLI + GUI binaries for Windows, macOS, and Linux are attached to GitHub Releases.
- The release binaries include the core SVG generator (`segno`) and the Qt GUI runtime.
- PNG/PDF/PS and GIF exports require extra dependencies (`cairosvg`, `Pillow`) and are not bundled by default.
  Build from source or make your own release if you need those features in the binary.

## Compile from source (binaries)
Install PyInstaller and build a onefile CLI executable:
```bash
python3 -m pip install --upgrade pip
python3 -m pip install segno pyinstaller
pyinstaller --onefile --name qr-generator legacy/python/python/qr_generator.py
```

Artifacts land in `dist/`:
- macOS/Linux: `dist/qr-generator`
- Windows: `dist/qr-generator.exe`

If you prefer easier debugging (or a faster startup), switch to `--onedir` instead of `--onefile`.

### Build GUI (Qt) locally
```bash
python3 -m pip install --upgrade pip
python3 -m pip install segno pyinstaller pyside6
pyinstaller --onefile --windowed --name qr-generator legacy/python/python/qr_gui.py --collect-all PySide6 --hidden-import PySide6.QtSvg
```

### Windows GUI build (with Cairo bundled)
On Windows, previews/PNG require Cairo DLLs. Use the helper script to bundle them:
```powershell
choco install -y gtk-runtime msys2
python -m pip install --upgrade pip
python -m pip install pyinstaller segno cairosvg pyside6
.\scripts\build_windows.ps1
```
The script outputs `dist/qr-generator.exe` and a `dist/cairo/` folder; keep that folder next to the EXE when distributing.

## Usage
```bash
python3 legacy/python/python/qr_generator.py "https://example.com" -o out/qr.svg
python3 legacy/python/python/qr_generator.py "hello" --variant rounded -o rounded.svg
python3 legacy/python/python/qr_generator.py --variant neon --scale 12 --border 3 "Designer ready"
python3 legacy/python/python/qr_generator.py --png "Preview me"
python3 legacy/python/python/qr_generator.py --png --png-scale 4 "Print-ready preview"
python3 legacy/python/python/qr_generator.py --gif "Wave me"
python3 legacy/python/python/qr_generator.py --animation --animation-variant wave "Wave me"
python3 legacy/python/python/qr_generator.py --animation --animation-variant wave-loop "Always waving"
python3 legacy/python/python/qr_generator.py --animation --animation-variant float "Smooth float"
python3 legacy/python/python/qr_generator.py --animation --animation-variant float-tilt-first "Vertical float"
python3 legacy/python/python/qr_generator.py --animation --animation-variant float-tilt-still "Tilted positions"
python3 legacy/python/python/qr_generator.py --animation --animation-variant float-jagged "Retro float"
python3 legacy/python/python/qr_generator.py --gif --readable-gif "Safer wave"
python3 legacy/python/python/qr_generator.py --pdf "Photoshop friendly"
```

You can also pipe data:
```bash
echo "https://example.com" | python3 legacy/python/python/qr_generator.py -o piped.svg
```

## GUI (experimental)
Launch the Qt GUI wrapper (requires PySide6):
```bash
python3 legacy/python/python/qr_gui.py
```

Notes:
- Live previews are built-in (no Cairo dependency).
- PNG/PDF/PS export requires `cairosvg`.
- The copy button copies a PNG preview to your clipboard.
- Presets saved in the GUI are stored in `qr_presets.json` (auto-loaded on launch).
- The right-side controls are scrollable if your window is small.

Tkinter GUI (legacy):
```bash
python3 legacy/python/python/legacy/qr_gui_tk.py
```
