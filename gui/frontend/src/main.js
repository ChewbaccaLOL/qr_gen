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
  shape: document.getElementById("shape"),
  error: document.getElementById("error"),
  scale: document.getElementById("scale"),
  border: document.getElementById("border"),
  dark: document.getElementById("dark"),
  darkPicker: document.getElementById("darkPicker"),
  light: document.getElementById("light"),
  lightPicker: document.getElementById("lightPicker"),
  radius: document.getElementById("radius"),
  transparent: document.getElementById("transparent"),
  gradientEnabled: document.getElementById("gradientEnabled"),
  gradientFrom: document.getElementById("gradientFrom"),
  gradientFromPicker: document.getElementById("gradientFromPicker"),
  gradientTo: document.getElementById("gradientTo"),
  gradientToPicker: document.getElementById("gradientToPicker"),
  gradientScope: document.getElementById("gradientScope"),
  gradientAngle: document.getElementById("gradientAngle"),
  gradientAngleValue: document.getElementById("gradientAngleValue"),
  gradientFromStop: document.getElementById("gradientFromStop"),
  gradientFromStopValue: document.getElementById("gradientFromStopValue"),
  gradientToStop: document.getElementById("gradientToStop"),
  gradientToStopValue: document.getElementById("gradientToStopValue"),
  bgGradientEnabled: document.getElementById("bgGradientEnabled"),
  bgGradientFrom: document.getElementById("bgGradientFrom"),
  bgGradientFromPicker: document.getElementById("bgGradientFromPicker"),
  bgGradientTo: document.getElementById("bgGradientTo"),
  bgGradientToPicker: document.getElementById("bgGradientToPicker"),
  bgGradientAngle: document.getElementById("bgGradientAngle"),
  bgGradientAngleValue: document.getElementById("bgGradientAngleValue"),
  bgGradientFromStop: document.getElementById("bgGradientFromStop"),
  bgGradientFromStopValue: document.getElementById("bgGradientFromStopValue"),
  bgGradientToStop: document.getElementById("bgGradientToStop"),
  bgGradientToStopValue: document.getElementById("bgGradientToStopValue"),
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
  debounceTimer: null,
  isWSL: false
};

const defaultData = "https://example.com";
const defaultGradientAngle = 45;
const defaultGradientFromStop = 0;
const defaultGradientToStop = 1;

elements.data.value = defaultData;

const hexShortRegex = /^#([0-9a-f]{3})$/i;
const hexLongRegex = /^#([0-9a-f]{6})$/i;

function normalizeHex(value) {
  const trimmed = value.trim();
  const shortMatch = trimmed.match(hexShortRegex);
  if (shortMatch) {
    const [r, g, b] = shortMatch[1].split("");
    return `#${r}${r}${g}${g}${b}${b}`.toLowerCase();
  }
  if (hexLongRegex.test(trimmed)) {
    return trimmed.toLowerCase();
  }
  return "";
}

function syncPickerToText(textInput, colorInput) {
  if (!textInput || !colorInput) {
    return;
  }
  const normalized = normalizeHex(textInput.value);
  if (normalized) {
    colorInput.value = normalized;
  }
}

function syncTextToPicker(textInput, colorInput) {
  if (!textInput || !colorInput) {
    return;
  }
  textInput.value = colorInput.value;
}

function updateRangeValue(input, output, formatter) {
  if (!input || !output) {
    return;
  }
  const value = Number.parseFloat(input.value);
  if (!Number.isFinite(value)) {
    return;
  }
  output.textContent = formatter ? formatter(value) : String(value);
}

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

function updateGradientFields(variant) {
  if (!elements.gradientEnabled || !elements.gradientFrom || !elements.gradientTo) {
    return;
  }
  elements.gradientEnabled.checked = Boolean(variant?.hasGradient);
  elements.gradientFrom.placeholder = variant?.gradientFrom || variant?.dark || "#ff7a59";
  elements.gradientTo.placeholder = variant?.gradientTo || variant?.dark || "#7a2cff";
  if (elements.gradientScope) {
    elements.gradientScope.value = variant?.gradientScope || "module";
  }
  if (elements.gradientAngle) {
    elements.gradientAngle.value = String(variant?.gradientAngle ?? defaultGradientAngle);
    updateRangeValue(elements.gradientAngle, elements.gradientAngleValue, (value) => `${Math.round(value)}°`);
  }
  if (elements.gradientFromStop) {
    elements.gradientFromStop.value = String(variant?.gradientFromStop ?? defaultGradientFromStop);
    updateRangeValue(elements.gradientFromStop, elements.gradientFromStopValue, (value) => value.toFixed(2));
  }
  if (elements.gradientToStop) {
    elements.gradientToStop.value = String(variant?.gradientToStop ?? defaultGradientToStop);
    updateRangeValue(elements.gradientToStop, elements.gradientToStopValue, (value) => value.toFixed(2));
  }
  updateGradientState();
  setPickerDefault(elements.gradientFrom, elements.gradientFromPicker, elements.gradientFrom.placeholder);
  setPickerDefault(elements.gradientTo, elements.gradientToPicker, elements.gradientTo.placeholder);
}

