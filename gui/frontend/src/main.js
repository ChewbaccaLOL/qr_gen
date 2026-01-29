import "./style.css";
import {
  DeleteCustomVariant,
  GeneratePNG,
  GenerateSVG,
  GetVariantCatalog,
  SaveCustomVariant,
  SavePNG,
  SaveSVG,
  SuggestSavePath
} from "../wailsjs/go/main/App";

const elements = {
  data: document.getElementById("data"),
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
  previewFrame: document.getElementById("previewFrame"),
  status: document.getElementById("status"),
  variantGallery: document.getElementById("variantGallery"),
  copyPng: document.getElementById("copyPng"),
  customName: document.getElementById("customName"),
  saveVariant: document.getElementById("saveVariant"),
  deleteVariant: document.getElementById("deleteVariant")
};

const state = {
  variants: [],
  variantMap: new Map(),
  previewCache: new Map(),
  selectedVariant: "",
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

function currentVariantName() {
  if (state.selectedVariant) {
    return state.selectedVariant;
  }
  if (state.variants.length > 0) {
    return state.variants[0].name;
  }
  return "classic";
}

function updateCustomNamePlaceholder(variant) {
  if (!elements.customName || !variant) {
    return;
  }
  if (elements.customName.value.trim()) {
    return;
  }
  elements.customName.placeholder = `${variant.name}-custom`;
}

function updatePlaceholders() {
  const selected = state.variantMap.get(currentVariantName());
  if (!selected) {
    return;
  }
  elements.dark.placeholder = selected.dark;
  elements.light.placeholder = selected.light || "transparent";
  elements.radius.placeholder = selected.radius.toFixed(2);
  updateCustomNamePlaceholder(selected);
}

function previewHasTransparentBackground() {
  if (elements.transparent.checked) {
    return true;
  }
  const selected = state.variantMap.get(currentVariantName());
  if (!selected) {
    return false;
  }
  if (elements.light.value.trim()) {
    return false;
  }
  return !selected.light;
}

function updatePreviewFrame() {
  if (!elements.previewFrame) {
    return;
  }
  elements.previewFrame.classList.toggle("transparent", previewHasTransparentBackground());
}

function buildRequest() {
  const radius = parseOptionalFloat(elements.radius.value);
  return {
    data: elements.data.value.trim(),
    variant: currentVariantName(),
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

function buildGalleryRequest(variantName) {
  return {
    data: defaultData,
    variant: variantName,
    errorLevel: "m",
    scale: 6,
    border: 2,
    dark: "",
    light: "",
    noBackground: false,
    radius: null,
    pngScale: 1
  };
}

function buildCustomVariantRequest() {
  return {
    name: elements.customName.value.trim(),
    baseVariant: currentVariantName(),
    dark: elements.dark.value.trim(),
    light: elements.light.value.trim(),
    noBackground: elements.transparent.checked,
    radius: parseOptionalFloat(elements.radius.value)
  };
}

async function refreshPreview() {
  setStatus("Rendering preview…");
  updatePreviewFrame();
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

function base64ToBlob(base64, contentType) {
  const bytes = atob(base64);
  const out = new Uint8Array(bytes.length);
  for (let i = 0; i < bytes.length; i += 1) {
    out[i] = bytes.charCodeAt(i);
  }
  return new Blob([out], { type: contentType });
}

async function copyPngToClipboard() {
  if (!navigator.clipboard || !navigator.clipboard.write) {
    setStatus("Clipboard image copy is not supported in this environment.", "error");
    return;
  }
  setStatus("Copying PNG…");
  try {
    const pngBase64 = await GeneratePNG(buildRequest());
    const blob = base64ToBlob(pngBase64, "image/png");
    await navigator.clipboard.write([new ClipboardItem({ "image/png": blob })]);
    setStatus("Copied PNG to clipboard.");
  } catch (error) {
    setStatus(error.message || String(error), "error");
  }
}

function renderVariantPreview(variant, imgEl) {
  const cached = state.previewCache.get(variant.name);
  if (cached) {
    imgEl.src = `data:image/svg+xml;base64,${cached}`;
    return;
  }
  imgEl.classList.add("loading");
  GenerateSVG(buildGalleryRequest(variant.name))
    .then((svgBase64) => {
      state.previewCache.set(variant.name, svgBase64);
      imgEl.src = `data:image/svg+xml;base64,${svgBase64}`;
    })
    .catch(() => {
      imgEl.alt = `Failed to render ${variant.name}`;
    })
    .finally(() => {
      imgEl.classList.remove("loading");
    });
}

function buildVariantCard(variant) {
  const card = document.createElement("button");
  card.type = "button";
  card.className = "variant-card";
  card.dataset.variant = variant.name;
  if (!variant.light) {
    card.classList.add("transparent");
  }
  if (variant.isCustom) {
    card.classList.add("custom");
  }
  if (variant.name === currentVariantName()) {
    card.classList.add("active");
  }

  const thumb = document.createElement("div");
  thumb.className = "variant-thumb";

  const img = document.createElement("img");
  img.alt = `${variant.name} QR preview`;
  img.loading = "lazy";

  thumb.appendChild(img);

  const caption = document.createElement("div");
  caption.className = "variant-caption";

  const name = document.createElement("span");
  name.textContent = variant.name;
  caption.appendChild(name);

  if (variant.isCustom) {
    const tag = document.createElement("span");
    tag.className = "variant-tag";
    tag.textContent = "custom";
    caption.appendChild(tag);
  }

  card.appendChild(thumb);
  card.appendChild(caption);

  card.addEventListener("click", () => {
    selectVariant(variant.name);
  });

  renderVariantPreview(variant, img);
  return card;
}

function updateVariantSelectionUI() {
  const cards = elements.variantGallery?.querySelectorAll(".variant-card") || [];
  cards.forEach((card) => {
    card.classList.toggle("active", card.dataset.variant === currentVariantName());
  });
}

function updateCustomActions() {
  const selected = state.variantMap.get(currentVariantName());
  const isCustom = Boolean(selected && selected.isCustom);
  if (elements.deleteVariant) {
    elements.deleteVariant.hidden = !isCustom;
    elements.deleteVariant.disabled = !isCustom;
  }
}

function selectVariant(name) {
  if (!state.variantMap.has(name)) {
    return;
  }
  state.selectedVariant = name;
  updateVariantSelectionUI();
  updatePlaceholders();
  updateCustomActions();
  debounceRefresh();
}

function renderVariantGallery(variants) {
  if (!elements.variantGallery) {
    return;
  }
  elements.variantGallery.innerHTML = "";
  variants.forEach((variant) => {
    const card = buildVariantCard(variant);
    elements.variantGallery.appendChild(card);
  });
}

function prunePreviewCache() {
  const names = new Set(state.variants.map((variant) => variant.name));
  for (const name of state.previewCache.keys()) {
    if (!names.has(name)) {
      state.previewCache.delete(name);
    }
  }
}

async function loadVariants(override) {
  const variants = override || await GetVariantCatalog();
  state.variants = variants;
  state.variantMap.clear();
  variants.forEach((variant) => {
    state.variantMap.set(variant.name, variant);
  });

  const fallback = variants.find((variant) => variant.name === "classic")?.name || variants[0]?.name || "";
  if (!state.selectedVariant || !state.variantMap.has(state.selectedVariant)) {
    state.selectedVariant = fallback;
  }

  prunePreviewCache();
  renderVariantGallery(variants);
  updateVariantSelectionUI();
  updatePlaceholders();
  updateCustomActions();
}

async function saveCustomVariant() {
  const request = buildCustomVariantRequest();
  if (!request.name) {
    setStatus("Custom variant name is required.", "error");
    return;
  }
  setStatus("Saving custom variant…");
  try {
    const variants = await SaveCustomVariant(request);
    elements.customName.value = "";
    state.selectedVariant = request.name;
    await loadVariants(variants);
    debounceRefresh();
    setStatus(`Saved custom variant ${request.name}.`);
  } catch (error) {
    setStatus(error.message || String(error), "error");
  }
}

async function deleteCustomVariant() {
  const name = currentVariantName();
  if (!name) {
    return;
  }
  const confirmed = window.confirm(`Delete custom variant '${name}'?`);
  if (!confirmed) {
    return;
  }
  setStatus(`Deleting ${name}…`);
  try {
    const variants = await DeleteCustomVariant(name);
    state.selectedVariant = "";
    await loadVariants(variants);
    debounceRefresh();
    setStatus(`Deleted custom variant ${name}.`);
  } catch (error) {
    setStatus(error.message || String(error), "error");
  }
}

async function init() {
  try {
    await loadVariants();
  } catch (error) {
    setStatus(error.message || String(error), "error");
  }
  await refreshPreview();
}

[elements.data, elements.error, elements.scale, elements.border, elements.dark, elements.light, elements.radius, elements.transparent, elements.pngScale].forEach((input) => {
  input.addEventListener("input", debounceRefresh);
});

elements.refresh.addEventListener("click", refreshPreview);

elements.exportSvg.addEventListener("click", () => exportFile("svg"));

elements.exportPng.addEventListener("click", () => exportFile("png"));

elements.copyPng?.addEventListener("click", copyPngToClipboard);

elements.saveVariant?.addEventListener("click", saveCustomVariant);

elements.deleteVariant?.addEventListener("click", deleteCustomVariant);

elements.customName?.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    event.preventDefault();
    saveCustomVariant();
  }
});

init();
