package communication

import (
	"context"

	"go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessagesLogic {
	return &GetMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMessagesLogic) GetMessages(req *types.GetMessagesRequest) (resp *types.GetMessagesResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	result, err := l.svcCtx.CommunicationRpc.GetMessages(l.ctx, &communication_pb.GetMessagesRequest{
		UserId: userID,
		Type:   req.Type,
		Cursor: req.Cursor,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetMessages")
	}

	items := make([]types.MessageInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, types.MessageInfo{
			MessageID:       item.MessageId,
			ReceiverID:      item.ReceiverId,
			Type:            item.Type,
			Title:           item.Title,
			Content:         item.Content,
			EventID:         item.EventId,
			SenderID:        item.SenderId,
			SenderNickname:  item.SenderNickname,
			SenderAvatarURL: item.SenderAvatarUrl,
			TargetID:        item.TargetId,
			TargetType:      item.TargetType,
			IsRead:          item.IsRead,
			CreatedAt:       item.CreatedAt,
			ReadAt:          item.ReadAt,
		})
	}

	return &types.GetMessagesResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		Items:      items,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	}, nil
}