function updateBackgroundGradientFields(variant) {
  if (!elements.bgGradientEnabled || !elements.bgGradientFrom || !elements.bgGradientTo) {
    return;
  }
  elements.bgGradientEnabled.checked = Boolean(variant?.hasBgGradient);
  elements.bgGradientFrom.placeholder = variant?.bgGradientFrom || variant?.light || "#ffffff";
  elements.bgGradientTo.placeholder = variant?.bgGradientTo || variant?.light || "#f0f0f0";
  if (elements.bgGradientAngle) {
    elements.bgGradientAngle.value = String(variant?.bgGradientAngle ?? defaultGradientAngle);
    updateRangeValue(elements.bgGradientAngle, elements.bgGradientAngleValue, (value) => `${Math.round(value)}°`);
  }
  if (elements.bgGradientFromStop) {
    elements.bgGradientFromStop.value = String(variant?.bgGradientFromStop ?? defaultGradientFromStop);
    updateRangeValue(elements.bgGradientFromStop, elements.bgGradientFromStopValue, (value) => value.toFixed(2));
  }
  if (elements.bgGradientToStop) {
    elements.bgGradientToStop.value = String(variant?.bgGradientToStop ?? defaultGradientToStop);
    updateRangeValue(elements.bgGradientToStop, elements.bgGradientToStopValue, (value) => value.toFixed(2));
  }
  updateBackgroundGradientState();
  setPickerDefault(elements.bgGradientFrom, elements.bgGradientFromPicker, elements.bgGradientFrom.placeholder);
  setPickerDefault(elements.bgGradientTo, elements.bgGradientToPicker, elements.bgGradientTo.placeholder);
}

function updateShapeSelection(variant) {
  if (!elements.shape || !variant) {
    return;
  }
  elements.shape.value = "";
}

function updatePlaceholders() {
  const selected = state.variantMap.get(currentVariantName());
  if (!selected) {
    return;
  }
  elements.dark.placeholder = selected.dark;
  elements.light.placeholder = selected.light || "transparent";
  elements.radius.placeholder = selected.radius.toFixed(2);
  setPickerDefault(elements.dark, elements.darkPicker, selected.dark);
  setPickerDefault(elements.light, elements.lightPicker, selected.light || "#ffffff");
  updateCustomNamePlaceholder(selected);
  updateGradientFields(selected);
  updateBackgroundGradientFields(selected);
  updateShapeSelection(selected);
}

