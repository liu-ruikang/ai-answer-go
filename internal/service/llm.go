package service

import (
	"context"
	"fmt"

	v1 "ai-answer-go/api/llm/v1"
	"ai-answer-go/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// LLMService 实现LLM服务
type LLMService struct {
	v1.UnimplementedLLMServer

	uc  *biz.LLMUsecase
	log *log.Helper
}

// NewLLMService 创建LLM服务
func NewLLMService(uc *biz.LLMUsecase, logger log.Logger) *LLMService {
	return &LLMService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// ChatDeepseekR1 实现LLM服务的ChatDeepseekR1方法
func (s *LLMService) ChatDeepseekR1(ctx context.Context, req *v1.ChatDeepseekR1Request) (*v1.ChatDeepseekR1Response, error) {
	s.log.WithContext(ctx).Infof("ChatDeepseekR1 request: %v", req.SessionId)
	fmt.Println("ChatDeepseekR1 request:", req.SessionId)
	return s.uc.ChatDeepseekR1(ctx, req)
}

// StreamChatDeepseekR1 实现LLM服务的StreamChatDeepseekR1方法
func (s *LLMService) StreamChatDeepseekR1(req *v1.ChatDeepseekR1Request, stream v1.LLM_StreamChatDeepseekR1Server) error {
	s.log.Infof("StreamChatDeepseekR1 request: %v", req.SessionId)

	fmt.Println("StreamChatDeepseekR1 request:", req.SessionId)
	// 创建回调函数，用于处理流式响应
	callback := func(resp *v1.ChatDeepseekR1Response) error {
		return stream.Send(resp)
	}

	// 调用业务逻辑层
	return s.uc.StreamChatDeepseekR1(stream.Context(), req, callback)
}
