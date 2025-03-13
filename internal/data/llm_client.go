package data

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/prometheus/client_golang/prometheus"

	v1 "ai-answer-go/api/llm/v1"
)

// LLMClient 定义LLM客户端接口
type LLMClient interface {
	// Chat 执行聊天请求
	Chat(ctx context.Context, req *v1.ChatDeepseekR1Request) (*v1.ChatDeepseekR1Response, error)
	// StreamChat 执行流式聊天请求
	StreamChat(ctx context.Context, req *v1.ChatDeepseekR1Request, callback func(*v1.ChatDeepseekR1Response) error) error
}

// DeepseekR1Client Deepseek R1模型客户端
type DeepseekR1Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *log.Helper
	// 指标收集
	requestCounter  *prometheus.CounterVec
	requestLatency  *prometheus.HistogramVec
	tokenUsageGauge *prometheus.GaugeVec
}

// DeepseekR1Message Deepseek API消息格式
type DeepseekR1Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepseekR1Request Deepseek API请求格式
type DeepseekR1Request struct {
	Model       string              `json:"model"`
	Messages    []DeepseekR1Message `json:"messages"`
	Temperature float32             `json:"temperature,omitempty"`
	TopP        float32             `json:"top_p,omitempty"`
	MaxTokens   int32               `json:"max_tokens,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
}

// DeepseekR1Response Deepseek API响应格式
type DeepseekR1Response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int               `json:"index"`
		Message      DeepseekR1Message `json:"message"`
		FinishReason string            `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int32 `json:"prompt_tokens"`
		CompletionTokens int32 `json:"completion_tokens"`
		TotalTokens      int32 `json:"total_tokens"`
	} `json:"usage"`
}

// DeepseekR1StreamResponse Deepseek API流式响应格式
type DeepseekR1StreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int               `json:"index"`
		Delta        DeepseekR1Message `json:"delta"`
		FinishReason string            `json:"finish_reason"`
	} `json:"choices"`
}

// NewDeepseekR1ClientWithConfig 创建新的Deepseek R1客户端
func NewDeepseekR1ClientWithConfig(apiKey, baseURL string, logger log.Logger) *DeepseekR1Client {
	// 创建HTTP客户端，设置超时
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	// 创建Prometheus指标
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "deepseek_r1_requests_total",
			Help: "Total number of requests to Deepseek R1 API",
		},
		[]string{"status", "endpoint"},
	)

	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "deepseek_r1_request_duration_seconds",
			Help:    "Duration of Deepseek R1 API requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	tokenUsageGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "deepseek_r1_token_usage",
			Help: "Token usage for Deepseek R1 API requests",
		},
		[]string{"type"},
	)

	// 注册Prometheus指标
	prometheus.MustRegister(requestCounter, requestLatency, tokenUsageGauge)

	return &DeepseekR1Client{
		apiKey:          apiKey,
		baseURL:         baseURL,
		httpClient:      httpClient,
		logger:          log.NewHelper(logger),
		requestCounter:  requestCounter,
		requestLatency:  requestLatency,
		tokenUsageGauge: tokenUsageGauge,
	}
}