function previewHasTransparentBackground() {
  if (elements.transparent.checked) {
    return true;
  }
  if (elements.bgGradientEnabled?.checked) {
    return false;
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

function resolveGradientColors(selected) {
  if (!elements.gradientEnabled?.checked) {
    return { gradientFrom: "", gradientTo: "" };
  }
  let gradientFrom = elements.gradientFrom?.value.trim() || "";
  let gradientTo = elements.gradientTo?.value.trim() || "";
  if (selected) {
    if (!gradientFrom) {
      gradientFrom = selected.gradientFrom || selected.dark;
    }
    if (!gradientTo) {
      gradientTo = selected.gradientTo || selected.dark;
    }
  }
  return { gradientFrom, gradientTo };
}

function resolveBackgroundGradientColors(selected) {
  if (!elements.bgGradientEnabled?.checked || elements.transparent.checked) {
    return { bgGradientFrom: "", bgGradientTo: "" };
  }
  let bgGradientFrom = elements.bgGradientFrom?.value.trim() || "";
  let bgGradientTo = elements.bgGradientTo?.value.trim() || "";
  if (selected) {
    if (!bgGradientFrom) {
      bgGradientFrom = selected.bgGradientFrom || selected.light || "#ffffff";
    }
    if (!bgGradientTo) {
      bgGradientTo = selected.bgGradientTo || selected.light || "#ffffff";
    }
  }
  return { bgGradientFrom, bgGradientTo };
}

function buildRequest() {
  const radius = parseOptionalFloat(elements.radius.value);
  const selected = state.variantMap.get(currentVariantName());
  const { gradientFrom, gradientTo } = resolveGradientColors(selected);
  const gradientEnabled = Boolean(elements.gradientEnabled?.checked);
  const bgGradientEnabled = Boolean(elements.bgGradientEnabled?.checked) && !elements.transparent.checked;
  const { bgGradientFrom, bgGradientTo } = resolveBackgroundGradientColors(selected);
  return {
    data: elements.data.value.trim(),
    variant: currentVariantName(),
    shape: elements.shape?.value.trim() || "",
    errorLevel: elements.error.value,
    scale: parseIntOr(elements.scale.value, 10),
    border: parseIntOr(elements.border.value, 4),
    dark: elements.dark.value.trim(),
    light: elements.light.value.trim(),
    noBackground: elements.transparent.checked,
    radius,
    gradientEnabled,
    gradientFrom,
    gradientTo,
    gradientAngle: gradientEnabled ? parseFloatOr(elements.gradientAngle?.value, defaultGradientAngle) : null,
    gradientFromStop: gradientEnabled ? parseFloatOr(elements.gradientFromStop?.value, defaultGradientFromStop) : null,
    gradientToStop: gradientEnabled ? parseFloatOr(elements.gradientToStop?.value, defaultGradientToStop) : null,
    gradientScope: gradientEnabled ? (elements.gradientScope?.value || "module") : "",
    bgGradientEnabled,
    bgGradientFrom,
    bgGradientTo,
    bgGradientAngle: bgGradientEnabled ? parseFloatOr(elements.bgGradientAngle?.value, defaultGradientAngle) : null,
    bgGradientFromStop: bgGradientEnabled ? parseFloatOr(elements.bgGradientFromStop?.value, defaultGradientFromStop) : null,
    bgGradientToStop: bgGradientEnabled ? parseFloatOr(elements.bgGradientToStop?.value, defaultGradientToStop) : null,
    pngScale: parseFloatOr(elements.pngScale.value, 3)
  };
}

function buildGalleryRequest(variant) {
  return {
    data: defaultData,
    variant: variant.name,
    shape: "",
    errorLevel: "m",
    scale: 6,
    border: 2,
    dark: "",
    light: "",
    noBackground: false,
    radius: null,
    gradientEnabled: Boolean(variant.hasGradient),
    gradientFrom: variant.gradientFrom || "",
    gradientTo: variant.gradientTo || "",
    gradientAngle: variant.gradientAngle ?? null,
    gradientFromStop: variant.gradientFromStop ?? null,
    gradientToStop: variant.gradientToStop ?? null,
    gradientScope: variant.gradientScope || "",
    bgGradientEnabled: Boolean(variant.hasBgGradient),
    bgGradientFrom: variant.bgGradientFrom || "",
    bgGradientTo: variant.bgGradientTo || "",
    bgGradientAngle: variant.bgGradientAngle ?? null,
    bgGradientFromStop: variant.bgGradientFromStop ?? null,
    bgGradientToStop: variant.bgGradientToStop ?? null,
    pngScale: 1
  };
}

function buildCustomVariantRequest() {
  const selected = state.variantMap.get(currentVariantName());
  const { gradientFrom, gradientTo } = resolveGradientColors(selected);
  const gradientEnabled = Boolean(elements.gradientEnabled?.checked);
  const bgGradientEnabled = Boolean(elements.bgGradientEnabled?.checked) && !elements.transparent.checked;
  const { bgGradientFrom, bgGradientTo } = resolveBackgroundGradientColors(selected);
  return {
    name: elements.customName.value.trim(),
    baseVariant: currentVariantName(),
    dark: elements.dark.value.trim(),
    light: elements.light.value.trim(),
    noBackground: elements.transparent.checked,
    radius: parseOptionalFloat(elements.radius.value),
    shape: elements.shape?.value.trim() || "",
    gradientEnabled,
    gradientFrom,
    gradientTo,
    gradientAngle: gradientEnabled ? parseFloatOr(elements.gradientAngle?.value, defaultGradientAngle) : null,
    gradientFromStop: gradientEnabled ? parseFloatOr(elements.gradientFromStop?.value, defaultGradientFromStop) : null,
    gradientToStop: gradientEnabled ? parseFloatOr(elements.gradientToStop?.value, defaultGradientToStop) : null,
    gradientScope: gradientEnabled ? (elements.gradientScope?.value || "module") : "",
    bgGradientEnabled,
    bgGradientFrom,
    bgGradientTo,
    bgGradientAngle: bgGradientEnabled ? parseFloatOr(elements.bgGradientAngle?.value, defaultGradientAngle) : null,
    bgGradientFromStop: bgGradientEnabled ? parseFloatOr(elements.bgGradientFromStop?.value, defaultGradientFromStop) : null,
    bgGradientToStop: bgGradientEnabled ? parseFloatOr(elements.bgGradientToStop?.value, defaultGradientToStop) : null,
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
  const wslNote = state.isWSL
    ? " (WSL: Windows clipboard may not update.)"
    : "";
  setStatus(`Copying PNG…${wslNote}`, state.isWSL ? "warning" : "");
  try {
    const pngBase64 = await GeneratePNG(buildRequest());
    const blob = base64ToBlob(pngBase64, "image/png");
    await navigator.clipboard.write([new ClipboardItem({ "image/png": blob })]);
    setStatus(`Copied PNG to clipboard.${wslNote}`, state.isWSL ? "warning" : "");
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
  GenerateSVG(buildGalleryRequest(variant))
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

function updateGradientState() {
  const enabled = Boolean(elements.gradientEnabled?.checked);
  const gradientInputs = [
    elements.gradientFrom,
    elements.gradientFromPicker,
    elements.gradientTo,
    elements.gradientToPicker,
    elements.gradientScope,
    elements.gradientAngle,
    elements.gradientFromStop,
    elements.gradientToStop
  ];
  gradientInputs.forEach((input) => {
    if (!input) {
      return;
    }
    input.disabled = !enabled;
  });
}

function updateBackgroundGradientState() {
  const enabled = Boolean(elements.bgGradientEnabled?.checked) && !elements.transparent.checked;
  if (elements.bgGradientEnabled) {
    elements.bgGradientEnabled.disabled = elements.transparent.checked;
  }
  const inputs = [
    elements.bgGradientFrom,
    elements.bgGradientFromPicker,
    elements.bgGradientTo,
    elements.bgGradientToPicker,
    elements.bgGradientAngle,
    elements.bgGradientFromStop,
    elements.bgGradientToStop
  ];
  inputs.forEach((input) => {
    if (!input) {
      return;
    }
    input.disabled = !enabled;
  });
}

function setPickerDefault(textInput, colorInput, fallback) {
  if (!textInput || !colorInput) {
    return;
  }
  const candidate = textInput.value.trim() || fallback || "";
  const normalized = normalizeHex(candidate);
  if (normalized) {
    colorInput.value = normalized;
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

async function detectWSL() {
  const wslCheck = window.go?.main?.App?.IsWSL;
  if (!wslCheck) {
    return;
  }
  try {
    state.isWSL = await wslCheck();
  } catch {
    state.isWSL = false;
  }
}

async function init() {
  await detectWSL();
  try {
    await loadVariants();
  } catch (error) {
    setStatus(error.message || String(error), "error");
  }
  await refreshPreview();
}

[elements.data, elements.shape, elements.error, elements.scale, elements.border, elements.dark, elements.light, elements.radius, elements.transparent, elements.gradientEnabled, elements.gradientFrom, elements.gradientTo, elements.gradientScope, elements.gradientAngle, elements.gradientFromStop, elements.gradientToStop, elements.bgGradientEnabled, elements.bgGradientFrom, elements.bgGradientTo, elements.bgGradientAngle, elements.bgGradientFromStop, elements.bgGradientToStop, elements.pngScale].forEach((input) => {
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

elements.gradientEnabled?.addEventListener("change", () => {
  updateGradientState();
  debounceRefresh();
});

elements.bgGradientEnabled?.addEventListener("change", () => {
  updateBackgroundGradientState();
  debounceRefresh();
});

elements.transparent?.addEventListener("change", () => {
  updateBackgroundGradientState();
});

[
  [elements.gradientAngle, elements.gradientAngleValue, (value) => `${Math.round(value)}°`],
  [elements.gradientFromStop, elements.gradientFromStopValue, (value) => value.toFixed(2)],
  [elements.gradientToStop, elements.gradientToStopValue, (value) => value.toFixed(2)],
  [elements.bgGradientAngle, elements.bgGradientAngleValue, (value) => `${Math.round(value)}°`],
  [elements.bgGradientFromStop, elements.bgGradientFromStopValue, (value) => value.toFixed(2)],
  [elements.bgGradientToStop, elements.bgGradientToStopValue, (value) => value.toFixed(2)]
].forEach(([input, output, formatter]) => {
  if (!input || !output) {
    return;
  }
  input.addEventListener("input", () => {
    updateRangeValue(input, output, formatter);
  });
});

[
  [elements.dark, elements.darkPicker],
  [elements.light, elements.lightPicker],
  [elements.gradientFrom, elements.gradientFromPicker],
  [elements.gradientTo, elements.gradientToPicker],
  [elements.bgGradientFrom, elements.bgGradientFromPicker],
  [elements.bgGradientTo, elements.bgGradientToPicker]
].forEach(([textInput, colorInput]) => {
  if (!textInput || !colorInput) {
    return;
  }
  textInput.addEventListener("input", () => {
    syncPickerToText(textInput, colorInput);
    debounceRefresh();
  });
  colorInput.addEventListener("input", () => {
    syncTextToPicker(textInput, colorInput);
    debounceRefresh();
  });
});

init();
