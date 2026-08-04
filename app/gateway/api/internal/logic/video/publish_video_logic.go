package video

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	videopb "go_zero-tiktok/app/video/rpc/video_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type PublishVideoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type publishVideoContextKey string

const (
	publishVideoFilenameKey publishVideoContextKey = "filename"
	publishVideoBytesKey    publishVideoContextKey = "video_bytes"
)

func WithPublishVideoFile(ctx context.Context, filename string, videoBytes []byte) context.Context {
	ctx = context.WithValue(ctx, publishVideoFilenameKey, filename)
	return context.WithValue(ctx, publishVideoBytesKey, videoBytes)
}

func NewPublishVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishVideoLogic {
	return &PublishVideoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *PublishVideoLogic) PublishVideo(req *types.PublishVideoRequest) (resp *types.PublishVideoResponse, err error) {
	authorID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	filename, ok := l.ctx.Value(publishVideoFilenameKey).(string)
	if !ok || filename == "" {
		return nil, xerr.NewInvalidParam("视频文件名缺失")
	}
	videoBytes, ok := l.ctx.Value(publishVideoBytesKey).([]byte)
	if !ok || len(videoBytes) == 0 {
		return nil, xerr.NewInvalidParam("视频文件内容为空")
	}

	rpcResp, err := l.svcCtx.VideoRpc.PublishVideo(l.ctx, &videopb.PublishVideoRequest{
		UserId:      authorID,
		Title:       req.Title,
		Description: req.Description,
		VideoData:   videoBytes,
		Filename:    filename,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "PublishVideo.CreateVideo")
	}

	resp = &types.PublishVideoResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		VideoID: rpcResp.VideoId,
	}

	return
}
