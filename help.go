package mcommon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	"github.com/twinj/uuid"
	"gopkg.in/go-playground/validator.v8"
)

// GetUUIDStr 获取唯一字符串
func GetUUIDStr() string {
	u1 := uuid.NewV4()
	return strings.Replace(u1.String(), "-", "", -1)
}

// IsStringInSlice 字符串是否在数组中
func IsStringInSlice(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

// IsIntInSlice 数字是否在数组中
func IsIntInSlice(arr []int64, str int64) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

// GinFillBindError 检测gin输入绑定错误
func GinFillBindError(c *gin.Context, err error) {
	validatorError, ok := err.(validator.ValidationErrors)
	if ok {
		errMsgList := make([]string, 0, 16)
		for _, v := range validatorError {
			errMsgList = append(errMsgList, fmt.Sprintf("[%s] is %s", strcase.ToSnake(v.Field), v.ActualTag))
		}
		c.JSON(
			http.StatusOK,
			gin.H{
				"error":   ErrorBind,
				"err_msg": strings.Join(errMsgList, ", "),
			},
		)
		return
	}
	unmarshalError, ok := err.(*json.UnmarshalTypeError)
	if ok {
		c.JSON(
			http.StatusOK,
			gin.H{
				"error":   ErrorBind,
				"err_msg": fmt.Sprintf("[%s] type error", unmarshalError.Field),
			},
		)
		return
	}
	if err == io.EOF {
		c.JSON(
			http.StatusOK,
			gin.H{
				"error":   ErrorBind,
				"err_msg": fmt.Sprintf("empty body"),
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		gin.H{
			"error":   ErrorInternal,
			"err_msg": ErrorInternalMsg,
		},
	)
}
