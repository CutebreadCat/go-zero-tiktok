package myutils

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"

	"github.com/h2non/filetype"
)

func CheckImage(file multipart.File) error {

	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件流失败: %v", err)
	}

	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	kind, _ := filetype.Match(buf)
	if kind == filetype.Unknown {
		return fmt.Errorf("文件类型错误：不是有效的图片文件")
	}

	if kind.Extension != "jpg" && kind.Extension != "jpeg" && kind.Extension != "png" && kind.Extension != "gif" {
		return fmt.Errorf("文件类型错误：仅支持jpg、jpeg、png和gif格式的图片")
	}
	return nil
}

func CheckVideo(file multipart.File) error {

	buf := make([]byte, 1024)
	_, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件流失败: %v", err)
	}

	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	kind, _ := filetype.Match(buf)

	if kind == filetype.Unknown {
		return fmt.Errorf("文件类型错误：不是有效的视频文件")
	}
	return nil
}
