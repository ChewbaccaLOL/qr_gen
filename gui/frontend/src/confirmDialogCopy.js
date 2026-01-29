export function buildDeleteVariantDialogCopy(name) {
  const trimmed = (name ?? "").trim();
  const message = trimmed
    ? `Delete the custom variant "${trimmed}"?`
    : "Delete this custom variant?";

  return {
    title: "Delete custom variant",
    message,
    note: "This action cannot be undone.",
    confirmLabel: "Delete variant",
    cancelLabel: "Cancel"
  };
}
