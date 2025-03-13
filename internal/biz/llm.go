package biz

import (
	"context"

	v1 "ai-answer-go/api/llm/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// LLMRepo 定义LLM仓库接口
type LLMRepo interface {
	// ChatDeepseekR1 调用Deepseek R1模型
	ChatDeepseekR1(ctx context.Context, req *v1.ChatDeepseekR1Request) (*v1.ChatDeepseekR1Response, error)
	// StreamChatDeepseekR1 流式调用Deepseek R1模型
	StreamChatDeepseekR1(ctx context.Context, req *v1.ChatDeepseekR1Request, callback func(*v1.ChatDeepseekR1Response) error) error
}

// LLMUsecase 定义LLM用例
type LLMUsecase struct {
	repo LLMRepo
	log  *log.Helper
}

// NewLLMUsecase 创建LLM用例
func NewLLMUsecase(repo LLMRepo, logger log.Logger) *LLMUsecase {
	return &LLMUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// ChatDeepseekR1 调用Deepseek R1模型
func (uc *LLMUsecase) ChatDeepseekR1(ctx context.Context, req *v1.ChatDeepseekR1Request) (*v1.ChatDeepseekR1Response, error) {
	uc.log.WithContext(ctx).Infof("ChatDeepseekR1: %v", req.SessionId)
	return uc.repo.ChatDeepseekR1(ctx, req)
}

// StreamChatDeepseekR1 流式调用Deepseek R1模型
func (uc *LLMUsecase) StreamChatDeepseekR1(ctx context.Context, req *v1.ChatDeepseekR1Request, callback func(*v1.ChatDeepseekR1Response) error) error {
	uc.log.WithContext(ctx).Infof("StreamChatDeepseekR1: %v", req.SessionId)
	return uc.repo.StreamChatDeepseekR1(ctx, req, callback)
}
