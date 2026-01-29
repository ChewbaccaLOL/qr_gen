import assert from "node:assert/strict";
import { resolveAutoLightValue } from "./backgroundDefaults.js";

function run() {
  assert.deepEqual(
    resolveAutoLightValue({
      variantLight: "#ffffff",
      currentLight: "#000000",
      autoLightValue: "#000000",
      fallbackLight: "#123456"
    }),
    { value: "", autoValue: "" },
    "clears auto light when variant supplies a background"
  );

  assert.equal(
    resolveAutoLightValue({
      variantLight: "#ffffff",
      currentLight: "",
      autoLightValue: "#000000",
      fallbackLight: "#123456"
    }),
    null,
    "keeps empty override when variant supplies a background"
  );

  assert.deepEqual(
    resolveAutoLightValue({
      variantLight: "",
      currentLight: "",
      autoLightValue: "",
      fallbackLight: "#abcdef"
    }),
    { value: "#abcdef", autoValue: "#abcdef" },
    "auto-fills when variant has no background"
  );

  assert.equal(
    resolveAutoLightValue({
      variantLight: null,
      currentLight: "#101010",
      autoLightValue: "#abcdef",
      fallbackLight: "#abcdef"
    }),
    null,
    "respects explicit light overrides"
  );

  assert.deepEqual(
    resolveAutoLightValue({
      variantLight: null,
      currentLight: "#abcdef",
      autoLightValue: "#abcdef",
      fallbackLight: "#fedcba"
    }),
    { value: "#fedcba", autoValue: "#fedcba" },
    "updates stale auto value when fallback changes"
  );
}

run();
console.log("backgroundDefaults tests passed");
