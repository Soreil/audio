package audio

import (
	"io"
	"os"
	"testing"
)

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

		"aacTest.mp4",

		"aacTest.aac",

		"itunes.m4a",
	}
	for _, input := range files {
		t.Log()
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
		t.Log("Audio duration: ", dec.Duration())
		if fmt := dec.AudioFormat(); fmt != "" {
			t.Log("Audio format: ", dec.AudioFormat())
			t.Log("Bitrate: ", dec.Bitrate()/1000, "kbps")
		}
		if fmt := dec.ImageFormat(); fmt != "" {
			t.Log("Image format: ", fmt)
			pic := dec.Picture()
			t.Log("Picture length: ", len(pic)/1024, "k")
		}
	}
}
