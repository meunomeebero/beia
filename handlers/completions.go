package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

type CompletionsHandler struct {
	openaiClient *openai.Client
}

type CompletionRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

type CompletionResponse struct {
	Content string `json:"content"`
}

func NewCompletionsHandler(apiKey string) *CompletionsHandler {
	return &CompletionsHandler{
		openaiClient: openai.NewClient(apiKey),
	}
}

func (h *CompletionsHandler) HandleCompletion(c *gin.Context) {
	fmt.Print("starting completion...")

	var req CompletionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body. 'prompt' field is required",
		})
		return
	}

	resp, err := h.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req.Prompt,
				},
			},
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate completion",
		})
		return
	}

	if len(resp.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No completion generated",
		})
		return
	}

	c.JSON(http.StatusOK, CompletionResponse{
		Content: resp.Choices[0].Message.Content,
	})
}
