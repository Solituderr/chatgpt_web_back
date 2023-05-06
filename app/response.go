package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 统一返回
func Response(c *gin.Context, code int, data interface{}, msg string) {
	c.JSON(code, gin.H{
		"code": code,
		"data": data,
		"msg":  msg,
	})
}

// Success 成功返回
func Success(c *gin.Context, data interface{}) {
	Response(c, http.StatusOK, data, "success")
}

// Error 失败返回
func Error(c *gin.Context, msg string) {
	Response(c, http.StatusOK, nil, msg)
}
