// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.8.4
// - protoc             v5.29.3
// source: llm/v1/llm.proto

package v1

import (
	context "context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationLLMChatDeepseekR1 = "/api.llm.v1.LLM/ChatDeepseekR1"

type LLMHTTPServer interface {
	// ChatDeepseekR1 调用Deepseek R1模型
	ChatDeepseekR1(context.Context, *ChatDeepseekR1Request) (*ChatDeepseekR1Response, error)
}

func RegisterLLMHTTPServer(s *http.Server, srv LLMHTTPServer) {
	r := s.Route("/")
	r.POST("/v1/llm/deepseek-r1/chat", _LLM_ChatDeepseekR10_HTTP_Handler(srv))
}

func _LLM_ChatDeepseekR10_HTTP_Handler(srv LLMHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ChatDeepseekR1Request
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationLLMChatDeepseekR1)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ChatDeepseekR1(ctx, req.(*ChatDeepseekR1Request))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*ChatDeepseekR1Response)
		return ctx.Result(200, reply)
	}
}

type LLMHTTPClient interface {
	ChatDeepseekR1(ctx context.Context, req *ChatDeepseekR1Request, opts ...http.CallOption) (rsp *ChatDeepseekR1Response, err error)
}

type LLMHTTPClientImpl struct {
	cc *http.Client
}

func NewLLMHTTPClient(client *http.Client) LLMHTTPClient {
	return &LLMHTTPClientImpl{client}
}

func (c *LLMHTTPClientImpl) ChatDeepseekR1(ctx context.Context, in *ChatDeepseekR1Request, opts ...http.CallOption) (*ChatDeepseekR1Response, error) {
	var out ChatDeepseekR1Response
	pattern := "/v1/llm/deepseek-r1/chat"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationLLMChatDeepseekR1))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
