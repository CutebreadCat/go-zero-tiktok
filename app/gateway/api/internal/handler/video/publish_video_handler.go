// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"
	"go_zero-tiktok/app/gateway/api/internal/logic/video"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/pkg/xerr"
)

func PublishVideoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(256 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("视频上传失败"))
			return
		}

		// 通道 1: multipart/form-data 真文件上传(优先)
		var (
			filename   string
			videoBytes []byte
		)
		if file, header, err := r.FormFile("file"); err == nil {
			videoBytes, err = io.ReadAll(file)
			file.Close()
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("读取视频文件失败"))
				return
			}
			filename = header.Filename
		}

		var req types.PublishVideoRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 通道 2: file 字段传 base64 文本(兜底)
		if len(videoBytes) == 0 && req.File != "" {
			b, err := base64.StdEncoding.DecodeString(req.File)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("文件内容格式错误"))
				return
			}
			videoBytes = b
			filename = "upload_" + time.Now().Format("20060102150405") + ".mp4"
		}

		if len(videoBytes) == 0 {
			httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("未找到视频文件"))
			return
		}

		ctx := video.WithPublishVideoFile(r.Context(), filename, videoBytes)
		l := video.NewPublishVideoLogic(ctx, svcCtx)
		resp, err := l.PublishVideo(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
