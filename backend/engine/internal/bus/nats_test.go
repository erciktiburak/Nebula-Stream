package bus

import "testing"

func TestDefaultNATSURL(t *testing.T) {
  if DefaultNATSURL == "" {
    t.Fatal("default URL must not be empty")
  }
}