// Chat 实现LLMClient接口的Chat方法
func (c *DeepseekR1Client) Chat(ctx context.Context, req *v1.ChatDeepseekR1Request) (*v1.ChatDeepseekR1Response, error) {
	startTime := time.Now()

	// 转换请求格式
	apiReq := &DeepseekR1Request{
		Model:       "deepseek-r1",
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stream:      false,
	}

	// 转换消息格式
	for _, msg := range req.Messages {
		apiReq.Messages = append(apiReq.Messages, DeepseekR1Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 序列化请求
	reqBody, err := json.Marshal(apiReq)
	if err != nil {
		c.logger.Errorf("failed to marshal request: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		c.logger.Errorf("failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.requestCounter.WithLabelValues("error", "chat").Inc()
		c.logger.Errorf("failed to send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 记录请求延迟
	c.requestLatency.WithLabelValues("chat").Observe(time.Since(startTime).Seconds())

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		c.requestCounter.WithLabelValues("error", "chat").Inc()
		body, _ := io.ReadAll(resp.Body)
		c.logger.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 解析响应
	var apiResp DeepseekR1Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		c.requestCounter.WithLabelValues("error", "chat").Inc()
		c.logger.Errorf("failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 记录成功请求
	c.requestCounter.WithLabelValues("success", "chat").Inc()

	// 记录token使用情况
	c.tokenUsageGauge.WithLabelValues("prompt").Set(float64(apiResp.Usage.PromptTokens))
	c.tokenUsageGauge.WithLabelValues("completion").Set(float64(apiResp.Usage.CompletionTokens))
	c.tokenUsageGauge.WithLabelValues("total").Set(float64(apiResp.Usage.TotalTokens))

	// 构建响应
	response := &v1.ChatDeepseekR1Response{
		SessionId: req.SessionId,
		Model:     apiResp.Model,
		TokenUsage: &v1.TokenUsage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		},
	}

	// 提取生成的内容
	if len(apiResp.Choices) > 0 {
		response.Content = apiResp.Choices[0].Message.Content
	}

	return response, nil
}

// StreamChat 实现LLMClient接口的StreamChat方法
func (c *DeepseekR1Client) StreamChat(ctx context.Context, req *v1.ChatDeepseekR1Request, callback func(*v1.ChatDeepseekR1Response) error) error {
	startTime := time.Now()

	// 转换请求格式
	apiReq := &DeepseekR1Request{
		Model:       "deepseek-r1",
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	}

	// 转换消息格式
	for _, msg := range req.Messages {
		apiReq.Messages = append(apiReq.Messages, DeepseekR1Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 序列化请求
	reqBody, err := json.Marshal(apiReq)
	if err != nil {
		c.logger.Errorf("failed to marshal request: %v", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		c.logger.Errorf("failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.requestCounter.WithLabelValues("error", "stream-chat").Inc()
		c.logger.Errorf("failed to send request: %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		c.requestCounter.WithLabelValues("error", "stream-chat").Inc()
		body, _ := io.ReadAll(resp.Body)
		c.logger.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
		return fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 记录请求延迟（仅记录到首次响应的时间）
	c.requestLatency.WithLabelValues("stream-chat").Observe(time.Since(startTime).Seconds())

	// 处理流式响应
	reader := bufio.NewReader(resp.Body)
	var fullContent string
	var model string

	for {
		// 读取一行数据
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			c.logger.Errorf("failed to read stream: %v", err)
			return fmt.Errorf("failed to read stream: %w", err)
		}

		// 跳过空行和SSE前缀
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		// 提取JSON数据
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		// 解析流式响应
		var streamResp DeepseekR1StreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			c.logger.Errorf("failed to unmarshal stream response: %v", err)
			continue
		}

		// 提取模型信息
		if model == "" {
			model = streamResp.Model
		}

		// 提取内容
		var deltaContent string
		if len(streamResp.Choices) > 0 {
			deltaContent = streamResp.Choices[0].Delta.Content
			fullContent += deltaContent

			// 检查是否结束
			if streamResp.Choices[0].FinishReason != "" {
				// 这里可以处理结束原因
			}
		}

		// 回调处理部分响应
		if err := callback(&v1.ChatDeepseekR1Response{
			SessionId: req.SessionId,
			Content:   deltaContent,
			Model:     model,
		}); err != nil {
			c.logger.Errorf("callback error: %v", err)
			return fmt.Errorf("callback error: %w", err)
		}
	}

	// 记录成功请求
	c.requestCounter.WithLabelValues("success", "stream-chat").Inc()

	// 注意：流式API通常不会返回token使用情况，这里需要单独调用API获取
	// 这里简化处理，实际应用中可能需要额外的API调用或估算

	return nil
}

// 确保DeepseekR1Client实现了LLMClient接口
var _ LLMClient = (*DeepseekR1Client)(nil)
