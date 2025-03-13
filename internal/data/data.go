package data

import (
	"ai-answer-go/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewGreeterRepo,
	NewDeepseekR1Client,
	NewLLMRepo,
)

// Data .
type Data struct {
	// TODO wrapped database client
	deepseekR1Client *DeepseekR1Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, deepseekR1Client *DeepseekR1Client) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		deepseekR1Client: deepseekR1Client,
	}, cleanup, nil
}

// NewDeepseekR1Client 创建Deepseek R1客户端
func NewDeepseekR1Client(c *conf.Data, logger log.Logger) *DeepseekR1Client {
	l := log.NewHelper(log.With(logger, "module", "data/deepseek-r1"))

	// 如果配置中没有Deepseek R1的配置，返回nil
	if c.Llm == nil || c.Llm.DeepseekR1 == nil {
		l.Warn("Deepseek R1 configuration not found, client will not be initialized")
		return nil
	}

	return NewDeepseekR1ClientWithConfig(
		c.Llm.DeepseekR1.ApiKey,
		c.Llm.DeepseekR1.BaseUrl,
		logger,
	)
}
