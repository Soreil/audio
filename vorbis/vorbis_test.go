package vorbis

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"testing"
)

const dataDir = "inputData/"

type testCase struct {
	isVorbis      bool
	inputFilename string
}

var cases = []testCase{
	{false, "exampleFalse.ogg"},
	{true, "exampleTrue.ogg"},
	{true, "exampleTrue2.ogg"},
	{true, "exampleJPG.ogg"},
	{true, "examplePNG.ogg"},
}

func TestVorbis(t *testing.T) {
	if err := os.Chdir(dataDir); err != nil {
		if err := os.Mkdir(dataDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(dataDir); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range cases {
		if IsVorbis(test.inputFilename) == test.isVorbis {
			if b, err := ExtractImage(test.inputFilename); err == nil {
				t.Log(test.inputFilename, "has an image inside of it")
				_, inFmt, err := image.DecodeConfig(bytes.NewReader(b))
				if err != nil {
					t.Fatal(err)
				}
				t.Log("Input format:", inFmt)
			} else {
				t.Log(err, test)
			}
			t.Log("LENGTH:", VorbisDuration(test.inputFilename))
		} else {
			t.Log("Vorbisness is not what was expected, expected: ", test.isVorbis, " got: ", !test.isVorbis)
		}
	}
}
