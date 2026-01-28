param(
  [switch]$NoPacman
)

$ErrorActionPreference = "Stop"

function Find-CairoBin {
  $candidatePaths = @(
    "C:\Program Files\GTK3-Runtime Win64\bin",
    "C:\Program Files\GTK3-Runtime\bin",
    "C:\gtk\bin",
    "$env:ChocolateyInstall\lib\gtk-runtime\tools\gtk3-runtime\bin",
    "$env:ChocolateyInstall\lib\gtk-runtime\tools\bin"
  )

  foreach ($path in $candidatePaths) {
    if (Test-Path $path) {
      $dll = Join-Path $path "libcairo-2.dll"
      if (Test-Path $dll) {
        return $path
      }
    }
  }

  $searchRoot = "$env:ChocolateyInstall\lib\gtk-runtime\tools"
  if (Test-Path $searchRoot) {
    $found = Get-ChildItem $searchRoot -Recurse -Filter "libcairo-2.dll" -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($found) {
      return $found.DirectoryName
    }
  }

  $msys = "C:\msys64"
  if ((Test-Path $msys) -and (-not $NoPacman)) {
    & "$msys\usr\bin\bash.exe" -lc "pacman -Sy --noconfirm" | Out-Host
    & "$msys\usr\bin\bash.exe" -lc "pacman -S --noconfirm mingw-w64-x86_64-cairo mingw-w64-x86_64-pango mingw-w64-x86_64-gdk-pixbuf2 || pacman -S --noconfirm mingw-w64-x86_64-gdk-pixbuf" | Out-Host
  }
  $msysBin = "$msys\mingw64\bin"
  if (Test-Path $msysBin) {
    $dll = Join-Path $msysBin "libcairo-2.dll"
    if (Test-Path $dll) {
      return $msysBin
    }
  }

  return $null
}

$gtk = Find-CairoBin
if (-not $gtk) {
  Write-Error "Cairo DLLs not found. Install gtk-runtime (choco) or msys2 and ensure libcairo-2.dll exists."
  exit 1
}

$dlls = Get-ChildItem $gtk -Filter *.dll
$args = @()
foreach ($dll in $dlls) {
  $args += "--add-binary"
  $args += "$($dll.FullName);cairo"
}

Write-Host "Bundling Cairo DLLs from: $gtk"

pyinstaller --onefile --name qr-generator-cli python/qr_generator.py @args
pyinstaller --onefile --noconsole --name qr-generator python/qr_gui.py --collect-all PySide6 --hidden-import PySide6.QtSvg @args

New-Item -ItemType Directory -Path dist\\cairo -Force | Out-Null
foreach ($dll in $dlls) {
  Copy-Item $dll.FullName dist\\cairo -Force
}
