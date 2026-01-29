package config

import "testing"

func TestEnabledVariantNamesFiltersDisabled(t *testing.T) {
	variants := map[string]Variant{
		"alpha": {Name: "alpha", Shape: "square", Dark: "#000000"},
		"beta":  {Name: "beta", Shape: "square", Dark: "#000000", Disabled: true},
		"zeta":  {Name: "zeta", Shape: "square", Dark: "#000000"},
	}
	names := EnabledVariantNames(variants)
	if len(names) != 2 {
		t.Fatalf("expected 2 enabled variants, got %d", len(names))
	}
	if names[0] != "alpha" || names[1] != "zeta" {
		t.Fatalf("expected sorted enabled names, got %v", names)
	}
}
