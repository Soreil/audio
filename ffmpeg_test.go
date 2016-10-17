package audio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConstruct(t *testing.T) {
	t.Parallel()

	files := [...]string{
		"exampleImage.mp3",
		"exampleTrue.mp3",
		"mpthreetest.mp3",
		"examplePNG.mp3",

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
	for i := range files {
		input := files[i]
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			t.Log("Filename: ", input)

			f, err := os.Open(filepath.Join("testdata", input))
			if err != nil {
				t.Fatal(err)
			}

			dec, err := NewDecoder(f)
			if err != nil {
				t.Fatal(err)
			}
			defer dec.Close()

			if fmt := dec.AudioFormat(); fmt != "" {
				t.Log("Audio duration: ", dec.Duration())
				t.Log("Audio format: ", dec.AudioFormat())
				t.Log("Bitrate: ", dec.Bitrate()/1000, "kbps")
			}
			if dec.HasImage() {
				pic := dec.Picture()
				t.Log("Picture length: ", len(pic)/1024, "k")
			}
		})
	}
}
