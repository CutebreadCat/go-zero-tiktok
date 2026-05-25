package myffmpeg

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TranscodeVideoFunc 是 TranscodeVideo 的函数类型
type TranscodeVideoFunc func(ctx context.Context, filepath string) ([]byte, error)

// MockFfmpeg 是 FfmpegInterface 的 mock 实现
type MockFfmpeg struct {
	TranscodeVideoFunc TranscodeVideoFunc
}

func (m *MockFfmpeg) TranscodeVideo(ctx context.Context, filepath string) ([]byte, error) {
	if m.TranscodeVideoFunc != nil {
		return m.TranscodeVideoFunc(ctx, filepath)
	}
	return []byte("mock video data"), nil
}

func TestTranscodeVideoWithMock(t *testing.T) {
	type args struct {
		ctx      context.Context
		filepath string
	}
	type want struct {
		data    []byte
		wantErr bool
	}
	testCases := map[string]struct {
		args args
		want want
		mock TranscodeVideoFunc
	}{
		"transcode video successfully": {
			args: args{
				ctx:      context.Background(),
				filepath: "test.mp4",
			},
			want: want{
				data:    []byte("mock video data"),
				wantErr: false,
			},
			mock: func(ctx context.Context, filepath string) ([]byte, error) {
				return []byte("mock video data"), nil
			},
		},
		"transcode video failed": {
			args: args{
				ctx:      context.Background(),
				filepath: "nonexistent.mp4",
			},
			want: want{
				data:    nil,
				wantErr: true,
			},
			mock: func(ctx context.Context, filepath string) ([]byte, error) {
				return nil, errors.New("ffmpeg error")
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mock := &MockFfmpeg{
				TranscodeVideoFunc: tc.mock,
			}
			data, err := mock.TranscodeVideo(tc.args.ctx, tc.args.filepath)
			if tc.want.wantErr {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want.data, data)
			}
		})
	}
}

func TestNewFfmpeg(t *testing.T) {
	f := NewFfmpeg()
	assert.NotNil(t, f)
	// 验证 Ffmpeg 实现了 FfmpegInterface 接口
	var _ FfmpegInterface = f
}
