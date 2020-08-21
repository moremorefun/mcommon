package mcommon

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/speps/go-hashids"
)

type H map[string]interface{}

type GinResp struct {
	ErrCode int64  `json:"error"`
	ErrMsg  string `json:"error_msg"`
	Data    gin.H  `json:"data,omitempty"`
}

// GinRespSuccess 成功返回
var GinRespSuccess = GinResp{
	ErrCode: ErrorSuccess,
	ErrMsg:  ErrorSuccessMsg,
}

// GinRespInternalErr 成功返回
var GinRespInternalErr = GinResp{
	ErrCode: ErrorInternal,
	ErrMsg:  ErrorInternalMsg,
}

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

// EncodeHashID 获取hash id
func EncodeHashID(salt string, minLen, id int) (string, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = minLen
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}
	e, err := h.Encode([]int{id})
	if err != nil {
		return "", err
	}
	return e, nil
}

// DecodeHashID 解析hash id
func DecodeHashID(salt string, minLen int, value string) (int, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = minLen
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return 0, err
	}
	e, err := h.DecodeWithError(value)
	if err != nil {
		return 0, err
	}
	if len(e) > 0 {
		return 0, fmt.Errorf("error value: %s", value)
	}
	return e[0], nil
}

// EncodeHashIDs 获取hash id
func EncodeHashIDs(salt string, minLen int, ids []int) (string, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = minLen
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}
	e, err := h.Encode(ids)
	if err != nil {
		return "", err
	}
	return e, nil
}

// DecodeHashIDs 解析hash id
func DecodeHashIDs(salt string, minLen int, value string) ([]int, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = minLen
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, err
	}
	e, err := h.DecodeWithError(value)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// IntArrToString 数组转换
func IntArrToString(A []int64, denim string) string {
	var buffer bytes.Buffer
	for i := 0; i < len(A); i++ {
		buffer.WriteString(strconv.FormatInt(A[i], 10))
		if i != len(A)-1 {
			buffer.WriteString(denim)
		}
	}
	return buffer.String()
}

// Ip2long 转换ip
func Ip2long(ipAddr string) (uint32, error) {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return 0, errors.New("wrong ipAddr format")
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip), nil
}

// GinRepeatReadBody 创建可重复度body
func GinRepeatReadBody(c *gin.Context) error {
	var err error
	var body []byte
	if cb, ok := c.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body == nil {
		body, err = ioutil.ReadAll(c.Request.Body)
		if err != nil {
			Log.Errorf("err: [%T] %s", err, err.Error())
			c.Abort()
			return err
		}
		c.Set(gin.BodyBytesKey, body)
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return nil
}

// GinShouldBindRepeat 可重复绑定参数
func GinShouldBindRepeat(c *gin.Context, obj interface{}) error {
	err := GinRepeatReadBody(c)
	if err != nil {
		return err
	}
	return c.ShouldBind(obj)
}

// GinFillBindError 检测gin输入绑定错误
func GinFillBindError(c *gin.Context, err error) {
	GinDoRespErr(
		c,
		ErrorBind,
		fmt.Sprintf("[%T] %s", err, err.Error()),
		nil,
	)
}

// GinFillSuccessData 填充返回数据
func GinFillSuccessData(data gin.H) GinResp {
	return GinResp{
		ErrCode: ErrorSuccess,
		ErrMsg:  ErrorSuccessMsg,
		Data:    data,
	}
}

// GinDoRespSuccess 返回成功信息
func GinDoRespSuccess(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, GinResp{
		ErrCode: ErrorSuccess,
		ErrMsg:  ErrorSuccessMsg,
		Data:    data,
	})
}

// GinDoRespInternalErr 返回错误信息
func GinDoRespInternalErr(c *gin.Context) {
	c.JSON(http.StatusOK, GinResp{
		ErrCode: ErrorInternal,
		ErrMsg:  ErrorInternalMsg,
	})
}

// GinDoRespErr 返回特殊错误
func GinDoRespErr(c *gin.Context, code int64, msg string, data gin.H) {
	c.JSON(http.StatusOK, GinResp{
		ErrCode: code,
		ErrMsg:  msg,
		Data:    data,
	})
}

// GinDoEncRespSuccess 返回成功信息
func GinDoEncRespSuccess(c *gin.Context, key string, isAll bool, data gin.H) {
	var err error
	resp := GinResp{
		ErrCode: ErrorSuccess,
		ErrMsg:  ErrorSuccessMsg,
		Data:    data,
	}
	respBs := []byte("")
	if data != nil {
		respBs, err = json.Marshal(data)
		if err != nil {
			GinDoRespInternalErr(c)
			return
		}
	}
	encResp, err := AesEncrypt(string(respBs), key)
	if err != nil {
		GinDoRespInternalErr(c)
		return
	}
	if isAll {
		resp.Data["enc"] = encResp
	} else {
		resp.Data = gin.H{
			"enc": encResp,
		}
	}
	c.JSON(http.StatusOK, resp)
}

// GinDoEncRespInternalErr 返回错误信息
func GinDoEncRespInternalErr(c *gin.Context, key string, isAll bool) {
	resp := GinResp{
		ErrCode: ErrorInternal,
		ErrMsg:  ErrorInternalMsg,
	}
	c.JSON(http.StatusOK, resp)
}

// GinDoEncRespErr 返回特殊错误
func GinDoEncRespErr(c *gin.Context, key string, isAll bool, code int64, msg string, data gin.H) {
	resp := GinResp{
		ErrCode: code,
		ErrMsg:  msg,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}
