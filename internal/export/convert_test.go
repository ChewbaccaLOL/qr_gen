package export

import (
	"strings"
	"testing"
)

func TestCairoSVGScriptPDF(t *testing.T) {
	script, err := CairoSVGScript(FormatPDF, "in.svg", "out.pdf")
	if err != nil {
		t.Fatalf("expected script, got error: %v", err)
	}
	if script == "" {
		t.Fatalf("expected script")
	}
	if want := "svg2pdf"; !strings.Contains(script, want) {
		t.Fatalf("expected %s in script: %s", want, script)
	}
}

func TestCairoSVGScriptUnsupported(t *testing.T) {
	if _, err := CairoSVGScript("nope", "in.svg", "out.ps"); err == nil {
		t.Fatalf("expected error for unsupported format")
	}
}

// strings.Contains used above
