import "./style.css";
import {
  GenerateSVG,
  GetVariantCatalog,
  GetVariantsPath,
  SavePNG,
  SaveSVG,
  SuggestSavePath
} from "../wailsjs/go/main/App";

const elements = {
  data: document.getElementById("data"),
  variant: document.getElementById("variant"),
  error: document.getElementById("error"),
  scale: document.getElementById("scale"),
  border: document.getElementById("border"),
  dark: document.getElementById("dark"),
  light: document.getElementById("light"),
  radius: document.getElementById("radius"),
  transparent: document.getElementById("transparent"),
  pngScale: document.getElementById("pngScale"),
  refresh: document.getElementById("refresh"),
  exportSvg: document.getElementById("exportSvg"),
  exportPng: document.getElementById("exportPng"),
  preview: document.getElementById("preview"),
  status: document.getElementById("status"),
  variantsPath: document.getElementById("variantsPath")
};

const state = {
  variants: [],
  variantMap: new Map(),
  debounceTimer: null
};

const defaultData = "https://example.com";

elements.data.value = defaultData;

function parseOptionalFloat(value) {
  const trimmed = value.trim();
  if (!trimmed) {
    return null;
  }
  const parsed = Number.parseFloat(trimmed);
  return Number.isFinite(parsed) ? parsed : null;
}

function parseIntOr(value, fallback) {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function parseFloatOr(value, fallback) {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function setStatus(message, type = "") {
  elements.status.textContent = message;
  elements.status.className = type ? `status ${type}` : "status";
}

function updatePlaceholders() {
  const selected = state.variantMap.get(elements.variant.value);
  if (!selected) {
    return;
  }
  elements.dark.placeholder = selected.dark;
  elements.light.placeholder = selected.light || "transparent";
  elements.radius.placeholder = selected.radius.toFixed(2);
}

function buildRequest() {
  const radius = parseOptionalFloat(elements.radius.value);
  return {
    data: elements.data.value.trim(),
    variant: elements.variant.value,
    errorLevel: elements.error.value,
    scale: parseIntOr(elements.scale.value, 10),
    border: parseIntOr(elements.border.value, 4),
    dark: elements.dark.value.trim(),
    light: elements.light.value.trim(),
    noBackground: elements.transparent.checked,
    radius,
    pngScale: parseFloatOr(elements.pngScale.value, 3)
  };
}

async function refreshPreview() {
  setStatus("Rendering preview…");
  try {
    const svgBase64 = await GenerateSVG(buildRequest());
    elements.preview.src = `data:image/svg+xml;base64,${svgBase64}`;
    setStatus("Preview updated.");
  } catch (error) {
    setStatus(error.message || String(error), "error");
  }
}

function debounceRefresh() {
  if (state.debounceTimer) {
    clearTimeout(state.debounceTimer);
  }
  state.debounceTimer = setTimeout(() => {
    refreshPreview();
  }, 250);
}

async function exportFile(format) {
  const path = await SuggestSavePath(format);
  if (!path) {
    return;
  }
  setStatus(`Exporting ${format.toUpperCase()}…`);
  try {
    const request = buildRequest();
    const savedPath = format === "png"
      ? await SavePNG(request, path)
      : await SaveSVG(request, path);
    setStatus(`Saved ${savedPath}`);
  } catch (error) {
    setStatus(error.message || String(error), "error");
  }
}

async function loadVariants() {
  const variants = await GetVariantCatalog();
  state.variants = variants;
  state.variantMap.clear();
  elements.variant.innerHTML = "";
  variants.forEach((variant) => {
    state.variantMap.set(variant.name, variant);
    const option = document.createElement("option");
    option.value = variant.name;
    option.textContent = variant.name;
    if (variant.name === "classic") {
      option.selected = true;
    }
    elements.variant.appendChild(option);
  });
  updatePlaceholders();
}

async function init() {
  try {
    await loadVariants();
    const path = await GetVariantsPath();
    elements.variantsPath.textContent = path || "variants.json not found";
  } catch (error) {
    setStatus(error.message || String(error), "error");
  }
  await refreshPreview();
}

[elements.data, elements.error, elements.scale, elements.border, elements.dark, elements.light, elements.radius, elements.transparent, elements.pngScale].forEach((input) => {
  input.addEventListener("input", debounceRefresh);
});

elements.variant.addEventListener("change", () => {
  updatePlaceholders();
  debounceRefresh();
});

elements.refresh.addEventListener("click", refreshPreview);

elements.exportSvg.addEventListener("click", () => exportFile("svg"));

elements.exportPng.addEventListener("click", () => exportFile("png"));

init();
