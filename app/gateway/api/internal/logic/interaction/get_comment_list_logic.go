package interaction

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb/interaction_pb"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListLogic {
	return &GetCommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCommentListLogic) GetCommentList(req *types.GetCommentListRequest) (resp *types.GetCommentListResponse, err error) {
	if req.VideoID == "" {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}
	if req.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量不能超过100")
	}

	rpcResp, err := l.svcCtx.InteractionRpc.GetCommentList(l.ctx, &interactionpb.GetCommentListRequest{
		VideoId:  req.VideoID,
		PageNum:  req.PageNumber,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
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

	resp = &types.GetCommentListResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "查询成功"},
		CommentList:  comments,
		CommentCount: int32(rpcResp.Total),
	}
	if resp.CommentList == nil {
		resp.CommentList = []types.CommentBaseinfo{}
	}

	return resp, nil
}
