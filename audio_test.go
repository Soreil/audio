package audio

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestConstruct(t *testing.T) {
	var files = []string{
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
		"traincrash.webm",
		"test.webm",
		"mpthreetest.mp3",
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
		d, err := dec.Duration()
		if err != nil {
			t.Log("Duration error:", err)
		} else {
			t.Log("Audio duration: ", d)
		}
		t.Log("Audio format: ", dec.AudioFormat())
		t.Log("Bitrate: ", dec.Bitrate()/1024, "kbps")
		t.Log("Image format: ", dec.ImageFormat())
		pic, err := dec.Picture()
		if err != nil {
			t.Log("Picture error : ", err)
		} else {
			t.Log("Picture length: ", fmt.Sprint(len(pic)/1024, "k"))
		}
	}
}
