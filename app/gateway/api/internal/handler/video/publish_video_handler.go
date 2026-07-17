// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"io"
	"log"
	"net/http"

	"go_zero-tiktok/app/gateway/api/internal/logic/video"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	myutils "go_zero-tiktok/pkg/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PublishVideoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublishVideoRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		reader, fileheader, err := r.FormFile("video_file")
		if err != nil {
			log.Printf("failed to get video file from form: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if err := myutils.CheckVideo(reader); err != nil {
			log.Printf("failed to check video file: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		videoBytes, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("failed to read video file bytes: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		ctx := video.WithPublishVideoFile(r.Context(), fileheader.Filename, videoBytes)

		l := video.NewPublishVideoLogic(ctx, svcCtx)
		resp, err := l.PublishVideo(&req)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
		} else {
			httpx.OkJsonCtx(ctx, w, resp)
		}
	}
}
