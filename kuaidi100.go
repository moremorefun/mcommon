package mcommon

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
)

type StKuaidi100Resp struct {
	Result     bool   `json:"result"`
	ReturnCode string `json:"returnCode"`
	Message    string `json:"message"`
}

// Poll 订阅邮件推送
func Kuaidi100Poll(key, company, number, tel, callbackurl string) (*StKuaidi100Resp, error) {
	retryCount := 0
	reqBody := gin.H{
		"company": company,
		"number":  number,
		"key":     key,
		"parameters": gin.H{
			"callbackurl": callbackurl,
			"resultv2":    1,
			"phone":       tel,
		},
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
GotoHttpRetry:

	_, body, errs := gorequest.New().
		Post("https://poll.kuaidi100.com/poll").
		Send("schema=json").
		Send(fmt.Sprintf("param=%s", reqBytes)).
		EndBytes()
	if errs != nil {
		retryCount++
		if retryCount < 3 {
			goto GotoHttpRetry
		}
		return nil, errs[0]
	}
	var apiResp StKuaidi100Resp
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		retryCount++
		if retryCount < 3 {
			goto GotoHttpRetry
		}
		return nil, fmt.Errorf("json err: %w %s", err, body)
	}
	if "200" != apiResp.ReturnCode && "501" != apiResp.ReturnCode {
		return nil, fmt.Errorf("wx jscode err: %s %s", apiResp.ReturnCode, apiResp.Message)
	}
	return &apiResp, nil
}
