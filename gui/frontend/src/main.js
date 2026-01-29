import "./style.css";
import {
  DeleteCustomVariant,
  GenerateGIF,
  GeneratePNG,
  GenerateSVG,
  GetAnimationConfig,
  GetVariantCatalog,
  SaveCustomVariant,
  SaveGIF,
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
  backgroundModeSolid: document.getElementById("backgroundModeSolid"),
  backgroundModeTransparent: document.getElementById("backgroundModeTransparent"),
  backgroundModeCutout: document.getElementById("backgroundModeCutout"),
  backgroundModeOptions: Array.from(document.querySelectorAll('input[name="backgroundMode"]')),
  radius: document.getElementById("radius"),
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
  deleteVariant: document.getElementById("deleteVariant"),
  tabButtons: Array.from(document.querySelectorAll("[data-tab]")),
  tabPanels: Array.from(document.querySelectorAll("[data-panel]")),
  animationVariant: document.getElementById("animationVariant"),
  gifScale: document.getElementById("gifScale"),
  gifFps: document.getElementById("gifFps"),
  gifFrames: document.getElementById("gifFrames"),
  gifHold: document.getElementById("gifHold"),
  waveAmp: document.getElementById("waveAmp"),
  wavePeriod: document.getElementById("wavePeriod"),
  floatAngle: document.getElementById("floatAngle"),
  floatCycles: document.getElementById("floatCycles"),
  readableGif: document.getElementById("readableGif"),
  renderAnimation: document.getElementById("renderAnimation"),
  exportGif: document.getElementById("exportGif"),
  animationAuto: document.getElementById("animationAuto"),
  animationInterval: document.getElementById("animationInterval"),
  animationPreview: document.getElementById("animationPreview"),
  animationPreviewFrame: document.getElementById("animationPreviewFrame"),
  animationStatus: document.getElementById("animationStatus"),
  animationGallery: document.getElementById("animationGallery"),
  animationPlaceholderTitle: document.getElementById("animationPlaceholderTitle"),
  animationPlaceholderCopy: document.getElementById("animationPlaceholderCopy"),
  animationDebug: document.getElementById("animationDebug")
};

const state = {
  variants: [],
  variantMap: new Map(),
  previewCache: new Map(),
  selectedVariant: "",
  debounceTimer: null,
  isRendering: false,
  pendingRender: false,
  isWSL: false,
  animationVariants: [],
  animationDefaults: null,
  selectedAnimationVariant: "",
  animationIsRendering: false,
  animationPendingRender: false,
  animationDirty: true,
  animationHasRender: false,
  animationAutoTimer: null,
  animationPreviewUrl: "",
  autoLightValue: "",
  autoBgGradientFrom: "",
  autoBgGradientTo: ""
};

const angleControls = new Map();

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

function normalizeNamedColor(value) {
  const trimmed = value.trim().toLowerCase();
  if (trimmed === "black") {
    return "#000000";
  }
  if (trimmed === "white") {
    return "#ffffff";
  }
  return "";
}

function normalizeColor(value) {
  return normalizeHex(value) || normalizeNamedColor(value);
}

function parseHexColor(value) {
  const normalized = normalizeColor(value);
  if (!normalized) {
    return null;
  }
  return {
    r: Number.parseInt(normalized.slice(1, 3), 16),
    g: Number.parseInt(normalized.slice(3, 5), 16),
    b: Number.parseInt(normalized.slice(5, 7), 16)
  };
}

function meanHex(a, b) {
  const ca = parseHexColor(a);
  const cb = parseHexColor(b);
  if (!ca && !cb) {
    return "";
  }
  if (!cb) {
    return normalizeColor(a);
  }
  if (!ca) {
    return normalizeColor(b);
  }
  const r = Math.round((ca.r + cb.r) / 2);
  const g = Math.round((ca.g + cb.g) / 2);
  const bVal = Math.round((ca.b + cb.b) / 2);
  return `#${r.toString(16).padStart(2, "0")}${g.toString(16).padStart(2, "0")}${bVal
    .toString(16)
    .padStart(2, "0")}`;
}

