package opus

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"testing"
)

const dataDir = "inputData/"

type testCase struct {
	isOpus        bool
	inputFilename string
}

var cases = []testCase{
	{false, "exampleFalse.opus"},
	{true, "exampleTrue.opus"},
	{true, "exampleTrue2.opus"},
	{true, "exampleJPG.opus"},
	{true, "examplePNG.opus"},
}

func TestOpus(t *testing.T) {
	if err := os.Chdir(dataDir); err != nil {
		if err := os.Mkdir(dataDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(dataDir); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range cases {
		if IsOpus(test.inputFilename) == test.isOpus {
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
			fmt.Println("LENGTH:", OpusDuration(test.inputFilename))
		} else {
			t.Log("Opusness is not what was expected, expected: ", test.isOpus, " got: ", !test.isOpus)
		}
	}
}
