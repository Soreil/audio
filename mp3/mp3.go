package mp3

//#cgo pkg-config: libavcodec libavformat
/*
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <stdlib.h>
#include <stdint.h>

int64_t mp3_duration(const char * url) {
	AVFormatContext *ctx = NULL;
	int err = avformat_open_input(&ctx,url,NULL,NULL);
	if (err < 0) {
		return -1;
	}
	err = avformat_find_stream_info(ctx,NULL);
	if (err < 0) {
		return -1;
	}
	return ctx->duration;
}

int mp3_check(const char * url) {
	AVFormatContext *ctx = NULL;
	int err = avformat_open_input(&ctx,url,NULL,NULL);
	if (err < 0) {
		return -1;
	}
	err = avformat_find_stream_info(ctx,NULL);
	if (err < 0) {
		return -1;
	}

	AVCodec * codec = NULL;
	int strm = av_find_best_stream(ctx, AVMEDIA_TYPE_AUDIO, -1, -1, &codec, 0);
	if (strm < 0) {
		return -1;
	}
	AVCodecContext * codecCtx = ctx->streams[strm]->codec;
	err = avcodec_open2(codecCtx, codec, NULL);
	if (err < 0) {
		return -1;
	}
	if (strcmp(codec->name , "mp3")==0) {
		//Either image data or we are some kind of multimedia codec with mp3 audio
		if (ctx->nb_streams > 1) {
			int strm = av_find_best_stream(ctx, AVMEDIA_TYPE_VIDEO, -1, -1, &codec, 0);
			if (strm < 0) {
				return -1;
			}
			AVCodecContext * codecCtx = ctx->streams[strm]->codec;
			err = avcodec_open2(codecCtx, codec, NULL);
			if (err < 0) {
				return -1;
			}

			//Lets assume this is our picture!
			if (strcmp(codec->name , "mjpeg")==0 || strcmp(codec->name , "png")==0) {
				return 0;
			}
			return -1;
		}
		return 0;
	}
	return -1;
}

AVPacket retrieve_album_art(const char *url) {
	AVPacket err;
	err.size = 0;

	if (!url) {
		return err;
	}

	AVFormatContext *pFormatCtx = avformat_alloc_context();

	if (avformat_open_input(&pFormatCtx, url, NULL, NULL) != 0) {
		return err;
	}

	// read the format headers
	if (pFormatCtx->iformat->read_header(pFormatCtx) < 0) {
		return err;
	}

	// find the first attached picture, if available
	for (int i = 0; i < pFormatCtx->nb_streams; i++) {
		if (pFormatCtx->streams[i]->disposition & AV_DISPOSITION_ATTACHED_PIC) {
			return pFormatCtx->streams[i]->attached_pic;
		}
	}
	return err;
}
*/
import "C"
import (
	"errors"
	"time"
	"unsafe"
)

func init() {
	C.av_register_all()
	C.avcodec_register_all()
}

//TODO(sjon): Use headers instead like in extract
//Returns whether or not the file is MP3 based on the streams that reside in it
func IsMP3(filename string) bool {
	cs := C.CString(filename)
	defer C.free(unsafe.Pointer(cs))
	if r := C.mp3_check(cs); r >= 0 {
		return true
	}
	return false
}

//Returns length of the audio
func MP3Duration(filename string) time.Duration {
	cs := C.CString(filename)
	defer C.free(unsafe.Pointer(cs))
	return time.Duration(C.mp3_duration(cs)) * 1000
}

//Extracts the first image we find in the MP3
func ExtractImage(filename string) ([]byte, error) {
	cs := C.CString(filename)
	defer C.free(unsafe.Pointer(cs))

	pkt := C.retrieve_album_art(cs)
	if pkt.size <= 0 {
		return nil, errors.New("Failed to extract image")
	}
	return C.GoBytes(unsafe.Pointer(pkt.data), pkt.size), nil
}
