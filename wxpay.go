package mcommon

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
)

// StWeChatCbBody 回调信息
type StWeChatCbBody struct {
	XMLName       xml.Name `xml:"xml"`
	Text          string   `xml:",chardata"`
	Appid         string   `xml:"appid"`
	Attach        string   `xml:"attach"`
	BankType      string   `xml:"bank_type"`
	FeeType       string   `xml:"fee_type"`
	IsSubscribe   string   `xml:"is_subscribe"`
	MchID         string   `xml:"mch_id"`
	NonceStr      string   `xml:"nonce_str"`
	Openid        string   `xml:"openid"`
	OutTradeNo    string   `xml:"out_trade_no"`
	ResultCode    string   `xml:"result_code"`
	ReturnCode    string   `xml:"return_code"`
	Sign          string   `xml:"sign"`
	TimeEnd       string   `xml:"time_end"`
	TotalFee      string   `xml:"total_fee"`
	CouponFee     string   `xml:"coupon_fee"`
	CouponCount   string   `xml:"coupon_count"`
	CouponType    string   `xml:"coupon_type"`
	CouponID      string   `xml:"coupon_id"`
	TradeType     string   `xml:"trade_type"`
	TransactionID string   `xml:"transaction_id"`
}

// WechatGetPrepay 获取预支付信息
func WechatGetPrepay(appID, mchID, mchKey, payBody, clientIP, cbURL, tradeType, openID string, totalFee int64) (gin.H, error) {
	retryCount := 0
	nonce := GetUUIDStr()
	outTradeNo := GetUUIDStr()
	sendBody := gin.H{
		"appid":            appID,
		"mch_id":           mchID,
		"nonce_str":        nonce,
		"body":             payBody,
		"out_trade_no":     outTradeNo,
		"total_fee":        totalFee,
		"spbill_create_ip": clientIP,
		"notify_url":       cbURL,
		"trade_type":       tradeType,
		"openid":           openID,
	}
	sendBody["sign"] = WechatGetSign(mchKey, sendBody)
	sendBodyBs, err := xml.Marshal(sendBody)
	if err != nil {
		return nil, err
	}
GotoHttpRetry:
	_, body, errs := gorequest.New().
		Post("https://api.mch.weixin.qq.com/pay/unifiedorder").
		Set("Content-Type", "application/xml").
		Send(string(sendBodyBs)).
		EndBytes()
	if errs != nil {
		retryCount++
		if retryCount < 3 {
			goto GotoHttpRetry
		}
		return nil, errs[0]
	}
	respMap, err := XMLWalk(body)
	if err != nil {
		return nil, err
	}
	if !WechatCheckSign(mchKey, respMap) {
		return nil, fmt.Errorf("sign error")
	}
	type stRespXMLXML struct {
		XMLName    xml.Name `xml:"xml"`
		Text       string   `xml:",chardata"`
		ReturnCode string   `xml:"return_code"`
		ReturnMsg  string   `xml:"return_msg"`
		Appid      string   `xml:"appid"`
		MchID      string   `xml:"mch_id"`
		NonceStr   string   `xml:"nonce_str"`
		Sign       string   `xml:"sign"`
		ResultCode string   `xml:"result_code"`
		PrepayID   string   `xml:"prepay_id"`
		TradeType  string   `xml:"trade_type"`
	}
	var respXML stRespXMLXML
	err = xml.Unmarshal(body, &respXML)
	if err != nil {
		return nil, err
	}
	if respXML.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("resp result code error")
	}
	jsMap := gin.H{
		"appId":     appID,
		"timeStamp": time.Now().Unix(),
		"nonceStr":  GetUUIDStr(),
		"package":   fmt.Sprintf("%s=%s", "prepay_id", respXML.PrepayID),
		"signType":  "MD5",
	}
	jsMap["paySign"] = WechatGetSign(mchKey, jsMap)
	return jsMap, nil
}

// WechatCheckCb 验证回调
func WechatCheckCb(mchKey string, body []byte) (*StWeChatCbBody, error) {
	bodyMap, err := XMLWalk(body)
	if err != nil {
		// 返回数据
		return nil, err
	}
	if !WechatCheckSign(mchKey, bodyMap) {
		return nil, fmt.Errorf("sign error")
	}

	var bodyXML StWeChatCbBody
	err = xml.Unmarshal(body, &bodyXML)
	if err != nil {
		return nil, err
	}
	if bodyXML.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("result code error")
	}
	return &bodyXML, nil
}
