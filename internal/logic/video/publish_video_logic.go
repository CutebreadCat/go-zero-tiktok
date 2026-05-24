package video

import (
	"bytes"
	"context"

	"go_zero-tiktok/internal/infra/storage/aliyun"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

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
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	VideoID := myutils.GenerateVideoID()
	filename, ok := l.ctx.Value("filename").(string)
	if !ok || filename == "" {
		return nil, xerr.NewInvalidParam("视频文件名缺失")
	}
	videoBytes, ok := l.ctx.Value("video_bytes").([]byte)
	if !ok || len(videoBytes) == 0 {
		return nil, xerr.NewInvalidParam("视频文件内容为空")
	}
	objectKey := authorID + "/" + VideoID + "/" + filename
	var Videourl string
	if Videourl, err = aliyun.UploadBytesToOSS(bytes.NewReader(videoBytes), objectKey); err != nil {
		return nil, xerr.HandleDaoError(err, "PublishVideo.UploadToOSS")
	}

	if err := l.svcCtx.Dal.Video.CreateVideoFromParams(l.ctx, VideoID, authorID, Videourl, "", req.Title, req.Description); err != nil {
		return nil, xerr.HandleDaoError(err, "PublishVideo.CreateVideo")
	}
	resp = &types.PublishVideoResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		VideoID: VideoID,
	}

	return
}
