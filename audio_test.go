package audio

import (
	"io"
	"os"
	"testing"
)

func TestConstruct(t *testing.T) {
	const input = "example.mp3"
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := NewDecoder(io.Reader(f))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dec.AudioFormat())
	t.Log(dec.Duration())
}
