package resample

import (
	"math"
	"testing"
)

func TestResampleLength(t *testing.T) {
	inRate, outRate := 44100, 48000
	r, err := New(inRate, outRate)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer r.Close()

	in := make([]float32, inRate)
	out1, err := r.Convert(in)
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	out2, err := r.Convert(nil)
	if err != nil {
		t.Fatalf("flush returned error: %v", err)
	}
	out := append(out1, out2...)

	expected := int(math.Round(float64(len(in)) * float64(outRate) / float64(inRate)))
	if diff := int(math.Abs(float64(len(out) - expected))); diff > 1 {
		t.Fatalf("unexpected output length: got %d want %d", len(out), expected)
	}
}

func TestConvertAfterClose(t *testing.T) {
	r, err := New(44100, 48000)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	r.Close()
	if _, err := r.Convert([]float32{0}); err == nil {
		t.Fatalf("expected error after Close")
	}
}
