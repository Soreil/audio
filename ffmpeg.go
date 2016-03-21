package audio

// #cgo pkg-config: libavcodec libavutil libavformat
/*

#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>
#include <libavutil/pixdesc.h>
#include <libavutil/avutil.h>
#include <libavformat/avformat.h>
#include <stdio.h>

#define BUFFER_SIZE 4096

struct buffer_data {
    uint8_t *ptr;
    size_t size; ///< size left in the buffer
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

AVCodecContext * create_context(unsigned char *opaque,size_t len)
{
	unsigned char *buffer = (unsigned char*)av_malloc(BUFFER_SIZE+FF_INPUT_BUFFER_PADDING_SIZE);

	struct buffer_data bd = {0};
	bd.ptr = opaque;
	bd.size = len;

	//Allocate avioContext
	AVIOContext *ioCtx = avio_alloc_context(buffer,BUFFER_SIZE,0,&bd,&read_packet,NULL,NULL);

	AVFormatContext * ctx = avformat_alloc_context();

	//Set up context to read from memory
	ctx->pb = ioCtx;

	//open takes a fake filename when the context pb field is set up
	int err = avformat_open_input(&ctx, "dummyFileName", NULL, NULL);
	if (err < 0) {
		return NULL;
	}

	err = avformat_find_stream_info(ctx,NULL);
	if (err < 0) {
		return NULL;
	}

	AVCodec * codec = NULL;
	int strm = av_find_best_stream(ctx, AVMEDIA_TYPE_AUDIO, -1, -1, &codec, 0);

	AVCodecContext * codecCtx = ctx->streams[strm]->codec;
	err = avcodec_open2(codecCtx, codec, NULL);
	if (err < 0) {
		return NULL;
	}
	return codecCtx;
}

AVCodec * get_codec(AVCodecContext *ctx,enum AVMediaType strmType) {
	int err = 0;
	AVCodec * codec = NULL;
	int strm = av_find_best_stream(ctx, strmType, -1, -1, &codec, 0);

	AVCodecContext * codecCtx = ctx->streams[strm]->codec;
	err = avcodec_open2(codecCtx, codec, NULL);
	if (err < 0) {
		return NULL;
	}
	return codec;
}

//Extract embedded images
AVPacket retrieve_album_art(AVCodecContext *ctx) {
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
	ctx *C.AVCodecContext
}

func init() {
	C.av_register_all()
	C.avcodec_register_all()
}

//Sets up a context for the file to use to probe for information
func NewDecoder(r io.Reader) (Decoder, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if ctx := C.create_context((*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data))); ctx != nil {
		return Decoder{ctx}
	} else {
		return nil, errors.New("Failed to create decoder context")
	}
}

func (d Decoder) Duration() time.Duration {
	return time.Duration(d.ctx.duration) * 100
}

func (d Decoder) AudioFormat() string {
	c := C.get_codec(d.ctx, C.AVMEDIA_TYPE_AUDIO)
	if c == nil {
		return ""
	}
	return C.GoString(c.name)
}

func (d Decoder) Picture() ([]byte, error) {
	img := C.retrieve_album_art(d.ctx)
	if img == nil || img.size <= 0 {
		return nil, errors.New("Failed to extract picture")
	}
	return C.GoBytes(unsafe.Pointer(pkt.data), pkt.size), nil
}
