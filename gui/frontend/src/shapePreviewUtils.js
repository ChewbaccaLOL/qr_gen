export function resolveShapePreviewRadius(shape, radiusValue) {
  if (shape === "dot") {
    return 50;
  }
  if (shape !== "rounded") {
    return 0;
  }
  const numeric = Number.isFinite(radiusValue) ? radiusValue : 0;
  const clamped = Math.min(Math.max(numeric, 0), 0.5);
  return clamped * 100;
}

export function resolveShapePreviewGap(shape) {
  if (shape === "dot") {
    return 4;
  }
  return 0;
}
