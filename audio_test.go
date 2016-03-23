package audio

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestConstruct(t *testing.T) {
	var files = []string{
		"detodos.opus",
		"ehren-paper_lights-64.opus",
		"exampleFalse.mp3",
		"exampleFalse.ogg",
		"exampleImage.mp3",
		"exampleJPG.ogg",
		"exampleJPG.opus",
		"examplePNG.mp3",
		"examplePNG.ogg",
		"examplePNG.opus",
		"exampleTrue2.opus",
		"exampleTrue.mp3",
		"exampleTrue.ogg",
		"exampleTrue.opus",
	}
	for _, input := range files {
		t.Log("Filename: ", input)
		f, err := os.Open(input)
		if err != nil {
			t.Fatal(err)
		}
		dec, err := NewDecoder(io.Reader(f))
		if err != nil {
			t.Log("Failed to create decoder", err)
			continue
		}
		t.Log("Audio format: ", dec.AudioFormat())
		t.Log("Image format: ", dec.ImageFormat())
		t.Log("Audio duration: ", dec.Duration())
		pic, err := dec.Picture()
		if err != nil {
			t.Log("Picture error : ", err)
		} else {
			t.Log("Picture length: ", fmt.Sprint(len(pic)))
		}
	}
}
