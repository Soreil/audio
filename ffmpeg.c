#include "ffmpeg.h"
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavformat/avio.h>
#include <libavutil/avutil.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

AVCodecContext *get_codecContext(AVFormatContext *ctx,
				 enum AVMediaType strmType)
{
	AVCodec *codec = NULL;

	int strm = av_find_best_stream(ctx, strmType, -1, -1, &codec, 0);
	if (strm < 0 || strm == AVERROR_STREAM_NOT_FOUND) {
		return NULL;
	}
	AVCodecContext *codecCtx = ctx->streams[strm]->codec;
	int err = avcodec_open2(codecCtx, codec, NULL);
	if (err < 0) {
		return NULL;
	}
	return codecCtx;
}

// Doesn't seem to produce any nice results sadly
int64_t get_duration(AVFormatContext *ctx)
{
	int strm =
	    av_find_best_stream(ctx, AVMEDIA_TYPE_AUDIO, -1, -1, NULL, 0);
	if (strm < 0 || strm == AVERROR_STREAM_NOT_FOUND) {
		return 0;
	}
	return ctx->streams[strm]->duration;
}

// Extract embedded images
AVPacket retrieve_album_art(AVFormatContext *ctx)
{
	AVPacket err;

	// find the first attached picture, if available
	for (int i = 0; i < ctx->nb_streams; i++) {
		if (ctx->streams[i]->disposition &
		    AV_DISPOSITION_ATTACHED_PIC) {
			return ctx->streams[i]->attached_pic;
		}
	}
	return err;
}

int has_image(AVFormatContext *ctx)
{
	// find the first attached picture, if available
	for (int i = 0; i < ctx->nb_streams; i++) {
		if (ctx->streams[i]->disposition &
		    AV_DISPOSITION_ATTACHED_PIC) {
			return 0;
		}
	}
	return 1;
}

void destroy(AVFormatContext *ctx)
{
	av_free(ctx->pb->buffer);
	ctx->pb->buffer = NULL;
	av_free(ctx->pb);
	avformat_close_input(&ctx);
}
