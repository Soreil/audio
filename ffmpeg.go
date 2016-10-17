// Package audio provides ffmpeg bindings for reading audio file metadata
package audio

// #cgo pkg-config: libavcodec libavformat libavutil
// #cgo CFLAGS: -std=c11
// #include "ffmpeg.h"
import "C"
import (
	"io"
	"time"
	"unsafe"

	"github.com/Soreil/video/avio"
)

// Decoder wraps around internal state, all methods are called on this.
type Decoder struct {
	avio.Context
	avIOCtx     *C.struct_AVIOContext
	avFormatCtx *C.struct_AVFormatContext
}

func init() {
	C.av_register_all()
	C.avcodec_register_all()
	C.av_log_set_level(16)
}

// NewDecoder sets up a context for the file to use to probe for information
func NewDecoder(r io.ReadSeeker) (*Decoder, error) {
	ctx, err := avio.NewContext(&avio.Handlers{
		ReadPacket: r.Read,
		Seek:       r.Seek,
	})
	if err != nil {
		return nil, err
	}

	return &Decoder{
		Context: *ctx,
		// C types from different packages are not equal. Cast them.
		avIOCtx:     (*C.struct_AVIOContext)(ctx.AVIOContext()),
		avFormatCtx: (*C.struct_AVFormatContext)(ctx.AVFormatContext()),
	}, nil
}

// Duration returns the duration of the file
func (d *Decoder) Duration() time.Duration {
	return time.Duration(d.avFormatCtx.duration) * 1000
}

// AudioFormat returns the format string of the file's audio stream
func (d *Decoder) AudioFormat() string {
	c := C.get_codecContext(d.avFormatCtx)
	defer C.avcodec_close(c)
	if c == nil {
		return ""
	}
	return C.GoString(c.codec.name)
}

// Bitrate returns the bitrate in bits per second. For some files this will be
// absolute, for some an estimate.
func (d *Decoder) Bitrate() int64 {
	c := C.get_codecContext(d.avFormatCtx)
	defer C.avcodec_close(c)
	if c == nil || c.bit_rate == 0 {
		//This is an estimate, it might not be accurate!
		return int64(d.avFormatCtx.bit_rate)
	}
	return int64(c.bit_rate)
}

// HasImage return whether or not the file has album art in it
func (d *Decoder) HasImage() bool {
	return C.has_image(d.avFormatCtx) == 0
}

// Picture extracts attached image. This function will only work if the decoder
// was given enough data.
func (d *Decoder) Picture() []byte {
	img := C.retrieve_album_art(d.avFormatCtx)
	if img.size <= 0 || img.data == nil {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(img.data), img.size)
}
