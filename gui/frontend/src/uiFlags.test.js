import assert from "node:assert/strict";
import { ANIMATION_DEBUG_ENABLED } from "./uiFlags.js";

function run() {
  assert.equal(ANIMATION_DEBUG_ENABLED, false, "animation debug is disabled by default");
}

run();
console.log("uiFlags tests passed");
