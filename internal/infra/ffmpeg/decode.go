package myffmpeg

import (
	"bytes"
	"context"
	"go_zero-tiktok/internal/shared/xerr"
	"log"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type FfmpegInterface interface {
	TranscodeVideo(ctx context.Context, filepath string) ([]byte, error)
}

type Ffmpeg struct{}

func NewFfmpeg() *Ffmpeg {
	return &Ffmpeg{}
}

func (f *Ffmpeg) TranscodeVideo(ctx context.Context, filepath string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(filepath).
		Output("pipe:", ffmpeg.KwArgs{"format": "mp4", "vcodec": "libx264", "preset": "fast"}).
		WithOutput(buf).
		Run()
	if err != nil {
		log.Println("ffmpeg transcode video failed")
		return nil, xerr.NewServerBusy()
	}
	return buf.Bytes(), nil
}
