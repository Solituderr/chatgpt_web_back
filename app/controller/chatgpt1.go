package controller

import (
	"bufio"
	"chatgpt/app"
	"chatgpt/app/request"
	"chatgpt/app/response"
	"chatgpt/app/service"
	"chatgpt/database"
	"chatgpt/model"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gookit/validate"
	"github.com/sashabaranov/go-openai"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// GetBalance 查询账户余额
//func GetBalance() fiber.Handler {
//	return func(c *fiber.Ctx) error {
//		var req request.BalanceRequest
//		if err := c.QueryParser(&req); err != nil {
//			return app.Error(c, "解析参数错误："+err.Error())
//		}
//
//		v := validate.New(req)
//		if !v.Validate() {
//			return app.Error(c, v.Errors.One())
//		}
//
//		// 查询余额
//		result, err := service.GetBalance(req.Key)
//		if err != nil {
//			return app.Error(c, err.Error())
//		}
//
//		return app.Success(c, response.BalanceResponse{
//			Total:   result.TotalGranted,
//			Used:    result.TotalUsed,
//			Balance: result.TotalAvailable,
//		})
//	}
//}

// CreateChatCompletion 发送聊天
func CreateChatCompletion() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request.ChatCompletionRequest
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Transfer-Encoding", "chunked")
		c.Header("Keep-Alive", "timeout=4")
		c.Header("Proxy-Connection", "keep-alive")
		c.Header("connection", "keep-alive")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "authorization, Content-Type")
		if err := c.ShouldBind(&req); err != nil {
			app.Error(c, "解析参数错误："+err.Error())
			return
		}

		v := validate.New(req)
		if !v.Validate() {
			app.Error(c, v.Errors.One())
			return
		}

		chatGPT := service.ChatGPTService{
			Key:   "",
			Proxy: os.Getenv("PROXY"),
			ChatCompletionRequest: openai.ChatCompletionRequest{
				Model:       req.Model,
				Temperature: req.Temperature,
				Stream:      true,
			},
		}

		// 处理是否有余额
		if service.DealCheckMoney(c, req.Token) {
			return
		}

		if req.DisableStream {
			chatGPT.ChatCompletionRequest.Stream = false
		}
		// 处理上下文
		if err := chatGPT.ContextHandler(req); err != nil {
			e := response.ErrorStream{
				Status:  "Fail",
				Data:    nil,
				Message: ErrorHandler(req.Key, err).Error(),
			}
			c.JSON(http.StatusOK, e)
			return
		}

		// 创建聊天
		if err := chatGPT.CreateChatCompletion(); err != nil {
			e := response.ErrorStream{
				Status:  "Fail",
				Data:    nil,
				Message: ErrorHandler(req.Key, err).Error(),
			}
			c.JSON(http.StatusOK, e)
			return
		}

		if req.DisableStream {
			var text string
			for _, message := range chatGPT.Response.ChatCompletionResponse.Choices {
				text += message.Message.Content
			}
			database.GPTCache.Add(chatGPT.Response.ChatCompletionResponse.ID, 1*time.Hour, database.GPTCacheItem{
				NowID:    chatGPT.Response.ChatCompletionResponse.ID,
				Prompt:   req.Prompt,
				ParentID: req.Options.ParentMessageId,
				Answer:   text,
			})
			app.Success(c, response.ChatCompletionResponse{
				Role:   openai.ChatMessageRoleAssistant,
				Id:     chatGPT.Response.ChatCompletionResponse.ID,
				Text:   text,
				Detail: chatGPT.Response.ChatCompletionResponse,
			})
			return
		}

		var gptResult response.ChatCompletionStreamResponse
		// 禁用缓冲
		//c.Writer.WriteHeader(http.StatusOK)
		//c.Writer.(http.Flusher).Flush()
		// fiber 返回stream
		for {
			r, err := chatGPT.Response.Stream.Recv()
			if errors.Is(err, io.EOF) {

				if gptResult.Id != "" && !database.GPTCache.Exists(gptResult.Id) {
					database.GPTCache.Add(gptResult.Id, 1*time.Hour, database.GPTCacheItem{
						NowID:    gptResult.Id,
						ParentID: req.Options.ParentMessageId,
						Prompt:   req.Prompt,
						Answer:   gptResult.Text,
					})
				}
				// 扣次数
				if err := model.DeleteChatNum(req.Token); err != nil {
					fmt.Println(err.Error())
					app.Error(c, "6")
				}
				return
			}
			if err != nil {
				fmt.Println(err)
				e := response.ErrorStream{
					Message: ErrorHandler(req.Key, err).Error(),
					Data:    nil,
					Status:  "Fail",
				}

				marshal, _ := json.Marshal(e)
				if _, err := fmt.Fprintf(c.Writer, "%s\n", marshal); err != nil {
					fmt.Println(err)
					return
				}
				return
			}

			stop := c.Stream(func(w io.Writer) bool {
				bw := bufio.NewWriter(w)
				if len(r.Choices) != 0 {
					gptResult.Detail = &r
					gptResult.Id = r.ID
					gptResult.Role = openai.ChatMessageRoleAssistant
					gptResult.Text += r.Choices[0].Delta.Content // 流传输
					marshal, _ := json.Marshal(gptResult)
					if _, err := fmt.Fprintf(bw, "%s\n", marshal); err != nil {
						fmt.Println(err)
						return true
					}
					bw.Flush()
				}
				return false
			}) //stop
			if stop {
				fmt.Println("666")
				break
			}
		} //for
		return
	}
}

func ErrorHandler(key string, err error) error {
	log.Println(err)
	// 无效key
	if strings.Contains(err.Error(), "Incorrect API key provided") {
		return errors.New("无效key，请检查key是否正确")

	}
	// 余额不足
	if strings.Contains(err.Error(), "You exceeded your current quota") {
		if key == "" {
			// 更换key
			if errs := service.ChangeKey(); errs != nil {
				return errors.New("更换key失败，请联系管理员")
			}
			return errors.New("请重试")

		}
		return errors.New("余额不足，请充值, 或更换key")
	}
	if strings.Contains(err.Error(), "You didn't provide an API key") {
		return errors.New("未提供key，请提供key")
	}
	if strings.Contains(err.Error(), "Rate limit reached for") {
		return errors.New("当前请求次数过多，请稍后再试即可")
	}
	// 未知错误
	if err != nil {
		return errors.New("未知错误，请联系管理员")
	}
	return err
}

// session
func CreateSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "authorization, Content-Type")

		c.JSON(http.StatusOK, response.SessionResponse{
			Status:  "Success",
			Message: "",
			Data: struct {
				Auth  bool   `json:"auth"`
				Model string `json:"model"`
			}(struct {
				Auth  bool
				Model string
			}{Auth: true, Model: "ChatGPTAPI"}),
		})
		return
	}
}

// verify

func GetVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "authorization, Content-Type")
		var req response.VerifyResponse
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusOK, response.ErrorStream{
				Status:  "Fail",
				Message: err.Error(),
				Data:    nil,
			})
		}
		token := req.Token
		fmt.Println(token)
		// 数据库比较
		if token != "hdu666" {
			c.JSON(http.StatusOK, response.ErrorStream{
				Status:  "Fail",
				Message: "密钥错误",
				Data:    nil,
			})
		} else {
			c.JSON(http.StatusOK, response.ErrorStream{
				Status:  "Success",
				Message: "Verify successfully",
				Data:    nil,
			})
		}
	}
}
