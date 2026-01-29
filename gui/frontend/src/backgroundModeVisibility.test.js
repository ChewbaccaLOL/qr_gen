import assert from "node:assert/strict";
import { sectionsForBackgroundMode } from "./backgroundModeVisibility.js";

function run() {
  assert.deepEqual(
    sectionsForBackgroundMode("solid"),
    { showForeground: true, showBackground: true },
    "solid shows both sections"
  );
  assert.deepEqual(
    sectionsForBackgroundMode("transparent"),
    { showForeground: true, showBackground: false },
    "transparent hides background section"
  );
  assert.deepEqual(
    sectionsForBackgroundMode("cutout"),
    { showForeground: false, showBackground: true },
    "cutout hides foreground section"
  );
  assert.deepEqual(
    sectionsForBackgroundMode("unknown"),
    { showForeground: true, showBackground: true },
    "unknown defaults to showing both"
  );
}

run();
console.log("backgroundModeVisibility tests passed");
