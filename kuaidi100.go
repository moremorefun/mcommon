package mcommon

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
)

const (
	/*
		0	在途	快件处于运输过程中
		1	揽收	快件已由快递公司揽收
		2	疑难	快递100无法解析的状态，或者是需要人工介入的状态， 比方说收件人电话错误。
		3	签收	正常签收
		4	退签	货物退回发货人并签收
		5	派件	货物正在进行派件
		6	退回	货物正处于返回发货人的途中
		7	转投	货物转给其他快递公司邮寄
		10	待清关	货物等待清关
		11	清关中	货物正在清关流程中
		12	已清关	货物已完成清关流程
		13	清关异常	货物在清关过程中出现异常
		14	拒签	收件人明确拒收
	*/
	Kuaidi100StateOnTheWay       = 0
	Kuaidi100StateCollect        = 1
	Kuaidi100StateDifficult      = 2
	Kuaidi100StateSignFor        = 3
	Kuaidi100StateReturnSignFor  = 4
	Kuaidi100StateDispatch       = 5
	Kuaidi100StateReturnOnTheWay = 6
	Kuaidi100StateSwitching      = 7
	Kuaidi100StateToBeCleared    = 10
	Kuaidi100StateClearing       = 11
	Kuaidi100StateCleared        = 12
	Kuaidi100StateClearError     = 13
	Kuaidi100StateReject         = 14
)

type StKuaidi100PollResp struct {
	Result     bool   `json:"result"`
	ReturnCode string `json:"returnCode"`
	Message    string `json:"message"`
}

type StKuaidi100GetResp struct {
	Message string `json:"message"`
	Nu      string `json:"nu"`
	Ischeck string `json:"ischeck"`
	Com     string `json:"com"`
	Status  string `json:"status"`
	Data    []struct {
		Time     string `json:"time"`
		Context  string `json:"context"`
		Ftime    string `json:"ftime"`
		AreaCode string `json:"areaCode"`
		AreaName string `json:"areaName"`
		Status   string `json:"status"`
	} `json:"data"`
	State     string `json:"state"`
	Condition string `json:"condition"`
	RouteInfo struct {
		From struct {
			Number string `json:"number"`
			Name   string `json:"name"`
		} `json:"from"`
		Cur struct {
			Number string `json:"number"`
			Name   string `json:"name"`
		} `json:"cur"`
		To struct {
			Number string `json:"number"`
			Name   string `json:"name"`
		} `json:"to"`
	} `json:"routeInfo"`
	IsLoop bool `json:"isLoop"`
}

type StKuaidi100CbBody struct {
	Message    string `json:"message"`
	ComOld     string `json:"comOld"`
	Status     string `json:"status"`
	LastResult struct {
		Nu        string `json:"nu"`
		Message   string `json:"message"`
		Ischeck   string `json:"ischeck"`
		Com       string `json:"com"`
		Condition string `json:"condition"`
		Status    string `json:"status"`
		State     string `json:"state"`
		Data      []struct {
			Time     string `json:"time"`
			AreaName string `json:"areaName,omitempty"`
			Status   string `json:"status"`
			AreaCode string `json:"areaCode,omitempty"`
			Context  string `json:"context"`
			Ftime    string `json:"ftime"`
		} `json:"data"`
	} `json:"lastResult"`
	ComNew     string `json:"comNew"`
	Billstatus string `json:"billstatus"`
	AutoCheck  string `json:"autoCheck"`
}

// Kuaidi100Poll 订阅邮件推送
func Kuaidi100Poll(key, company, number, tel, callbackurl string) (*StKuaidi100PollResp, error) {
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
	var apiResp StKuaidi100PollResp
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

// Kuaidi100Query 获取快递信息
func Kuaidi100Query(customer, key, company, number, tel string) (*StKuaidi100GetResp, error) {
	retryCount := 0
	reqBody := gin.H{
		"com":      company,
		"num":      number,
		"phone":    tel,
		"resultv2": 2,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	rawStr := fmt.Sprintf("%s%s%s", reqBytes, key, customer)
	data := []byte(rawStr)
	r := md5.Sum(data)
	signedString := hex.EncodeToString(r[:])
	sign := strings.ToUpper(signedString)
GotoHttpRetry:
	_, body, errs := gorequest.New().
		Post("https://poll.kuaidi100.com/poll/query.do").
		Send(fmt.Sprintf("customer=%s", customer)).
		Send(fmt.Sprintf("sign=%s", sign)).
		Send(fmt.Sprintf("param=%s", reqBytes)).
		EndBytes()
	if errs != nil {
		retryCount++
		if retryCount < 3 {
			goto GotoHttpRetry
		}
		return nil, errs[0]
	}
	var apiResp StKuaidi100GetResp
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		retryCount++
		if retryCount < 3 {
			goto GotoHttpRetry
		}
		return nil, fmt.Errorf("json err: %w %s", err, body)
	}
	return &apiResp, nil
}
