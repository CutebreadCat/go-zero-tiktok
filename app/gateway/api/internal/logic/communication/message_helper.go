package communication

import (
	"context"
	"fmt"
	"time"

	"go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	appLogger "go_zero-tiktok/pkg/logger"
)

// CreateMessageForInteraction 在互动成功后创建消息通知。
// 先查询视频作者作为接收人，失败仅记日志，不影响主链路响应。
func CreateMessageForInteraction(ctx context.Context, svcCtx *svc.ServiceContext, senderID int64, videoID int64, msgType, title, content string) error {
	if senderID == 0 || videoID == 0 {
		return nil
	}

	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	videoResp, err := svcCtx.VideoRpc.GetVideosByIDs(callCtx, &video_pb.GetVideosByIDsRequest{
		VideoIds: []int64{videoID},
	})
	if err != nil {
		appLogger.Errorf("CreateMessageForInteraction query video failed video_id=%d: %v", videoID, err)
		return err
	}
	if len(videoResp.Videos) == 0 {
		appLogger.Warnf("CreateMessageForInteraction video not found video_id=%d", videoID)
		return nil
	}
	receiverID := videoResp.Videos[0].AuthorId
	if receiverID == 0 || receiverID == senderID {
		return nil
	}

	eventID := fmt.Sprintf("%s:%d:%d:%d", msgType, senderID, videoID, time.Now().UnixMilli())
	_, err = svcCtx.CommunicationRpc.CreateMessage(callCtx, &communication_pb.CreateMessageRequest{
		ReceiverId: receiverID,
		Type:       msgType,
		Title:      title,
		Content:    content,
		EventId:    eventID,
		SenderId:   senderID,
		TargetId:   videoID,
		TargetType: "video",
	})
	if err != nil {
		appLogger.Errorf("CreateMessageForInteraction failed receiver=%d video_id=%d: %v", receiverID, videoID, err)
		return err
	}
	return nil
}
