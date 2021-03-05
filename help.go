package mcommon

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/speps/go-hashids"
)

// H 通用map
type H map[string]interface{}

// GinResp 通用返回
type GinResp struct {
	ErrCode int64  `json:"error"`
	ErrMsg  string `json:"error_msg"`
	Data    gin.H  `json:"data,omitempty"`
}

// XMLNode xml结构
type XMLNode struct {
	XMLName xml.Name
	Content string    `xml:",chardata"`
	Nodes   []XMLNode `xml:",any"`
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

func GinBodyRepeat(r io.Reader) (io.ReadCloser, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &nopBodyRepeat{body: body}, nil
}

type nopBodyRepeat struct {
	body []byte
	i    int
}

func (o *nopBodyRepeat) Read(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return 0, nil
	}
	remain := len(o.body) - o.i
	if remain < n {
		n = copy(p, o.body[o.i:])
		o.i = 0
		return n, io.EOF
	}
	n = copy(p, o.body[o.i:])
	o.i += n
	return n, nil
}

func (*nopBodyRepeat) Close() error { return nil }

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

// IsInSlice 数字是否在数组中
func IsInSlice(arr []interface{}, iv interface{}) bool {
	for _, v := range arr {
		if v == iv {
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

// IP2long 转换ip
func IP2long(ipAddr string) (uint32, error) {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return 0, errors.New("wrong ipAddr format")
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip), nil
}

// GetHash 获取hash
func GetHash(in string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(in))
	if err != nil {
		return "", err
	}
	out := fmt.Sprintf("%x", h.Sum(nil))
	return out, nil
}

// WechatGetSign 获取签名
func WechatGetSign(appSecret string, paramsMap gin.H) string {
	var args []string
	var keys []string
	for k := range paramsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := fmt.Sprintf("%s=%v", k, paramsMap[k])
		args = append(args, v)
	}
	baseString := strings.Join(args, "&")
	baseString += fmt.Sprintf("&key=%s", appSecret)
	data := []byte(baseString)
	r := md5.Sum(data)
	signedString := hex.EncodeToString(r[:])
	return strings.ToUpper(signedString)
}

// WechatCheckSign 检查签名
func WechatCheckSign(appSecret string, paramsMap gin.H) bool {
	noSignMap := gin.H{}
	for k, v := range paramsMap {
		if k != "sign" {
			noSignMap[k] = v
		}
	}
	getSign := WechatGetSign(appSecret, noSignMap)
	if getSign != paramsMap["sign"] {
		return false
	}
	return true
}

func walk(nodes []XMLNode, f func(XMLNode) bool) {
	for _, n := range nodes {
		if f(n) {
			walk(n.Nodes, f)
		}
	}
}

// XMLWalk 遍历xml
func XMLWalk(bs []byte) (map[string]interface{}, error) {
	buf := bytes.NewBuffer(bs)
	dec := xml.NewDecoder(buf)
	r := make(map[string]interface{})
	var n XMLNode
	err := dec.Decode(&n)
	if err != nil {
		return nil, err
	}
	walk([]XMLNode{n}, func(n XMLNode) bool {
		content := strings.TrimSpace(n.Content)
		if content != "" {
			r[n.XMLName.Local] = n.Content
		}
		return true
	})
	return r, nil
}

