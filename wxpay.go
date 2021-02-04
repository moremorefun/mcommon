package mcommon

import (
	"crypto/tls"
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

// StRefundRespXML 回复内容
type StRefundRespXML struct {
	XMLName           xml.Name `xml:"xml"`
	Text              string   `xml:",chardata"`
	ReturnCode        string   `xml:"return_code"`
	ReturnMsg         string   `xml:"return_msg"`
	Appid             string   `xml:"appid"`
	MchID             string   `xml:"mch_id"`
	NonceStr          string   `xml:"nonce_str"`
	Sign              string   `xml:"sign"`
	ResultCode        string   `xml:"result_code"`
	TransactionID     string   `xml:"transaction_id"`
	OutTradeNo        string   `xml:"out_trade_no"`
	OutRefundNo       string   `xml:"out_refund_no"`
	RefundID          string   `xml:"refund_id"`
	RefundChannel     string   `xml:"refund_channel"`
	RefundFee         int64    `xml:"refund_fee"`
	CouponRefundFee   int64    `xml:"coupon_refund_fee"`
	TotalFee          int64    `xml:"total_fee"`
	CashFee           int64    `xml:"cash_fee"`
	CouponRefundCount int64    `xml:"coupon_refund_count"`
	CashRefundFee     int64    `xml:"cash_refund_fee"`
}

// WechatGetPrepay 获取预支付信息
func WechatGetPrepay(appID, mchID, mchKey, payBody, outTradeNo, clientIP, cbURL, tradeType, openID string, totalFee int64) (gin.H, error) {
	retryCount := 0
	nonce := GetUUIDStr()
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
		return nil, fmt.Errorf("sign error: %s", body)
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

// WechatRefund 申请退款
func WechatRefund(cer tls.Certificate, appID, mchID, mchKey, transactionID, outRefundNo, cbURL string, totalFee, refundFee int64) (*StRefundRespXML, error) {
	retryCount := 0
	nonce := GetUUIDStr()
	sendBody := gin.H{
		"appid":          appID,
		"mch_id":         mchID,
		"nonce_str":      nonce,
		"transaction_id": transactionID,
		"out_refund_no":  outRefundNo,
		"total_fee":      totalFee,
		"refund_fee":     refundFee,
		"notify_url":     cbURL,
	}
	sendBody["sign"] = WechatGetSign(mchKey, sendBody)
	sendBodyBs, err := xml.Marshal(sendBody)
	if err != nil {
		return nil, err
	}
GotoHttpRetry:
	_, body, errs := gorequest.New().
		Post("https://api.mch.weixin.qq.com/secapi/pay/refund").
		TLSClientConfig(&tls.Config{Certificates: []tls.Certificate{cer}}).
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
		return nil, fmt.Errorf("sign error: %s", body)
	}
	var respXML StRefundRespXML
	err = xml.Unmarshal(body, &respXML)
	if err != nil {
		return nil, err
	}
	if respXML.ReturnCode != "SUCCESS" {
		Log.Errorf("refund err: %s", body)
		return nil, fmt.Errorf("resp return code error %s", respXML.ReturnCode)
	}
	if respXML.ResultCode != "SUCCESS" {
		Log.Errorf("refund err: %s", body)
		return nil, fmt.Errorf("resp result code error %s", respXML.ResultCode)
	}
	return &respXML, nil
}
