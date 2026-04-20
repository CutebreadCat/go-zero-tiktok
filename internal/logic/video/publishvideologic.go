// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"bytes"
	"context"

	"go_zero-tiktok/internal/mw/ali"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/utils"

	"log"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublishVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishVideoLogic {
	return &PublishVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PublishVideoLogic) PublishVideo(req *types.PublishVideoRequest) (resp *types.PublishVideoResponse, err error) {
	authorID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.New(401, "用户身份信息无效，请重新登录")
	}
	VideoID := myutils.GenerateVideoID()
	filename, ok := l.ctx.Value("filename").(string)
	if !ok || filename == "" {
		return nil, xerr.New(400, "视频文件名缺失")
	}
	videoBytes, ok := l.ctx.Value("video_bytes").([]byte)
	if !ok || len(videoBytes) == 0 {
		return nil, xerr.New(400, "视频文件内容为空")
	}
	objectKey := authorID + "/" + VideoID + "/" + filename
	var Videourl string
	if Videourl, err = ali.UploadBytesToOSS(bytes.NewReader(videoBytes), objectKey); err != nil {
		log.Printf("failed to upload video to OSS: %v", err)
		return nil, xerr.New(1004, "视频上传失败，请稍后重试")
	}

	video := &types.VideoBaseinfo{
		VideoID:     VideoID,
		AuthorID:    authorID,
		VideoURL:    Videourl,
		CoverURL:    "",
		Title:       req.Title,
		Description: req.Description,
	}

	if err := l.svcCtx.Dal.Video.CreateVideo(l.ctx, video); err != nil {
		log.Printf("创建视频记录失败: %v", err)
		return nil, xerr.New(1002, "发布视频失败，请稍后重试")
	}
	resp = &types.PublishVideoResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		VideoID: video.VideoID,
	}

	return
}
