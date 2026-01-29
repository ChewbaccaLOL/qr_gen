import assert from "node:assert/strict";
import { buildDeleteVariantDialogCopy } from "./confirmDialogCopy.js";

function run() {
  const named = buildDeleteVariantDialogCopy("Midnight");
  assert.equal(named.title, "Delete custom variant", "uses consistent title");
  assert.equal(
    named.message,
    "Delete the custom variant \"Midnight\"?",
    "includes the variant name when provided"
  );
  assert.equal(named.confirmLabel, "Delete variant", "uses delete button label");
  assert.equal(named.cancelLabel, "Cancel", "uses cancel label");
  assert.equal(named.note, "This action cannot be undone.", "includes warning note");

  const unnamed = buildDeleteVariantDialogCopy("");
  assert.equal(unnamed.message, "Delete this custom variant?", "falls back for empty names");
}

run();
console.log("confirmDialogCopy tests passed");
