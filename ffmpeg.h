#ifndef CGO_AUDIO_H
#define CGO_AUDIO_H

#include <libavformat/avformat.h>

AVCodecContext *get_codecContext(AVFormatContext *ctx);
int64_t get_duration(AVFormatContext *ctx);
AVPacket retrieve_album_art(AVFormatContext *ctx);
int has_image(AVFormatContext *ctx);
void destroy(AVFormatContext *ctx);

#endif
