package logic

import (
	"bytes"
	"context"
	"strings"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/storage/aliyun"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type PublishVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewPublishVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishVideoLogic {
	return &PublishVideoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *PublishVideoLogic) PublishVideo(in *video_pb.PublishVideoRequest) (*video_pb.PublishVideoResponse, error) {
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.Title == "" {
		return nil, xerr.NewInvalidParam("视频标题不能为空")
	}
	if len(in.VideoData) == 0 {
		return nil, xerr.NewInvalidParam("视频文件内容为空")
	}

	idempotencyKey := strings.TrimSpace(in.IdempotencyKey)
	if len(idempotencyKey) > 64 {
		return nil, xerr.NewInvalidParam("幂等键长度不能超过64")
	}

	// 幂等命中：直接返回首次发布结果，不重复上传 OSS、不重复写库
	if idempotencyKey != "" {
		videoID, publishAt, found, err := l.svcCtx.VideoService.FindPublishedVideoByKey(l.ctx, in.UserId, idempotencyKey)
		if err != nil {
			return nil, xerr.HandleDaoError(err, "PublishVideo.CheckIdempotency")
		}
		if found {
			return &video_pb.PublishVideoResponse{
				VideoId:     videoID,
				PublishedAt: publishAt.UnixMilli(),
			}, nil
		}
	}

	videoID := myutils.GenerateVideoID()
	videoObjectKey := aliyun.BuildObjectKey(aliyun.ObjectTypeVideo, in.UserId, videoID, in.Filename)

	if _, err := l.svcCtx.Storage.UploadFile(bytes.NewReader(in.VideoData), videoObjectKey); err != nil {
		return nil, xerr.HandleDaoError(err, "PublishVideo.UploadToOSS")
	}

	actualID, publishAt, err := l.svcCtx.VideoService.PublishVideo(l.ctx, videoID, in.UserId, videoObjectKey, "", in.Title, in.Description, idempotencyKey)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "PublishVideo.CreateVideo")
	}

	// 并发重复请求：本次生成的 ID 未被采用，刚上传的对象无人引用，清理避免 OSS 孤儿文件
	if actualID != videoID {
		if err := l.svcCtx.Storage.DeleteFile(videoObjectKey); err != nil {
			l.Errorf("delete duplicated upload object %s failed: %v", videoObjectKey, err)
		}
	}

	return &video_pb.PublishVideoResponse{
		VideoId:     actualID,
		PublishedAt: publishAt.UnixMilli(),
	}, nil
}
