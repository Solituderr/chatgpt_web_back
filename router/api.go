package router

import (
	"chatgpt/app/controller"
	"github.com/gin-gonic/gin"
)

func Api(app *gin.Engine) {

	api := app.Group("/api")
	//api.Get("/get-balance", controller.GetBalance())                // 查询余额
	api.POST("/chat-process", controller.CreateChatCompletion())    // 发送聊天
	api.OPTIONS("/chat-process", controller.CreateChatCompletion()) // 发送聊天
	api.POST("/session", controller.CreateSession())                // 创建会话
	api.OPTIONS("/session", controller.CreateSession())
	api.POST("/verify", controller.GetVerify())
	api.OPTIONS("/verify", controller.GetVerify())
}
