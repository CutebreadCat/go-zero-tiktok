package interaction

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetCommentListLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListLogic {
	return &GetCommentListLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetCommentListLogic) GetCommentList(req *types.GetCommentListRequest) (resp *types.GetCommentListResponse, err error) {
	if req.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}
	if req.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量不能超过100")
	}

	rpcResp, err := l.svcCtx.InteractionRpc.GetCommentList(l.ctx, &interactionpb.GetCommentListRequest{
		VideoId:  req.VideoId,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetCommentList.GetCommentList")
	}

	comments := make([]types.CommentBaseinfo, 0, len(rpcResp.Comments))
	for _, c := range rpcResp.Comments {
		comments = append(comments, types.CommentBaseinfo{
			CommentID:       c.CommentId,
			UserID:          c.UserId,
			VideoID:         c.VideoId,
			Content:         c.Content,
			ParentCommentID: c.ParentCommentId,
			CreatedAt:       c.CreatedAt,
		})
	}

	return &types.GetCommentListResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "查询成功"},
		CommentList:  comments,
		CommentCount: rpcResp.Total,
	}, nil
}
