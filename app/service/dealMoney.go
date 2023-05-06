package service

import (
	"chatgpt/app/response"
	"chatgpt/model"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func DealCheckMoney(c *gin.Context, token string) bool {
	// test 处理是否余额不足
	check, _ := model.CheckMoney(token)
	if check == "fail" {
		err1 := errors.New("余额不足，联系管理员充值")
		e := response.ErrorStream{
			Status:  "Fail",
			Data:    nil,
			Message: err1.Error(),
		}
		c.JSON(http.StatusOK, e)
		return true
	}
	return false
}
