package mcommon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"

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

// WxAppCodeResp jscode回复
type WxAppCodeResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Openid       string `json:"openid"`
	Scope        string `json:"scope"`
	ErrCode      int64  `json:"errcode"`
}

// WxAppUserInfoResp userinfo回复
type WxAppUserInfoResp struct {
	Openid     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        uint64 `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Headimgurl string `json:"headimgurl"`
	Unionid    string `json:"unionid"`
	ErrCode    int64  `json:"errcode"`
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

// WxAppCode app登录
func WxAppCode(appID, appSecret, code string) (*WxAppCodeResp, error) {
	retryCount := 0
GotoHttpRetry:
	_, body, errs := gorequest.New().
		Get("https://api.weixin.qq.com/sns/oauth2/access_token").
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
	var apiResp WxAppCodeResp
	err := json.Unmarshal(body, &apiResp)
	if err != nil {
		retryCount++
		if retryCount < 3 {
			goto GotoHttpRetry
		}
		return nil, fmt.Errorf("json err: %w %s", err, body)
	}
	if 0 != apiResp.ErrCode {
		return nil, fmt.Errorf("wx jscode err: %d", apiResp.ErrCode)
	}
	return &apiResp, nil
}

// WxAppUserInfo 用户信息
func WxAppUserInfo(accessToken, openID string) (*WxAppUserInfoResp, error) {
	retryCount := 0
GotoHttpRetry:
	_, body, errs := gorequest.New().
		Get("https://api.weixin.qq.com/sns/userinfo").
		Query(gin.H{
			"access_token":      accessToken,
			"openid":     openID,
			"lang":     "zh_CN",,
		}).EndBytes()
	if errs != nil {
		retryCount++
		if retryCount < 3 {
			goto GotoHttpRetry
		}
		return nil, errs[0]
	}
	var apiResp WxAppUserInfoResp
	err := json.Unmarshal(body, &apiResp)
	if err != nil {
		retryCount++
		if retryCount < 3 {
			goto GotoHttpRetry
		}
		return nil, fmt.Errorf("json err: %w %s", err, body)
	}
	if 0 != apiResp.ErrCode {
		return nil, fmt.Errorf("wx jscode err: %d", apiResp.ErrCode)
	}
	return &apiResp, nil
}

// SQLRedisGetWxToken 获取小程序token
func SQLRedisGetWxToken(c context.Context, tx DbExeAble, redisClient *redis.Client, appID string,
	funcSQLGetToken func(context.Context, DbExeAble, string) (string, string, int64, error),
	funcSQLSetToken func(context.Context, DbExeAble, string, string, string, int64) error,
) (string, error) {
	redisKey := fmt.Sprintf("wx_token_%s", appID)
	token, err := RedisGet(
		c,
		redisClient,
		redisKey,
	)
	if err != nil {
		Log.Errorf("err: [%T] %s", err, err.Error())
	} else {
		if "" != token {
			return token, nil
		}
	}
	type apiRespSt struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Errcode     int64  `json:"errcode"`
		Errmsg      string `json:"errmsg"`
	}
	secret, accessToken, expiresAt, err := funcSQLGetToken(
		c,
		tx,
		appID,
	)
	if err != nil {
		return "", err
	}
	if secret == "" {
		return "", fmt.Errorf("no app of %s", appID)
	}
	// 1. 在有效期内
	if expiresAt > time.Now().Unix() {
		return accessToken, nil
	}
	_, body, errs := gorequest.New().Get("https://api.weixin.qq.com/cgi-bin/token").Query(gin.H{
		"appid":      appID,
		"secret":     secret,
		"grant_type": "client_credential",
	}).End()
	if errs != nil {
		return "", errs[0]
	}
	var apiResp apiRespSt
	err = json.Unmarshal([]byte(body), &apiResp)
	if err != nil {
		return "", err
	}
	if 0 != apiResp.Errcode {
		return "", errors.New("api resp error")
	}
	err = funcSQLSetToken(
		c,
		tx,
		appID,
		secret,
		accessToken,
		time.Now().Unix()+apiResp.ExpiresIn-10*60,
	)
	if err != nil {
		return "", err
	}
	err = RedisSet(
		c,
		redisClient,
		redisKey,
		apiResp.AccessToken,
		time.Second*time.Duration(apiResp.ExpiresIn-10*60),
	)
	if err != nil {
		Log.Errorf("err: [%T] %s", err, err.Error())
	}
	return apiResp.AccessToken, nil
}

// SQLRedisRestWxToken 重置小程序token
func SQLRedisRestWxToken(c context.Context, tx DbExeAble, redisClient *redis.Client, appID string,
	funcSQLResetToken func(context.Context, DbExeAble, string) error,
) {
	redisKey := fmt.Sprintf("wx_token_%s", appID)
	err := RedisRm(
		c,
		redisClient,
		redisKey,
	)
	if err != nil {
		Log.Errorf("err: [%T] %s", err, err.Error())
	}
	err = funcSQLResetToken(
		c,
		tx,
		appID,
	)
	if err != nil {
		Log.Errorf("err: [%T] %s", err, err.Error())
	}
}
