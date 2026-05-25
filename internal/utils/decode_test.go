package myutils

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type testMultipartFile struct {
	*bytes.Reader
	err error
}

func newTestMultipartFile(data []byte) *testMultipartFile {
	return &testMultipartFile{Reader: bytes.NewReader(data)}
}

func (f *testMultipartFile) Close() error {
	return nil
}

func (f *testMultipartFile) Read(p []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.Reader.Read(p)
}

func TestCheckImage(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

	file := newTestMultipartFile(png)
	if err := CheckImage(file); err != nil {
		t.Fatalf("CheckImage png unexpected error: %v", err)
	}
	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		t.Fatalf("seek current: %v", err)
	}
	if pos != 0 {
		t.Fatalf("file position = %d, want reset to 0", pos)
	}

	if err := CheckImage(newTestMultipartFile([]byte("not-image"))); err == nil {
		t.Fatal("expected invalid image error")
	}

	readErr := errors.New("read failed")
	if err := CheckImage(&testMultipartFile{Reader: bytes.NewReader(nil), err: readErr}); err == nil {
		t.Fatal("expected read error")
	}
}

func TestCheckVideo(t *testing.T) {
	mp4 := append([]byte{0x00, 0x00, 0x00, 0x18}, []byte("ftypmp42")...)
	if err := CheckVideo(newTestMultipartFile(mp4)); err != nil {
		t.Fatalf("CheckVideo mp4 unexpected error: %v", err)
	}

	if err := CheckVideo(newTestMultipartFile([]byte("not-video"))); err == nil {
		t.Fatal("expected invalid video error")
	}

	readErr := errors.New("read failed")
	if err := CheckVideo(&testMultipartFile{Reader: bytes.NewReader(nil), err: readErr}); err == nil {
		t.Fatal("expected read error")
	}
}
