import assert from "node:assert/strict";
import { isBackgroundGradientFill, isGradientFill, resolveFillMode } from "./fillModeUtils.js";

function run() {
  assert.equal(resolveFillMode(true), "gradient", "maps true to gradient");
  assert.equal(resolveFillMode(false), "solid", "maps false to solid");
  assert.equal(isGradientFill("gradient"), true, "gradient fill is enabled");
  assert.equal(isGradientFill("solid"), false, "solid fill disables gradient");
  assert.equal(
    isBackgroundGradientFill("gradient", "solid"),
    true,
    "background gradient allowed when not transparent"
  );
  assert.equal(
    isBackgroundGradientFill("gradient", "transparent"),
    false,
    "background gradient disabled when transparent"
  );
}

run();
console.log("fillModeUtils tests passed");
