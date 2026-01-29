export function resolveAutoLightValue({
  variantLight,
  currentLight,
  autoLightValue,
  fallbackLight
}) {
  const hasVariantLight = Boolean((variantLight || "").trim());
  const current = (currentLight || "").trim();
  const auto = (autoLightValue || "").trim();
  if (hasVariantLight) {
    if (current && auto && current === auto) {
      return { value: "", autoValue: "" };
    }
    return null;
  }
  if (!fallbackLight) {
    return null;
  }
  if (!current || (auto && current === auto)) {
    return { value: fallbackLight, autoValue: fallbackLight };
  }
  return null;
}
