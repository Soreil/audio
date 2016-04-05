package audio

// #cgo pkg-config: libavcodec libavformat libavutil
// #cgo CFLAGS: -std=c11
/*

#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
#include <libavformat/avio.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

//Placeholder, 256K is enough to make Opus with JPG succeed, 512K Opus with PNG.
//#define BUFFER_SIZE 524288
#define BUFFER_SIZE 4096

struct buffer_data {
	uint8_t *start_ptr; ///< start of buffer
    uint8_t *ptr_pos; ///<current index in buffer
    size_t size_left; ///< size left in the buffer
    size_t size; ///< size of the buffer
};

static int read_packet(void *opaque, uint8_t *buf, int buf_size)
{
    struct buffer_data *bd = (struct buffer_data *)opaque;
	if( buf_size > bd->size_left) {
		buf_size = bd->size_left;
	}

    // copy internal buffer data to buf
    memcpy(buf, bd->ptr_pos, buf_size);
    bd->ptr_pos  += buf_size;
    bd->size_left -= buf_size;
    return buf_size;
}
static int64_t seek(void *opaque, int64_t offset, int whence)
{
    struct buffer_data *bd = (struct buffer_data *)opaque;

	if (whence == AVSEEK_SIZE) {
        return bd->size; // "size of my handle in bytes"
        return -1; // "size of my handle in bytes UNIMPLEMENTED"
	}
	if (whence == SEEK_CUR) { // relative to start of file
		bd->ptr_pos += offset;
		bd->size_left -= offset;
    }
	if (whence == SEEK_END) { // relative to end of file
        bd->ptr_pos = bd->start_ptr+bd->size + offset;
		bd->size_left = bd->size+offset;
    }
	if (whence == SEEK_SET) { // relative to start of file
		bd->ptr_pos = bd->start_ptr+offset;
		bd->size_left = bd->size-offset;
	}

	return bd->size-bd->size_left;
}

AVFormatContext * create_context(unsigned char *opaque,size_t len)
{
	unsigned char *buffer = (unsigned char*)av_malloc(BUFFER_SIZE);

	struct buffer_data bd = {0};
	bd.start_ptr = opaque;
	bd.ptr_pos = opaque;
	bd.size_left = len;
	bd.size = len;

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
	//TODO(sjon): This is changed in FFMPEG 3.0 but should behave the same
	//ctx->max_analyze_duration = 100000000;
	err = avformat_find_stream_info(ctx,NULL);
	if (err < 0) {
		return NULL;
	}

	return ctx;
}

AVCodecContext * get_codecContext(AVFormatContext *ctx,enum AVMediaType strmType) {
	AVCodec * codec = NULL;

	int strm = av_find_best_stream(ctx, strmType, -1, -1, &codec, 0);
	if (strm < 0 || strm == AVERROR_STREAM_NOT_FOUND){
		return NULL;
	}
	AVCodecContext * codecCtx = ctx->streams[strm]->codec;
	int err = avcodec_open2(codecCtx, codec, NULL);
	if (err < 0 ) {
		return NULL;
	}
	return codecCtx;
}

//Doesn't seem to produce any nice results sadly
int64_t get_duration(AVFormatContext *ctx) {
	int strm = av_find_best_stream(ctx, AVMEDIA_TYPE_AUDIO, -1, -1, NULL, 0);
	if (strm < 0|| strm == AVERROR_STREAM_NOT_FOUND ){
		return 0;
	}
	return ctx->streams[strm]->duration;
}

//Extract embedded images
AVPacket retrieve_album_art(AVFormatContext *ctx) {
	AVPacket err;

	// find the first attached picture, if available
	for (int i = 0; i < ctx->nb_streams; i++) {
		if (ctx->streams[i]->disposition & AV_DISPOSITION_ATTACHED_PIC) {
			return ctx->streams[i]->attached_pic;
		}
	}
	return err;
}

int has_image(AVFormatContext *ctx) {
	// find the first attached picture, if available
	for (int i = 0; i < ctx->nb_streams; i++) {
		if (ctx->streams[i]->disposition & AV_DISPOSITION_ATTACHED_PIC) {
			return 0;
		}
	}
	return 1;
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

//Wrapper around internal state, all methods are called on this.
type Decoder struct {
	ctx *C.AVFormatContext
}

func init() {
	C.av_register_all()
	C.avcodec_register_all()
	C.av_log_set_level(48)
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

//Sets up a context for the file to use to probe for information.
func NewDecoder(r io.Reader) (Decoder, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return Decoder{}, err
	}
	if len(data) <= 0 {
		return Decoder{}, errors.New("No input data provided")
	}

	if ctx := C.create_context((*C.uchar)(byteSliceToCArray(data)), C.size_t(len(data))); ctx != nil {
		return Decoder{ctx: ctx}, nil
	} else {
		return Decoder{}, errors.New("Failed to create decoder context")
	}
}

//Gets duration of audio track.
func (d Decoder) Duration() time.Duration {
	return time.Duration(d.ctx.duration) * 1000
}

//Get audio format string
func (d Decoder) AudioFormat() string {
	c := C.get_codecContext(d.ctx, C.AVMEDIA_TYPE_AUDIO)
	if c == nil {
		return ""
	}
	return C.GoString(c.codec.name)
}

//Returns bitrate in bps.
func (d Decoder) Bitrate() int64 {
	c := C.get_codecContext(d.ctx, C.AVMEDIA_TYPE_AUDIO)
	if c == nil {
		return int64(d.ctx.bit_rate)
	}
	if c.bit_rate != 0 {
		return int64(c.bit_rate)
	} else {
		//This is an estimate, it might not be accurate!
		return int64(d.ctx.bit_rate)
	}
}

//Whether or not the file has album art in it
func (d Decoder) hasImage() bool {
	return C.has_image(d.ctx) == 0
}

//Get image format string TO BE REMOVED
func (d Decoder) ImageFormat() string {
	c := C.get_codecContext(d.ctx, C.AVMEDIA_TYPE_VIDEO)
	if c == nil {
		return ""
	}
	fmt := C.GoString(c.codec.name)
	if fmt == "mjpeg" {
		return "jpeg"
	} else {
		return fmt
	}
}

//Extract attached image
func (d Decoder) Picture() []byte {
	img := C.retrieve_album_art(d.ctx)
	if img.size <= 0 || img.data == nil {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(img.data), img.size)
}
