//The audio package provides a generic way to work with audio formats
package audio

import (
	"io"
	"time"
)

const (
	mp3  = "mp3"
	opus = "opus"
	ogg  = "vorbis"
	jpg  = "mjpeg"
	png  = "png"
	gif  = "gif"
)

type Audio interface {
	ExtractImage() ([]byte, error)
	Length() time.Duration
}

func Decode(r io.Reader) (Audio, string, error) {
	return nil, "", nil
}
