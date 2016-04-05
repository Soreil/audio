# audio
Audio package to go along with the image package in helping the thumbnailers needs.
This package uses CGo.

Only the first few kilobytes of the file is needed normally, 512K should work for just about any file.
MP4 needs the full file to create the context.

In case of only passing a part of the file you should not expect image extraction to work or bitrate and duration to be accurate if those are estimates.

Tested formats:
MP3
Opus
Ogg/Vorbis
Ogg/Opus
AAC
M4A/AAC
MP4/AAC
Webm/Vorbis
Webm/Opus
