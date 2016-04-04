package audio

// #cgo pkg-config: libavcodec libavutil libavformat
// #cgo CFLAGS: -std=c11
/*

#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>
#include <libavutil/pixdesc.h>
#include <libavutil/avutil.h>
#include <libavutil/opt.h>
#include <libavformat/avformat.h>
#include <libavformat/avio.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

#define BUFFER_SIZE 4096

struct buffer_data {
	uint8_t *start_ptr; ///< start of buffer
    uint8_t *ptr; ///<current index in buffer
    size_t size; ///< size left in the buffer
    size_t len; ///< size of the buffer
};

static int read_packet(void *opaque, uint8_t *buf, int buf_size)
{
    struct buffer_data *bd = (struct buffer_data *)opaque;
    buf_size = FFMIN(buf_size, bd->size);
    // copy internal buffer data to buf
    memcpy(buf, bd->ptr, buf_size);
    bd->ptr  += buf_size;
    bd->size -= buf_size;
    return buf_size;
}
static int64_t seek(void *opaque, int64_t offset, int whence)
{
	if (whence == AVSEEK_SIZE) {
        return -1; // "size of my handle in bytes UNIMPLEMENTED"
	}

    struct buffer_data *bd = (struct buffer_data *)opaque;

	if (whence == SEEK_CUR) { // relative to start of file
		bd->ptr += offset;
	//	bd->size -= offset;
    }
//	if (whence == SEEK_END) { // relative to end of file
//        bd->ptr = bd->start_ptr+bd->len + offset;
//	//	bd->size = bd->len+offset;
//    }
//	if (whence == SEEK_SET) { // relative to start of file
//		bd->ptr = bd->start_ptr+offset;
//	//	bd->size = offset;
//	}

	return bd->len-bd->size;
}

AVFormatContext * create_context(unsigned char *opaque,size_t len)
{
	unsigned char *buffer = (unsigned char*)av_malloc(BUFFER_SIZE);

	struct buffer_data bd = {0};
	bd.start_ptr = opaque;
	bd.ptr = opaque;
	bd.size = len;
	bd.len = len;

	AVIOContext *ioCtx = avio_alloc_context(buffer,BUFFER_SIZE,0,&bd,&read_packet,NULL,&seek);

	AVFormatContext * ctx = avformat_alloc_context();
	if (!ctx) {
		return NULL;
	}

	//Set up context to read from memory
	ctx->pb = ioCtx;

	int err = avformat_open_input(&ctx, NULL, NULL, NULL);
	if (err < 0) {
		return NULL;
	}

	err = avformat_find_stream_info(ctx,NULL);
	if (err < 0) {
		return NULL;
	}

 //   av_dump_format(ctx, 0, NULL, 0);

	return ctx;
}

AVCodec * get_codec(AVFormatContext *ctx,enum AVMediaType strmType) {
	AVCodec * codec = NULL;
	av_find_best_stream(ctx, strmType, -1, -1, &codec, 0);
	return codec;
}

//Extract embedded images
AVPacket retrieve_album_art(AVFormatContext *ctx) {
	AVPacket err;
	// read the format headers
	if (ctx->iformat->read_header(ctx) < 0) {
		return err;
	}

	// find the first attached picture, if available
	for (int i = 0; i < ctx->nb_streams; i++) {
		if (ctx->streams[i]->disposition & AV_DISPOSITION_ATTACHED_PIC) {
			return ctx->streams[i]->attached_pic;
		}
	}
	return err;
}
*/
import "C"
import (
	"errors"
	"io"
	"io/ioutil"
	"time"
	"unsafe"
)

type Decoder struct {
	ctx *C.AVFormatContext
}

func init() {
	C.av_register_all()
	C.avcodec_register_all()
	C.av_log_set_level(32)
}

func byteSliceToCArray(byteSlice []byte) unsafe.Pointer {
	var array = unsafe.Pointer(C.calloc(C.size_t(len(byteSlice)), 1))
	var arrayptr = uintptr(array)

	for i := 0; i < len(byteSlice); i++ {
		*(*C.uchar)(unsafe.Pointer(arrayptr)) = C.uchar(byteSlice[i])
		arrayptr++
	}

	return array
}

//Sets up a context for the file to use to probe for information
func NewDecoder(r io.Reader) (Decoder, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return Decoder{}, err
	}
	if len(data) <= 0 {
		return Decoder{}, errors.New("No input data provided")
	}

	if ctx := C.create_context((*C.uchar)(byteSliceToCArray(data)), C.size_t(len(data))); ctx != nil {
		//if ctx := C.create_context((*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data))); ctx != nil {
		return Decoder{ctx}, nil
	} else {
		return Decoder{}, errors.New("Failed to create decoder context")
	}
}

//TODO:C code is broken for formats other than mp3, will need manual calculation
func (d Decoder) Duration() (time.Duration, error) {
	if d.ctx.duration == C.AV_NOPTS_VALUE {
		return 0, errors.New("Context has no duration set")
	}
	return time.Duration(d.ctx.duration) * 1000, nil

}

//Get audio format string
func (d Decoder) AudioFormat() string {
	c := C.get_codec(d.ctx, C.AVMEDIA_TYPE_AUDIO)
	if c == nil {
		return ""
	}
	return C.GoString(c.name)
}

//Get image format string
func (d Decoder) ImageFormat() string {
	c := C.get_codec(d.ctx, C.AVMEDIA_TYPE_VIDEO)
	if c == nil {
		return ""
	}
	fmt := C.GoString(c.name)
	if fmt == "mjpeg" {
		return "jpeg"
	} else {
		return fmt
	}
}

//Extract raw image
func (d Decoder) Picture() ([]byte, error) {
	img := C.retrieve_album_art(d.ctx)
	if img.size <= 0 {
		return nil, errors.New("Failed to extract picture")
	}
	return C.GoBytes(unsafe.Pointer(img.data), img.size), nil
}
