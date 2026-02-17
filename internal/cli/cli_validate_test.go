package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "flow.mmd")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func TestValidate_OK(t *testing.T) {
	path := writeTemp(t, `flowchart TD
A[cmd]
B[cmd]
A --> B
`)
	var out, errOut bytes.Buffer
	code := Validate(path, &out, &errOut)
	if code != 0 {
		t.Fatalf("want 0 got %d; err=%s", code, errOut.String())
	}
	if got := out.String(); got == "" {
		t.Fatalf("expected some ok output, got empty")
	}
}

func TestValidate_Error(t *testing.T) {
	// undefined node reference (B is not defined)
	path := writeTemp(t, `flowchart TD
A[cmd]
A --> B
`)
	var out, errOut bytes.Buffer
	code := Validate(path, &out, &errOut)
	if code != 2 {
		t.Fatalf("want 2 got %d", code)
	}
	if errOut.String() == "" {
		t.Fatalf("expected error output")
	}
}