function invertHex(value) {
  const normalized = normalizeColor(value);
  if (!normalized) {
    return "";
  }
  const r = 255 - Number.parseInt(normalized.slice(1, 3), 16);
  const g = 255 - Number.parseInt(normalized.slice(3, 5), 16);
  const b = 255 - Number.parseInt(normalized.slice(5, 7), 16);
  return `#${r.toString(16).padStart(2, "0")}${g.toString(16).padStart(2, "0")}${b.toString(16).padStart(2, "0")}`;
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

function parseOptionalInt(value) {
  const trimmed = value.trim();
  if (!trimmed) {
    return null;
  }
  const parsed = Number.parseInt(trimmed, 10);
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

function clampValue(value, min, max) {
  return Math.min(Math.max(value, min), max);
}

function stepPrecision(step) {
  if (!Number.isFinite(step) || step <= 0) {
    return 0;
  }
  const parts = String(step).split(".");
  return parts[1] ? parts[1].length : 0;
}

function updateAngleControl(targetId) {
  const control = angleControls.get(targetId);
  if (!control) {
    return;
  }
  control.update();
  control.updateColors?.();
}

function setAngleControlEnabled(targetId, enabled) {
  angleControls.get(targetId)?.setEnabled(enabled);
}

function updateAngleControlColors(targetId) {
  angleControls.get(targetId)?.updateColors?.();
}

function updateAngleControlColorsForInput(inputId) {
  if (inputId === "gradientFrom" || inputId === "gradientTo") {
    updateAngleControlColors("gradientAngle");
  }
  if (inputId === "bgGradientFrom" || inputId === "bgGradientTo") {
    updateAngleControlColors("bgGradientAngle");
  }
}

function setupAngleControl(control) {
  const targetId = control.dataset.target;
  if (!targetId) {
    return null;
  }
  const input = document.getElementById(targetId);
  const dial = control.querySelector(".angle-dial");
  const fromInput = control.dataset.from ? document.getElementById(control.dataset.from) : null;
  const toInput = control.dataset.to ? document.getElementById(control.dataset.to) : null;
  if (!input || !dial) {
    return null;
  }

  const min = parseFloatOr(input.min, 0);
  const max = parseFloatOr(input.max, 360);

  if (control.dataset.label) {
    dial.setAttribute("aria-label", control.dataset.label);
  }
  dial.setAttribute("aria-valuemin", String(min));
  dial.setAttribute("aria-valuemax", String(max));

  const updateDial = () => {
    const value = parseFloatOr(input.value, min);
    dial.style.setProperty("--angle", `${value}deg`);
    dial.setAttribute("aria-valuenow", String(Math.round(value)));
    dial.setAttribute("aria-valuetext", `${Math.round(value)} degrees`);
  };

  const resolveControlColor = (colorInput) => {
    if (!colorInput) {
      return "";
    }
    return colorInput.value.trim() || colorInput.placeholder || "";
  };

  const updateColors = () => {
    const fromColor = resolveControlColor(fromInput);
    const toColor = resolveControlColor(toInput);
    if (fromColor) {
      dial.style.setProperty("--angle-from", fromColor);
    } else {
      dial.style.removeProperty("--angle-from");
    }
    if (toColor) {
      dial.style.setProperty("--angle-to", toColor);
    } else {
      dial.style.removeProperty("--angle-to");
    }
  };

  const applyAngle = (rawAngle) => {
    const step = parseFloatOr(input.step, 1);
    const precision = stepPrecision(step);
    const snapped = step ? Math.round(rawAngle / step) * step : rawAngle;
    const clamped = clampValue(snapped, min, max);
    const value = precision ? clamped.toFixed(precision) : String(Math.round(clamped));
    input.value = value;
    updateDial();
    input.dispatchEvent(new Event("input", { bubbles: true }));
  };

  const angleFromPointer = (event) => {
    const rect = dial.getBoundingClientRect();
    const dx = event.clientX - (rect.left + rect.width / 2);
    const dy = event.clientY - (rect.top + rect.height / 2);
    if (dx === 0 && dy === 0) {
      return null;
    }
    const radians = Math.atan2(dy, dx);
    const degrees = (radians * 180) / Math.PI;
    return (degrees + 360) % 360;
  };

  let dragging = false;
  dial.addEventListener("pointerdown", (event) => {
    if (input.disabled || event.button !== 0) {
      return;
    }
    dragging = true;
    dial.setPointerCapture(event.pointerId);
    const angle = angleFromPointer(event);
    if (angle !== null) {
      applyAngle(angle);
    }
    event.preventDefault();
  });

  dial.addEventListener("pointermove", (event) => {
    if (!dragging) {
      return;
    }
    const angle = angleFromPointer(event);
    if (angle !== null) {
      applyAngle(angle);
    }
  });

  dial.addEventListener("pointerup", (event) => {
    if (!dragging) {
      return;
    }
    dragging = false;
    dial.releasePointerCapture(event.pointerId);
  });

  dial.addEventListener("pointercancel", () => {
    dragging = false;
  });

  dial.addEventListener("wheel", (event) => {
    if (input.disabled) {
      return;
    }
    event.preventDefault();
    const delta = Math.sign(event.deltaY);
    if (!delta) {
      return;
    }
    const step = parseFloatOr(input.step, 1);
    const current = parseFloatOr(input.value, min);
    applyAngle(current - delta * step);
  }, { passive: false });

  dial.addEventListener("keydown", (event) => {
    if (input.disabled) {
      return;
    }
    const step = parseFloatOr(input.step, 1);
    const current = parseFloatOr(input.value, min);
    switch (event.key) {
      case "ArrowUp":
      case "ArrowRight":
        applyAngle(current + step);
        event.preventDefault();
        break;
      case "ArrowDown":
      case "ArrowLeft":
        applyAngle(current - step);
        event.preventDefault();
        break;
      case "Home":
        applyAngle(min);
        event.preventDefault();
        break;
      case "End":
        applyAngle(max);
        event.preventDefault();
        break;
      default:
        break;
    }
  });

  input.addEventListener("input", updateDial);
  fromInput?.addEventListener("input", updateColors);
  toInput?.addEventListener("input", updateColors);
  updateDial();
  updateColors();

  const setEnabled = (enabled) => {
    control.classList.toggle("disabled", !enabled);
    dial.tabIndex = enabled ? 0 : -1;
    dial.setAttribute("aria-disabled", String(!enabled));
  };

  setEnabled(!input.disabled);

  return { targetId, update: updateDial, updateColors, setEnabled };
}

function setupAngleControls() {
  const controls = document.querySelectorAll(".angle-control");
  controls.forEach((control) => {
    const handle = setupAngleControl(control);
    if (handle) {
      angleControls.set(handle.targetId, handle);
    }
  });
}

function setStatus(message, type = "") {
  elements.status.textContent = message;
  elements.status.className = type ? `status ${type}` : "status";
}

function setAnimationStatus(message, type = "") {
  if (!elements.animationStatus) {
    return;
  }
  elements.animationStatus.textContent = message;
  elements.animationStatus.className = type ? `status ${type}` : "status";
}

function setActiveTab(tabName) {
  elements.tabButtons.forEach((button) => {
    const isActive = button.dataset.tab === tabName;
    button.classList.toggle("active", isActive);
    button.setAttribute("aria-selected", String(isActive));
  });
  elements.tabPanels.forEach((panel) => {
    const isActive = panel.dataset.panel === tabName;
    panel.classList.toggle("active", isActive);
    panel.hidden = !isActive;
  });
}

function setupTabs() {
  if (!elements.tabButtons.length || !elements.tabPanels.length) {
    return;
  }
  elements.tabButtons.forEach((button) => {
    button.addEventListener("click", () => {
      setActiveTab(button.dataset.tab);
    });
  });
  setActiveTab("qr");
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

function currentAnimationVariant() {
  if (state.selectedAnimationVariant) {
    return state.selectedAnimationVariant;
  }
  if (state.animationVariants.length > 0) {
    return state.animationVariants[0];
  }
  return "wave";
}

function animationDefaultsForState() {
  if (!state.animationDefaults) {
    return null;
  }
  if (elements.readableGif?.checked) {
    return state.animationDefaults.readable;
  }
  return state.animationDefaults;
}

function setAnimationPlaceholder(title, copy) {
  if (elements.animationPlaceholderTitle) {
    elements.animationPlaceholderTitle.textContent = title;
  }
  if (elements.animationPlaceholderCopy) {
    elements.animationPlaceholderCopy.textContent = copy;
  }
}

function updateAnimationPreviewFrame() {
  if (!elements.animationPreviewFrame) {
    return;
  }
  elements.animationPreviewFrame.classList.toggle("transparent", previewHasTransparentBackground());
}

function updateAnimationDebug() {
  if (!elements.animationDebug) {
    return;
  }
  const req = buildAnimationRequest();
  const payload = {
    variant: req.variant,
    shape: req.shape,
    animationVariant: req.animationVariant,
    scale: req.scale,
    gifScale: req.gifScale ?? 1,
    border: req.border,
    radius: req.radius,
    gradientEnabled: req.gradientEnabled,
    bgGradientEnabled: req.bgGradientEnabled,
    noBackground: req.noBackground,
    gifFps: req.gifFps ?? "(default)",
    gifFrames: req.gifFrames ?? "(default)",
    gifHold: req.gifHold ?? "(default)",
    waveAmp: req.waveAmp ?? "(default)",
    wavePeriod: req.wavePeriod ?? "(default)",
    floatAngle: req.floatAngle ?? "(default)",
    floatCycles: req.floatCycles ?? "(default)",
    readableGif: req.readableGif
  };
  elements.animationDebug.textContent = JSON.stringify(payload, null, 2);
}

function markAnimationDirty() {
  state.animationDirty = true;
  if (state.animationAutoTimer) {
    return;
  }
  setAnimationStatus("Animation settings changed. Render to update.");
  updateAnimationDebug();
}

function updateAnimationHoldPlaceholder() {
  if (!elements.gifHold || !state.animationDefaults) {
    return;
  }
  const defaults = animationDefaultsForState();
  if (!defaults) {
    return;
  }
  let holdValue = defaults.gifHold;
  const variant = currentAnimationVariant();
  if (variant === "wave-loop") {
    holdValue = 0;
  }
  if (variant.startsWith("float") && state.animationDefaults.floatHold > 0) {
    holdValue = state.animationDefaults.floatHold;
  }
  elements.gifHold.placeholder = String(holdValue);
}

function updateAnimationFloatAnglePlaceholder() {
  if (!elements.floatAngle || !state.animationDefaults) {
    return;
  }
  const variant = currentAnimationVariant();
  let angle = state.animationDefaults.floatAngle;
  if (variant !== "float-tilt-first") {
    angle += state.animationDefaults.floatTilt;
  }
  elements.floatAngle.placeholder = String(Math.round(angle));
}

function updateAnimationPlaceholders() {
  if (!state.animationDefaults) {
    return;
  }
  const defaults = animationDefaultsForState();
  if (!defaults) {
    return;
  }
  if (elements.gifFps) {
    elements.gifFps.placeholder = String(defaults.gifFps);
  }
  if (elements.gifScale) {
    elements.gifScale.placeholder = "1";
  }
  if (elements.gifFrames) {
    elements.gifFrames.placeholder = String(defaults.gifFrames);
  }
  if (elements.waveAmp) {
    elements.waveAmp.placeholder = String(defaults.waveAmp);
  }
  if (elements.wavePeriod) {
    elements.wavePeriod.placeholder = String(defaults.wavePeriod);
  }
  if (elements.floatCycles) {
    elements.floatCycles.placeholder = String(state.animationDefaults.floatCycles || 1);
  }
  updateAnimationHoldPlaceholder();
  updateAnimationFloatAnglePlaceholder();
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
  elements.gradientFrom.value = variant?.hasGradient ? (variant?.gradientFrom || "") : "";
  elements.gradientTo.value = variant?.hasGradient ? (variant?.gradientTo || "") : "";
  if (elements.gradientScope) {
    elements.gradientScope.value = variant?.gradientScope || "global";
  }
  if (elements.gradientAngle) {
    elements.gradientAngle.value = String(variant?.gradientAngle ?? defaultGradientAngle);
    updateRangeValue(elements.gradientAngle, elements.gradientAngleValue, (value) => `${Math.round(value)}°`);
    updateAngleControl("gradientAngle");
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
  elements.bgGradientFrom.value = variant?.hasBgGradient ? (variant?.bgGradientFrom || "") : "";
  elements.bgGradientTo.value = variant?.hasBgGradient ? (variant?.bgGradientTo || "") : "";
  if (elements.bgGradientAngle) {
    elements.bgGradientAngle.value = String(variant?.bgGradientAngle ?? defaultGradientAngle);
    updateRangeValue(elements.bgGradientAngle, elements.bgGradientAngleValue, (value) => `${Math.round(value)}°`);
    updateAngleControl("bgGradientAngle");
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
  syncBackgroundModeFromVariant(selected);
  ensureBackgroundDefaults(selected);
}

function syncBackgroundModeFromVariant(variant) {
  if (!variant) {
    return;
  }
  const shouldBeTransparent = !variant.light;
  if (shouldBeTransparent && elements.backgroundModeTransparent) {
    elements.backgroundModeTransparent.checked = true;
  } else if (elements.backgroundModeSolid) {
    elements.backgroundModeSolid.checked = true;
  }
  updateBackgroundGradientState();
  updatePreviewFrame();
  updateAnimationPreviewFrame();
}

function currentBackgroundMode() {
  if (elements.backgroundModeCutout?.checked) {
    return "cutout";
  }
  if (elements.backgroundModeTransparent?.checked) {
    return "transparent";
  }
  return "solid";
}

function backgroundModeIsTransparent() {
  return currentBackgroundMode() === "transparent";
}

function backgroundModeIsCutout() {
  return currentBackgroundMode() === "cutout";
}

function previewHasTransparentBackground() {
  const mode = currentBackgroundMode();
  if (mode === "transparent" || mode === "cutout") {
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
  if (!elements.bgGradientEnabled?.checked || backgroundModeIsTransparent()) {
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
  const backgroundMode = currentBackgroundMode();
  const bgGradientEnabled = Boolean(elements.bgGradientEnabled?.checked) && backgroundMode !== "transparent";
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
    noBackground: backgroundMode === "transparent",
    cutout: backgroundMode === "cutout",
    radius,
    gradientEnabled,
    gradientFrom,
    gradientTo,
    gradientAngle: gradientEnabled ? parseFloatOr(elements.gradientAngle?.value, defaultGradientAngle) : null,
    gradientFromStop: gradientEnabled ? parseFloatOr(elements.gradientFromStop?.value, defaultGradientFromStop) : null,
    gradientToStop: gradientEnabled ? parseFloatOr(elements.gradientToStop?.value, defaultGradientToStop) : null,
    gradientScope: gradientEnabled ? (elements.gradientScope?.value || "global") : "",
    bgGradientEnabled,
    bgGradientFrom,
    bgGradientTo,
    bgGradientAngle: bgGradientEnabled ? parseFloatOr(elements.bgGradientAngle?.value, defaultGradientAngle) : null,
    bgGradientFromStop: bgGradientEnabled ? parseFloatOr(elements.bgGradientFromStop?.value, defaultGradientFromStop) : null,
    bgGradientToStop: bgGradientEnabled ? parseFloatOr(elements.bgGradientToStop?.value, defaultGradientToStop) : null,
    pngScale: parseFloatOr(elements.pngScale.value, 3)
  };
}

function buildAnimationRequest() {
  const request = buildRequest();
  const selected = state.variantMap.get(currentVariantName());
  if (!request.shape && selected?.shape) {
    request.shape = selected.shape;
  }
  const gifScale = parseFloatOr(elements.gifScale?.value || "", 1);
  const scaleValue = parseIntOr(elements.scale.value, 10);
  const scaled = Math.max(1, Math.round(scaleValue * gifScale));
  request.scale = scaled;
  return {
    ...request,
    animationVariant: currentAnimationVariant(),
    gifScale,
    gifFps: parseOptionalInt(elements.gifFps?.value || ""),
    gifFrames: parseOptionalInt(elements.gifFrames?.value || ""),
    gifHold: parseOptionalInt(elements.gifHold?.value || ""),
    waveAmp: parseOptionalFloat(elements.waveAmp?.value || ""),
    wavePeriod: parseOptionalFloat(elements.wavePeriod?.value || ""),
    floatAngle: parseOptionalFloat(elements.floatAngle?.value || ""),
    floatCycles: parseOptionalInt(elements.floatCycles?.value || ""),
    readableGif: Boolean(elements.readableGif?.checked)
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
    cutout: false,
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
  const bgGradientEnabled = Boolean(elements.bgGradientEnabled?.checked) && !backgroundModeIsTransparent();
  const { bgGradientFrom, bgGradientTo } = resolveBackgroundGradientColors(selected);
  return {
    name: elements.customName.value.trim(),
    baseVariant: currentVariantName(),
    dark: elements.dark.value.trim(),
    light: elements.light.value.trim(),
    noBackground: backgroundModeIsTransparent(),
    radius: parseOptionalFloat(elements.radius.value),
    shape: elements.shape?.value.trim() || "",
    gradientEnabled,
    gradientFrom,
    gradientTo,
    gradientAngle: gradientEnabled ? parseFloatOr(elements.gradientAngle?.value, defaultGradientAngle) : null,
    gradientFromStop: gradientEnabled ? parseFloatOr(elements.gradientFromStop?.value, defaultGradientFromStop) : null,
    gradientToStop: gradientEnabled ? parseFloatOr(elements.gradientToStop?.value, defaultGradientToStop) : null,
    gradientScope: gradientEnabled ? (elements.gradientScope?.value || "global") : "",
    bgGradientEnabled,
    bgGradientFrom,
    bgGradientTo,
    bgGradientAngle: bgGradientEnabled ? parseFloatOr(elements.bgGradientAngle?.value, defaultGradientAngle) : null,
    bgGradientFromStop: bgGradientEnabled ? parseFloatOr(elements.bgGradientFromStop?.value, defaultGradientFromStop) : null,
    bgGradientToStop: bgGradientEnabled ? parseFloatOr(elements.bgGradientToStop?.value, defaultGradientToStop) : null,
  };
}

async function refreshPreview() {
  if (state.isRendering) {
    state.pendingRender = true;
    return;
  }
  state.isRendering = true;
  setStatus("Rendering preview…");
  updatePreviewFrame();
  try {
    const svgBase64 = await GenerateSVG(buildRequest());
    elements.preview.src = `data:image/svg+xml;base64,${svgBase64}`;
    setStatus("Preview updated.");
  } catch (error) {
    setStatus(error.message || String(error), "error");
  } finally {
    state.isRendering = false;
    if (state.pendingRender) {
      state.pendingRender = false;
      requestAnimationFrame(() => {
        refreshPreview();
      });
    }
  }
}

async function refreshAnimationPreview() {
  if (state.animationIsRendering) {
    state.animationPendingRender = true;
    return;
  }
  updateAnimationDebug();
  state.animationIsRendering = true;
  if (elements.renderAnimation) {
    elements.renderAnimation.disabled = true;
  }
  setAnimationStatus("Rendering animation…");
  updateAnimationPreviewFrame();
  if (elements.animationPreviewFrame) {
    elements.animationPreviewFrame.classList.add("loading");
  }
  setAnimationPlaceholder("Rendering animation…", "This can take a few seconds.");
  try {
    const gifBase64 = await GenerateGIF(buildAnimationRequest());
    setAnimationPreviewSource(gifBase64);
    if (elements.animationPreviewFrame) {
      elements.animationPreviewFrame.classList.add("has-media");
    }
    state.animationHasRender = true;
    state.animationDirty = false;
    setAnimationStatus("Animation preview updated.");
  } catch (error) {
    setAnimationStatus(error.message || String(error), "error");
    state.animationDirty = false;
    state.animationHasRender = false;
    if (elements.animationPreviewFrame) {
      elements.animationPreviewFrame.classList.remove("has-media");
    }
    setAnimationPlaceholder("Render failed", "Adjust settings and try again.");
  } finally {
    state.animationIsRendering = false;
    if (elements.renderAnimation) {
      elements.renderAnimation.disabled = false;
    }
    if (elements.animationPreviewFrame) {
      elements.animationPreviewFrame.classList.remove("loading");
    }
    if (state.animationPendingRender) {
      state.animationPendingRender = false;
      requestAnimationFrame(() => {
        refreshAnimationPreview();
      });
    }
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

function restartAnimationAutoRender() {
  if (!elements.animationAuto?.checked) {
    return;
  }
  const intervalSeconds = clampValue(parseIntOr(elements.animationInterval?.value || "20", 20), 5, 120);
  if (elements.animationInterval) {
    elements.animationInterval.value = String(intervalSeconds);
  }
  if (state.animationAutoTimer) {
    clearInterval(state.animationAutoTimer);
  }
  state.animationAutoTimer = setInterval(() => {
    if (state.animationDirty) {
      refreshAnimationPreview();
    }
  }, intervalSeconds * 1000);
  if (state.animationDirty) {
    refreshAnimationPreview();
  } else {
    setAnimationStatus("Auto-render armed. Waiting for changes.");
  }
  if (state.animationDirty) {
    setAnimationStatus(`Auto-rendering every ${intervalSeconds}s.`);
  }
}

function stopAnimationAutoRender() {
  if (state.animationAutoTimer) {
    clearInterval(state.animationAutoTimer);
    state.animationAutoTimer = null;
  }
}

function toggleAnimationAutoRender() {
  if (elements.animationAuto?.checked) {
    restartAnimationAutoRender();
  } else {
    stopAnimationAutoRender();
    setAnimationStatus("Auto-render paused.");
  }
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

async function exportGif() {
  const path = await SuggestSavePath("gif");
  if (!path) {
    return;
  }
  setAnimationStatus("Exporting GIF…");
  try {
    const savedPath = await SaveGIF(buildAnimationRequest(), path);
    setAnimationStatus(`Saved ${savedPath}`);
  } catch (error) {
    setAnimationStatus(error.message || String(error), "error");
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

function setAnimationPreviewSource(base64) {
  if (!elements.animationPreview) {
    return;
  }
  const blob = base64ToBlob(base64, "image/gif");
  const url = URL.createObjectURL(blob);
  if (state.animationPreviewUrl) {
    URL.revokeObjectURL(state.animationPreviewUrl);
  }
  state.animationPreviewUrl = url;
  elements.animationPreview.src = url;
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

function updateAnimationVariantControls() {
  const isFloat = currentAnimationVariant().startsWith("float");
  const isLoop = currentAnimationVariant() === "wave-loop";
  document.querySelectorAll("[data-float-only]").forEach((block) => {
    block.classList.toggle("is-hidden", !isFloat);
  });
  if (elements.gifHold) {
    elements.gifHold.disabled = isLoop;
  }
  updateAnimationHoldPlaceholder();
  updateAnimationFloatAnglePlaceholder();
}

function updateAnimationVariantSelection() {
  if (elements.animationVariant) {
    elements.animationVariant.value = currentAnimationVariant();
  }
  document.querySelectorAll(".motion-card").forEach((card) => {
    card.classList.toggle("active", card.dataset.animation === currentAnimationVariant());
  });
}

function selectAnimationVariant(name) {
  if (!name) {
    return;
  }
  state.selectedAnimationVariant = name;
  updateAnimationVariantSelection();
  updateAnimationVariantControls();
  markAnimationDirty();
}

function buildAnimationCard(name) {
  const card = document.createElement("button");
  card.type = "button";
  card.className = "variant-card motion-card";
  card.dataset.animation = name;
  if (name === currentAnimationVariant()) {
    card.classList.add("active");
  }

  const thumb = document.createElement("div");
  thumb.className = "variant-thumb motion-thumb";

  const glyph = document.createElement("div");
  glyph.className = "motion-glyph";
  glyph.textContent = name.startsWith("float") ? "float" : "wave";
  thumb.appendChild(glyph);

  const caption = document.createElement("div");
  caption.className = "variant-caption";

  const title = document.createElement("span");
  title.textContent = name.replace(/-/g, " ");
  caption.appendChild(title);

  const tag = document.createElement("span");
  tag.className = "variant-tag";
  tag.textContent = name.startsWith("float") ? "float" : "wave";
  caption.appendChild(tag);

  card.appendChild(thumb);
  card.appendChild(caption);
  card.addEventListener("click", () => {
    selectAnimationVariant(name);
  });
  return card;
}

function renderAnimationGallery(variants) {
  if (!elements.animationGallery) {
    return;
  }
  elements.animationGallery.innerHTML = "";
  variants.forEach((name) => {
    const card = buildAnimationCard(name);
    elements.animationGallery.appendChild(card);
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
  updateAnimationPreviewFrame();
  markAnimationDirty();
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
  setAngleControlEnabled("gradientAngle", enabled);
}

function updateBackgroundGradientState() {
  const enabled = Boolean(elements.bgGradientEnabled?.checked) && !backgroundModeIsTransparent();
  if (elements.bgGradientEnabled) {
    elements.bgGradientEnabled.disabled = backgroundModeIsTransparent();
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
  setAngleControlEnabled("bgGradientAngle", enabled);
}

function setPickerDefault(textInput, colorInput, fallback) {
  if (!textInput || !colorInput) {
    return;
  }
  const candidate = textInput.value.trim() || fallback || "";
  const normalized = normalizeColor(candidate);
  if (normalized) {
    colorInput.value = normalized;
  }
}

function computeForegroundBaseColor(variant) {
  const gradientEnabled = Boolean(elements.gradientEnabled?.checked);
  if (gradientEnabled) {
    const from = elements.gradientFrom?.value.trim() || variant?.gradientFrom || "";
    const to = elements.gradientTo?.value.trim() || variant?.gradientTo || "";
    if (from || to) {
      return meanHex(from, to);
    }
  }
  const darkCandidate =
    elements.dark.value.trim() || variant?.dark || elements.dark.placeholder || "";
  return normalizeColor(darkCandidate);
}

function ensureBackgroundDefaults(variant) {
  if (backgroundModeIsTransparent()) {
    return;
  }
  if (!elements.light) {
    return;
  }
  const baseColor = computeForegroundBaseColor(variant);
  const opposite = invertHex(baseColor);
  if (opposite) {
    const lightValue = elements.light.value.trim();
    if (!lightValue || lightValue === state.autoLightValue) {
      elements.light.value = opposite;
      setPickerDefault(elements.light, elements.lightPicker, opposite);
      state.autoLightValue = opposite;
    }
  }

  if (elements.bgGradientEnabled?.checked) {
    let fallbackFrom = "";
    let fallbackTo = "";
    if (Boolean(elements.gradientEnabled?.checked)) {
      const fgFrom = elements.gradientFrom?.value.trim() || variant?.gradientFrom || "";
      const fgTo = elements.gradientTo?.value.trim() || variant?.gradientTo || "";
      fallbackFrom = invertHex(fgFrom);
      fallbackTo = invertHex(fgTo);
    }
    if (!fallbackFrom || !fallbackTo) {
      const fallback = elements.light.value.trim() || opposite;
      fallbackFrom = fallbackFrom || fallback;
      fallbackTo = fallbackTo || fallback;
    }
    if (fallbackFrom && elements.bgGradientFrom) {
      const current = elements.bgGradientFrom.value.trim();
      if (!current || current === state.autoBgGradientFrom) {
        elements.bgGradientFrom.value = fallbackFrom;
        setPickerDefault(elements.bgGradientFrom, elements.bgGradientFromPicker, fallbackFrom);
        state.autoBgGradientFrom = fallbackFrom;
      }
    }
    if (fallbackTo && elements.bgGradientTo) {
      const current = elements.bgGradientTo.value.trim();
      if (!current || current === state.autoBgGradientTo) {
        elements.bgGradientTo.value = fallbackTo;
        setPickerDefault(elements.bgGradientTo, elements.bgGradientToPicker, fallbackTo);
        state.autoBgGradientTo = fallbackTo;
      }
    }
  }
}

function renderAnimationVariantSelect(variants) {
  if (!elements.animationVariant) {
    return;
  }
  elements.animationVariant.innerHTML = "";
  variants.forEach((name) => {
    const option = document.createElement("option");
    option.value = name;
    option.textContent = name.replace(/-/g, " ");
    elements.animationVariant.appendChild(option);
  });
  elements.animationVariant.value = currentAnimationVariant();
}

async function loadAnimationConfig() {
  const config = await GetAnimationConfig();
  state.animationVariants = config.variants || [];
  state.animationDefaults = config.defaults || null;
  if (!state.selectedAnimationVariant || !state.animationVariants.includes(state.selectedAnimationVariant)) {
    state.selectedAnimationVariant = state.animationVariants[0] || "wave";
  }
  renderAnimationVariantSelect(state.animationVariants);
  renderAnimationGallery(state.animationVariants);
  updateAnimationPlaceholders();
  updateAnimationVariantSelection();
  updateAnimationVariantControls();
  updateAnimationDebug();
  setAnimationPlaceholder(
    "Animation preview",
    "Use the QR tab to set colors and data, then render a GIF here."
  );
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
    await loadAnimationConfig();
  } catch (error) {
    setAnimationStatus(error.message || String(error), "error");
  }
  try {
    await loadVariants();
  } catch (error) {
    setStatus(error.message || String(error), "error");
  }
  await refreshPreview();
}

[
  elements.data,
  elements.shape,
  elements.error,
  elements.scale,
  elements.border,
  elements.dark,
  elements.light,
  elements.radius,
  elements.gradientEnabled,
  elements.gradientFrom,
  elements.gradientTo,
  elements.gradientScope,
  elements.gradientFromStop,
  elements.gradientToStop,
  elements.bgGradientEnabled,
  elements.bgGradientFrom,
  elements.bgGradientTo,
  elements.bgGradientFromStop,
  elements.bgGradientToStop,
  elements.pngScale
].forEach((input) => {
  if (!input) {
    return;
  }
  input.addEventListener("input", () => {
    debounceRefresh();
    markAnimationDirty();
  });
});

[elements.gradientAngle, elements.bgGradientAngle].forEach((input) => {
  if (!input) {
    return;
  }
  input.addEventListener("input", () => {
    refreshPreview();
    markAnimationDirty();
  });
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
  markAnimationDirty();
});

elements.bgGradientEnabled?.addEventListener("change", () => {
  updateBackgroundGradientState();
  ensureBackgroundDefaults(state.variantMap.get(currentVariantName()));
  debounceRefresh();
  updateAnimationPreviewFrame();
  markAnimationDirty();
});

elements.backgroundModeOptions.forEach((option) => {
  option.addEventListener("change", () => {
    updateBackgroundGradientState();
    ensureBackgroundDefaults(state.variantMap.get(currentVariantName()));
    updatePreviewFrame();
    updateAnimationPreviewFrame();
    debounceRefresh();
    markAnimationDirty();
  });
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
    updateAngleControlColorsForInput(textInput.id);
    debounceRefresh();
    markAnimationDirty();
  });
  colorInput.addEventListener("input", () => {
    syncTextToPicker(textInput, colorInput);
    updateAngleControlColorsForInput(textInput.id);
    debounceRefresh();
    markAnimationDirty();
  });
});

[
  elements.gifFps,
  elements.gifScale,
  elements.gifFrames,
  elements.gifHold,
  elements.waveAmp,
  elements.wavePeriod,
  elements.floatAngle,
  elements.floatCycles
].forEach((input) => {
  if (!input) {
    return;
  }
  input.addEventListener("input", () => {
    markAnimationDirty();
  });
});

elements.animationVariant?.addEventListener("change", () => {
  selectAnimationVariant(elements.animationVariant.value);
});

elements.readableGif?.addEventListener("change", () => {
  updateAnimationPlaceholders();
  markAnimationDirty();
});

elements.renderAnimation?.addEventListener("click", refreshAnimationPreview);

elements.exportGif?.addEventListener("click", exportGif);

elements.animationAuto?.addEventListener("change", toggleAnimationAutoRender);

elements.animationInterval?.addEventListener("input", () => {
  if (elements.animationAuto?.checked) {
    restartAnimationAutoRender();
  }
});

setupAngleControls();
setupTabs();
init();
