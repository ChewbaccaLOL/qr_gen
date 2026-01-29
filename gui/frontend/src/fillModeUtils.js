export function resolveFillMode(hasGradient) {
  return hasGradient ? "gradient" : "solid";
}

export function isGradientFill(fillMode) {
  return fillMode === "gradient";
}

export function isBackgroundGradientFill(fillMode, backgroundMode) {
  return fillMode === "gradient" && backgroundMode !== "transparent";
}
