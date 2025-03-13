package data

import (
	"context"
	"fmt"

	v1 "ai-answer-go/api/llm/v1"
	"ai-answer-go/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type llmRepo struct {
	data *Data
	log  *log.Helper
}

// NewLLMRepo 创建LLM仓库
func NewLLMRepo(data *Data, logger log.Logger) biz.LLMRepo {
	return &llmRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// ChatDeepseekR1 实现LLMRepo接口的ChatDeepseekR1方法
func (r *llmRepo) ChatDeepseekR1(ctx context.Context, req *v1.ChatDeepseekR1Request) (*v1.ChatDeepseekR1Response, error) {
	if r.data.deepseekR1Client == nil {
		return nil, ErrDeepseekR1ClientNotInitialized
	}
	return r.data.deepseekR1Client.Chat(ctx, req)
}

// StreamChatDeepseekR1 实现LLMRepo接口的StreamChatDeepseekR1方法
func (r *llmRepo) StreamChatDeepseekR1(ctx context.Context, req *v1.ChatDeepseekR1Request, callback func(*v1.ChatDeepseekR1Response) error) error {
	if r.data.deepseekR1Client == nil {
		return ErrDeepseekR1ClientNotInitialized
	}
	return r.data.deepseekR1Client.StreamChat(ctx, req, callback)
}

// 错误定义
var ErrDeepseekR1ClientNotInitialized = fmt.Errorf("deepseek R1 client not initialized")
