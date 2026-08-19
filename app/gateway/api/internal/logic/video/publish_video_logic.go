package video

import (
	"context"
	"time"

	communicationpb "go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	videopb "go_zero-tiktok/app/video/rpc/video_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
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

	// fanoutFansPageSize 扇出时拉取粉丝列表的每页大小
	fanoutFansPageSize = 1000
	// maxFanoutFans 扇出上限保护：大 V 粉丝超过该值时只扇给前 N 个粉丝，
	// 超出的部分依赖 feed:global 全站池兜底可见，后续"异步扇出"再放开。
	maxFanoutFans = 10000
)

// WithPublishVideoFile 将视频文件名与内容注入 context，供 PublishVideo 使用。
// 由 handler 在解析 multipart 后调用。
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

	// 关注流扇出（尽力而为）：拉取作者粉丝 → 调 video.rpc.FeedFanout 写 feed:inbox:{uid}。
	// 失败不阻断发布主流程，仅记日志；未关注任何人的作者无粉丝，直接跳过。
	if rpcResp.VideoId > 0 {
		l.fanoutToFollowers(rpcResp.VideoId, authorID, rpcResp.PublishedAt)
	}

	return &types.PublishVideoResponse{
		Base:    types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		VideoID: rpcResp.VideoId,
	}, nil
}

// fanoutToFollowers 关注流扇出：拉取作者全部粉丝列表 → 调 video.rpc.FeedFanout 写入每个粉丝的收件箱。
// 跨服务调用仅存在于 gateway 编排层（video.rpc / communication.rpc 之间禁止互调）。
func (l *PublishVideoLogic) fanoutToFollowers(videoID, authorID, publishAt int64) {
	defer func() {
		if r := recover(); r != nil {
			l.Errorf("fanout to followers panic, videoID=%d authorID=%d: %v", videoID, authorID, r)
		}
	}()

	// 独立超时上下文，扇出同步执行，最坏情况控制在 10s 内，不拖垮发布请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fans, err := l.fetchAllFans(ctx, authorID)
	if err != nil {
		l.Errorf("fetch fans of author %d failed, skip fanout: %v", authorID, err)
		return
	}
	if len(fans) == 0 {
		return
	}

	if _, err := l.svcCtx.VideoRpc.FeedFanout(ctx, &videopb.FeedFanoutRequest{
		VideoId:   videoID,
		UserIds:   fans,
		PublishAt: publishAt,
	}); err != nil {
		l.Errorf("feed fanout failed, videoID=%d fans=%d: %v", videoID, len(fans), err)
	}
}

// fetchAllFans 循环翻页拉取作者全部粉丝，带扇出上限保护。
func (l *PublishVideoLogic) fetchAllFans(ctx context.Context, authorID int64) ([]int64, error) {
	fans := make([]int64, 0, fanoutFansPageSize)
	page := int32(1)
	for {
		resp, err := l.svcCtx.CommunicationRpc.GetFansList(ctx, &communicationpb.GetFansListRequest{
			UserId:   authorID,
			PageNum:  page,
			PageSize: fanoutFansPageSize,
		})
		if err != nil {
			return nil, err
		}
		fans = append(fans, resp.UserIds...)

		// 扇出上限保护：只扇给前 maxFanoutFans 个粉丝，超出的粉丝靠 feed:global 兜底可见
		if len(fans) >= maxFanoutFans {
			l.Infof("fanout reach limit %d, truncated for author %d", maxFanoutFans, authorID)
			return fans[:maxFanoutFans], nil
		}
		// 本页为空或已拉完所有粉丝则结束
		if len(resp.UserIds) == 0 || int64(len(fans)) >= resp.Total {
			break
		}
		page++
	}
	return fans, nil
}
