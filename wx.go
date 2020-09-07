package mcommon

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
)

// WxJsCodeResp jscode回复
type WxJsCodeResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	Unionid    string `json:"unionid"`
	Errcode    int64  `json:"errcode"`
	Errmsg     string `json:"errmsg"`
}

// WxJsCode js登录
func WxJsCode(appID, appSecret, code string) (*WxJsCodeResp, error) {
	retryCount := 0
GotoHttpRetry:
	_, body, errs := gorequest.New().
		Get("https://api.weixin.qq.com/sns/jscode2session").
		Query(gin.H{
			"appid":      appID,
			"secret":     appSecret,
			"js_code":    code,
			"grant_type": "authorization_code",
		}).EndBytes()
	if errs != nil {
		retryCount++
		if retryCount < 3 {
			goto GotoHttpRetry
		}
		return nil, errs[0]
	}
	var apiResp WxJsCodeResp
	err := json.Unmarshal(body, &apiResp)
	if err != nil {
		retryCount++
		if retryCount < 3 {
			goto GotoHttpRetry
		}
		return nil, fmt.Errorf("json err: %w %s", err, body)
	}
	if 0 != apiResp.Errcode {
		return nil, fmt.Errorf("wx jscode err: %s", apiResp.Errmsg)
	}
	return &apiResp, nil
}
