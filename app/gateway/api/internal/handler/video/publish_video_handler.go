// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"io"
	"net/http"

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
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("未找到视频文件"))
			return
		}
		defer file.Close()

		videoBytes, err := io.ReadAll(file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("读取视频文件失败"))
			return
		}
		ctx := video.WithPublishVideoFile(r.Context(), header.Filename, videoBytes)

		var req types.PublishVideoRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := video.NewPublishVideoLogic(ctx, svcCtx)
		resp, err := l.PublishVideo(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}