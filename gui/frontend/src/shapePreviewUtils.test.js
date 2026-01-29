import assert from "node:assert/strict";
import { resolveShapePreviewGap, resolveShapePreviewRadius } from "./shapePreviewUtils.js";

function run() {
  assert.equal(resolveShapePreviewRadius("square", 0.3), 0, "square stays sharp");
  assert.equal(resolveShapePreviewRadius("dot", 0.3), 50, "dot is fully round");
  assert.equal(resolveShapePreviewRadius("rounded", 0.25), 25, "rounded maps to percent");
  assert.equal(resolveShapePreviewRadius("rounded", 0.9), 50, "rounded clamps to 50%");
  assert.equal(resolveShapePreviewRadius("rounded", -0.1), 0, "rounded clamps to 0%");
  assert.equal(resolveShapePreviewRadius("island-4", 0.3), 30, "island uses radius");
  assert.equal(resolveShapePreviewRadius("island-8", 0.3), 30, "island uses radius");
  assert.equal(resolveShapePreviewGap("square"), 0, "square modules touch");
  assert.equal(resolveShapePreviewGap("rounded"), 0, "rounded modules touch");
  assert.equal(resolveShapePreviewGap("dot"), 4, "dot modules remain separated");
}

run();
console.log("shapePreviewUtils tests passed");