// TimeGetMillisecond 获取毫秒
func TimeGetMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// GinFillBindError 检测gin输入绑定错误
func GinFillBindError(c *gin.Context, err error) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	Log.Warnf("bind error body is: %s", body)
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
	respBs := []byte("{}")
	if data != nil {
		respBs, err = jsoniter.Marshal(data)
		if err != nil {
			GinDoRespInternalErr(c)
			return
		}
	} else {
		resp.Data = gin.H{}
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

// GinMidRepeatReadBody 创建可重复度body
func GinMidRepeatReadBody(c *gin.Context) {
	var err error
	c.Request.Body, err = GinBodyRepeat(c.Request.Body)
	if err != nil {
		Log.Errorf("err: [%T] %s", err, err.Error())
		GinDoRespInternalErr(c)
		c.Abort()
		return
	}
}

// GinMinTokenToUserID token转换为user_id
func GinMinTokenToUserID(tx DbExeAble, getUserIDByToken func(ctx context.Context, tx DbExeAble, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		err := c.ShouldBind(&req)
		if err != nil {
			Log.Errorf("err: [%T] %s", err, err.Error())
			GinFillBindError(c, err)
			c.Abort()
			return
		}
		userID, err := getUserIDByToken(c, tx, req.Token)
		if err != nil {
			Log.Errorf("err: [%T] %s", err, err.Error())
			GinDoRespInternalErr(c)
			c.Abort()
			return
		}
		if userID == 0 {
			GinDoRespErr(c, ErrorToken, ErrorTokenMsg, nil)
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

// GinMinTokenToUserIDRedis token转换为user_id
func GinMinTokenToUserIDRedis(tx DbExeAble, redisClient *redis.Client, getUserIDByToken func(ctx context.Context, tx DbExeAble, redisClient *redis.Client, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		err := c.ShouldBind(&req)
		if err != nil {
			Log.Errorf("err: [%T] %s", err, err.Error())
			GinFillBindError(c, err)
			c.Abort()
			return
		}
		userID, err := getUserIDByToken(c, tx, redisClient, req.Token)
		if err != nil {
			Log.Errorf("err: [%T] %s", err, err.Error())
			GinDoRespInternalErr(c)
			c.Abort()
			return
		}
		if userID == 0 {
			GinDoRespErr(c, ErrorToken, ErrorTokenMsg, nil)
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

// GinMinTokenToUserIDRedisIgnore token转换为user_id
func GinMinTokenToUserIDRedisIgnore(tx DbExeAble, redisClient *redis.Client, getUserIDByToken func(ctx context.Context, tx DbExeAble, redisClient *redis.Client, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"omitempty"`
		}
		err := c.ShouldBind(&req)
		if err != nil {
			c.Next()
			return
		}
		userID, err := getUserIDByToken(c, tx, redisClient, req.Token)
		if err != nil {
			Log.Errorf("err: [%T] %s", err, err.Error())
			GinDoRespInternalErr(c)
			c.Abort()
			return
		}
		if userID == 0 {
			c.Next()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func GinCors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if len(origin) == 0 {
			// request is not a CORS request
			return
		}
		reqHeader := c.Request.Header.Get("Access-Control-Request-Headers")
		method := c.Request.Method
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Max-Age", "43200")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", reqHeader)

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

// ModelRowToStruct 填充结构体
func ModelRowToStruct(m map[string]string, v interface{}) error {
	b, err := jsoniter.Marshal(m)
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	return nil
}

// ModelRowToInterface 填充结构体
func ModelRowToInterface(m map[string]string, intCols []string, floatCols []string) (map[string]interface{}, error) {
	o := map[string]interface{}{}
	for k, v := range m {
		if IsStringInSlice(intCols, k) {
			vInt, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, err
			}
			o[k] = vInt
		} else if IsStringInSlice(floatCols, k) {
			vFloat, err := strconv.ParseFloat(v, 10)
			if err != nil {
				return nil, err
			}
			o[k] = vFloat
		} else {
			o[k] = v
		}
	}
	return o, nil
}

// ModelRowInterfaceToStruct 填充结构体
func ModelRowInterfaceToStruct(m map[string]interface{}, v interface{}) error {
	b, err := jsoniter.Marshal(m)
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	return nil
}

// ModelRowsToStruct 填充结构体
func ModelRowsToStruct(rows []map[string]string, v interface{}) error {
	b, err := jsoniter.Marshal(rows)
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	return nil
}

// ModelRowsInterfaceToStruct 填充结构体
func ModelRowsInterfaceToStruct(rows []map[string]interface{}, v interface{}) error {
	b, err := jsoniter.Marshal(rows)
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	return nil
}

// ModelRowsToInterface 填充结构体
func ModelRowsToInterface(ms []map[string]string, intCols []string, floatCols []string) ([]map[string]interface{}, error) {
	var os []map[string]interface{}
	for _, m := range ms {
		v, err := ModelRowToInterface(m, intCols, floatCols)
		if err != nil {
			return nil, err
		}
		os = append(os, v)
	}
	return os, nil
}
