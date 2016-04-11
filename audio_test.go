package audio

import (
	"io"
	"os"
	"testing"
)

func init() {
}

func TestConstruct(t *testing.T) {
	var files = []string{
		"exampleFalse.mp3",
		"exampleImage.mp3",
		"exampleTrue.mp3",
		"mpthreetest.mp3",
		"examplePNG.mp3",

		"exampleFalse.ogg",
		"exampleJPG.ogg",
		"examplePNG.ogg",
		"exampleTrue.ogg",

		"exampleJPG.opus",
		"examplePNG.opus",
		"exampleTrue2.opus",
		"exampleTrue.opus",

		"traincrash.webm",
		"test.webm",
		"slam.webm",

		"aacTest.mp4",

		"aacTest.aac",

		"itunes.m4a",
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
		if fmt := dec.AudioFormat(); fmt != "" {
			t.Log("Audio duration: ", dec.Duration())
			t.Log("Audio format: ", dec.AudioFormat())
			t.Log("Bitrate: ", dec.Bitrate()/1000, "kbps")
		}
		if dec.HasImage() {
			t.Log("Image format: ", dec.imageFormat())
			pic := dec.Picture()
			t.Log("Picture length: ", len(pic)/1024, "k")
		}
		dec.Destroy()
	}
}
