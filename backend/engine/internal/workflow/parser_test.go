package workflow

import "testing"

func TestParseYAML(t *testing.T) {
	raw := []byte(`version: v1
name: image-classify
triggers:
  - type: manual
steps:
  - id: decode
    type: wasm.decode
    input:
      codec: jpeg
  - id: infer
    type: ai.onnx
`)

	def, err := ParseYAML(raw)
	if err != nil {
		t.Fatalf("parse workflow: %v", err)
	}

	if def.Name != "image-classify" {
		t.Fatalf("unexpected name: %s", def.Name)
	}

	if len(def.Steps) != 2 {
		t.Fatalf("unexpected step count: %d", len(def.Steps))
	}
}

func TestParseYAMLValidationError(t *testing.T) {
	raw := []byte(`version: v1
name: bad
triggers:
  - type: manual
steps:
  - id: duplicate
    type: wasm.decode
  - id: duplicate
    type: ai.onnx
`)

	_, err := ParseYAML(raw)
	if err == nil {
		t.Fatal("expected duplicate id validation error")
	}
}
