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

extern int readCallBackAudio(void*, uint8_t*, int);
extern int writeCallBackAudio(void*, uint8_t*, int);
extern int64_t seekCallBackAudio(void*, int64_t, int);

static inline AVFormatContext * create_context(AVFormatContext *ctx)
{
	char errstringbuf[1024];
	//int64_t read = 0;
	//read = read_packet(&bd,buffer,BUFFER_SIZE);
	//seek(&bd,0,SEEK_SET);

	//AVProbeData probeData;
	//probeData.buf = buffer;
	//probeData.buf_size = read-1;
	//probeData.filename = "";

	//// Determine the input-format:
	//ctx->iformat = av_probe_input_format(&probeData, 0);
	//
	//ctx->flags = AVFMT_FLAG_CUSTOM_IO;

	int err = avformat_open_input(&ctx, NULL, NULL, NULL);
	if (err < 0) {
		av_strerror(err,errstringbuf,1024);
		fprintf(stderr,"%s\n",errstringbuf);
		return NULL;
	}
	err = avformat_find_stream_info(ctx,NULL);
	if (err < 0) {
		av_strerror(err,errstringbuf,1024);
		fprintf(stderr,"%s\n",errstringbuf);
		return NULL;
	}

	return ctx;
}

static inline AVCodecContext * get_codecContext(AVFormatContext *ctx,enum AVMediaType strmType) {
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
static inline int64_t get_duration(AVFormatContext *ctx) {
	int strm = av_find_best_stream(ctx, AVMEDIA_TYPE_AUDIO, -1, -1, NULL, 0);
	if (strm < 0|| strm == AVERROR_STREAM_NOT_FOUND ){
		return 0;
	}
	return ctx->streams[strm]->duration;
}

//Extract embedded images
static inline AVPacket retrieve_album_art(AVFormatContext *ctx) {
	AVPacket err;

	// find the first attached picture, if available
	for (int i = 0; i < ctx->nb_streams; i++) {
		if (ctx->streams[i]->disposition & AV_DISPOSITION_ATTACHED_PIC) {
			return ctx->streams[i]->attached_pic;
		}
	}
	return err;
}

static inline int has_image(AVFormatContext *ctx) {
	// find the first attached picture, if available
	for (int i = 0; i < ctx->nb_streams; i++) {
		if (ctx->streams[i]->disposition & AV_DISPOSITION_ATTACHED_PIC) {
			return 0;
		}
	}
	return 1;
}

static inline void destroy(AVFormatContext *ctx) {
	av_free(ctx->pb->buffer);
	ctx->pb->buffer = NULL;
	av_free(ctx->pb);
	avformat_close_input(&ctx);
}
*/
import "C"
import (
	"errors"
	"fmt"
	"io"
	"time"
	"unsafe"
)

//Decoder wraps around internal state, all methods are called on this.
type Decoder struct {
	ctx   *C.AVFormatContext
	ioCtx *avIOContext
}

var (
	IO_BUFFER_SIZE int = 4096
)

func init() {
	C.av_register_all()
	C.avcodec_register_all()
	C.av_log_set_level(16)
}

/////////////////////////////////////
// Functions prototypes for custom IO. Implement necessary prototypes and pass instance pointer to NewAVIOContext.
//
// E.g.:
// 	func gridFsReader() ([]byte, int) {
// 		... implementation ...
//		return data, length
//	}
//
//	avoictx := NewAVIOContext(ctx, &AVIOHandlers{ReadPacket: gridFsReader})
type avIOHandlers struct {
	ReadPacket  func([]byte) (int, error)
	WritePacket func([]byte) (int, error)
	Seek        func(int64, int) (int64, error)
}

// Global map of AVIOHandlers
// one handlers struct per format context. Using ctx pointer address as a key.
var handlersMap map[uintptr]*avIOHandlers

type avIOContext struct {
	// avAVIOContext *_Ctype_AVIOContext
	avAVIOContext *C.struct_AVIOContext
	handlerKey    uintptr
}

// AVIOContext constructor. Use it only if You need custom IO behaviour!
func newAVIOContext(ctx *C.AVFormatContext, handlers *avIOHandlers) (*avIOContext, error) {
	this := &avIOContext{}

	buffer := (*C.uchar)(C.av_malloc(C.size_t(IO_BUFFER_SIZE)))

	if buffer == nil {
		return nil, errors.New("unable to allocate buffer")
	}

	// we have to explicitly set it to nil, to force library using default handlers
	var ptrRead, ptrWrite, ptrSeek *[0]byte = nil, nil, nil

	if handlers != nil {
		if handlersMap == nil {
			handlersMap = make(map[uintptr]*avIOHandlers)
		}

		handlersMap[uintptr(unsafe.Pointer(ctx))] = handlers
		this.handlerKey = uintptr(unsafe.Pointer(ctx))
	}

	if handlers.ReadPacket != nil {
		ptrRead = (*[0]byte)(C.readCallBackAudio)
	}

	if handlers.WritePacket != nil {
		ptrWrite = (*[0]byte)(C.writeCallBackAudio)
	}

	if handlers.Seek != nil {
		ptrSeek = (*[0]byte)(C.seekCallBackAudio)
	}

	if this.avAVIOContext = C.avio_alloc_context(buffer, C.int(IO_BUFFER_SIZE), 0, unsafe.Pointer(ctx), ptrRead, ptrWrite, ptrSeek); this.avAVIOContext == nil {
		return nil, errors.New("unable to initialize avio context")
	}

	return this, nil
}

func (this *avIOContext) Free() {
	delete(handlersMap, this.handlerKey)
}

//export readCallBackAudio
func readCallBackAudio(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	handlers, found := handlersMap[uintptr(opaque)]
	if !found {
		panic(fmt.Sprintf("No handlers instance found, according pointer: %v", opaque))
	}

	if handlers.ReadPacket == nil {
		panic("No reader handler initialized")
	}
	s := (*[1 << 30]byte)(unsafe.Pointer(buf))[:buf_size:buf_size]
	n, err := handlers.ReadPacket(s)
	if err != nil {
		return -1
	}
	return C.int(n)
}

//export writeCallBackAudio
func writeCallBackAudio(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	handlers, found := handlersMap[uintptr(opaque)]
	if !found {
		panic(fmt.Sprintf("No handlers instance found, according pointer: %v", opaque))
	}

	if handlers.WritePacket == nil {
		panic("No writer handler initialized.")
	}

	n, err := handlers.WritePacket(C.GoBytes(unsafe.Pointer(buf), buf_size))
	if err != nil {
		return -1
	}
	return C.int(n)
}

//export seekCallBackAudio
func seekCallBackAudio(opaque unsafe.Pointer, offset C.int64_t, whence C.int) C.int64_t {
	handlers, found := handlersMap[uintptr(opaque)]
	if !found {
		panic(fmt.Sprintf("No handlers instance found, according pointer: %v", opaque))
	}

	if handlers.Seek == nil {
		panic("No seek handler initialized.")
	}

	n, err := handlers.Seek(int64(offset), int(whence))
	if err != nil {
		return -1
	}
	return C.int64_t(n)
}

/////////////////////////////////////

var DecoderError = errors.New("Failed to create decoder context")

//NewDecoder sets up a context for the file to use to probe for information.
func NewDecoder(r io.ReadSeeker) (Decoder, error) {
	ctx := C.avformat_alloc_context()
	avioCtx, err := newAVIOContext(ctx, &avIOHandlers{ReadPacket: r.Read, Seek: r.Seek})
	if err != nil {
		panic(err)
	}
	ctx.pb = avioCtx.avAVIOContext
	if ctx = C.create_context(ctx); ctx != nil {
		return Decoder{ctx: ctx, ioCtx: avioCtx}, nil
	}
	avioCtx.Free()
	return Decoder{}, DecoderError
}

//Duration gets the duration of the file.
func (d Decoder) Duration() time.Duration {
	return time.Duration(d.ctx.duration) * 1000
}

//AudioFormat gets format string
func (d Decoder) AudioFormat() string {
	c := C.get_codecContext(d.ctx, C.AVMEDIA_TYPE_AUDIO)
	defer C.avcodec_close(c)
	if c == nil {
		return ""
	}
	return C.GoString(c.codec.name)
}

//Bitrate returns the bitrate in bits per second. For some files this will be absolute, for some an estimate.
func (d Decoder) Bitrate() int64 {
	c := C.get_codecContext(d.ctx, C.AVMEDIA_TYPE_AUDIO)
	defer C.avcodec_close(c)
	if c == nil || c.bit_rate == 0 {
		//This is an estimate, it might not be accurate!
		return int64(d.ctx.bit_rate)
	}
	return int64(c.bit_rate)
}

//HasImage return whether or not the file has album art in it
func (d Decoder) HasImage() bool {
	return C.has_image(d.ctx) == 0
}

//Get image format string TO BE REMOVED
func (d Decoder) imageFormat() string {
	c := C.get_codecContext(d.ctx, C.AVMEDIA_TYPE_VIDEO)
	defer C.avcodec_close(c)
	if c == nil {
		return ""
	}
	fmt := C.GoString(c.codec.name)
	if fmt == "mjpeg" {
		return "jpeg"
	}
	return fmt
}

//Picture extracts attached image. This function will only work if the decoder was given enough data.
func (d Decoder) Picture() []byte {
	img := C.retrieve_album_art(d.ctx)
	if img.size <= 0 || img.data == nil {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(img.data), img.size)
}

//Destroy frees the decoder, it should not be used after this point with a NewDecoder call.
func (d *Decoder) Destroy() {
	C.destroy(d.ctx)
	d.ctx = nil
	d = nil
}
