// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package interaction

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCommentLogic) DeleteComment(req *types.DeleteCommentRequest) (resp *types.DeleteCommentResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
