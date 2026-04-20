// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package interaction

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

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
		return nil, xerr.New(400, "视频ID不能为空")
	}
	if req.PageNumber <= 0 {
		req.PageNumber = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		return nil, xerr.New(400, "每页数量不能超过100")
	}

	comments, total, err := l.svcCtx.Dal.Comment.GetCommentsByVideoID(l.ctx, req.VideoID, req.PageNumber, req.PageSize)
	if err != nil {
		return nil, xerr.New(1002, "获取评论列表失败，请稍后重试")
	}

	resp = &types.GetCommentListResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "查询成功"},
		CommentList:  comments,
		CommentCount: int32(total),
	}
	if resp.CommentList == nil {
		resp.CommentList = []types.CommentBaseinfo{}
	}

	return resp, nil
}
