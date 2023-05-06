package main

import (
	"chatgpt/model"
	"chatgpt/router"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	app := gin.Default()
	// 加载请求记录
	//app.Use(
	//	logger.New(
	//		logger.Config{
	//			Format:     "${time} ${status} - ${ip}:${port} - ${method} ${path} \n",
	//			TimeFormat: "2006-01-02 15:04:05.000",
	//			TimeZone:   "Asia/Shanghai",
	//		},
	//	),
	//)
	//Test()
	// 加载路由
	model.Init()
	router.Api(app)
	godotenv.Load(".env")
	app.Run(":3000")
}
