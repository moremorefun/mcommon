package mcommon

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"

	"github.com/parnurzeal/gorequest"
)

// StWxPayRawResp 回复
type StWxPayRawResp struct {
	ID           string    `json:"id"`
	CreateTime   time.Time `json:"create_time"`
	ResourceType string    `json:"resource_type"`
	EventType    string    `json:"event_type"`
	Summary      string    `json:"summary"`
	Resource     struct {
		OriginalType   string `json:"original_type"`
		Algorithm      string `json:"algorithm"`
		Ciphertext     string `json:"ciphertext"`
		AssociatedData string `json:"associated_data"`
		Nonce          string `json:"nonce"`
	} `json:"resource"`
}

// StWxPayResp 回复
type StWxPayResp struct {
	Mchid          string    `json:"mchid"`
	Appid          string    `json:"appid"`
	OutTradeNo     string    `json:"out_trade_no"`
	TransactionID  string    `json:"transaction_id"`
	TradeType      string    `json:"trade_type"`
	TradeState     string    `json:"trade_state"`
	TradeStateDesc string    `json:"trade_state_desc"`
	BankType       string    `json:"bank_type"`
	Attach         string    `json:"attach"`
	SuccessTime    time.Time `json:"success_time"`
	Payer          struct {
		Openid string `json:"openid"`
	} `json:"payer"`
	Amount struct {
		Total         int    `json:"total"`
		PayerTotal    int    `json:"payer_total"`
		Currency      string `json:"currency"`
		PayerCurrency string `json:"payer_currency"`
	} `json:"amount"`
}

// StWxRefundCb 回调
type StWxRefundCb struct {
	XMLName             xml.Name `xml:"root"`
	Text                string   `xml:",chardata"`
	OutRefundNo         string   `xml:"out_refund_no"`
	OutTradeNo          string   `xml:"out_trade_no"`
	RefundAccount       string   `xml:"refund_account"`
	RefundFee           string   `xml:"refund_fee"`
	RefundID            string   `xml:"refund_id"`
	RefundRecvAccout    string   `xml:"refund_recv_accout"`
	RefundRequestSource string   `xml:"refund_request_source"`
	RefundStatus        string   `xml:"refund_status"`
	SettlementRefundFee string   `xml:"settlement_refund_fee"`
	SettlementTotalFee  string   `xml:"settlement_total_fee"`
	SuccessTime         string   `xml:"success_time"`
	TotalFee            string   `xml:"total_fee"`
	TransactionID       string   `xml:"transaction_id"`
}

type StWxV3RefundResp struct {
	Amount struct {
		Currency         string `json:"currency"`
		DiscountRefund   int    `json:"discount_refund"`
		PayerRefund      int    `json:"payer_refund"`
		PayerTotal       int    `json:"payer_total"`
		Refund           int    `json:"refund"`
		SettlementRefund int    `json:"settlement_refund"`
		SettlementTotal  int    `json:"settlement_total"`
		Total            int    `json:"total"`
	} `json:"amount"`
	Channel             string        `json:"channel"`
	CreateTime          time.Time     `json:"create_time"`
	FundsAccount        string        `json:"funds_account"`
	OutRefundNo         string        `json:"out_refund_no"`
	OutTradeNo          string        `json:"out_trade_no"`
	PromotionDetail     []interface{} `json:"promotion_detail"`
	RefundID            string        `json:"refund_id"`
	Status              string        `json:"status"`
	TransactionID       string        `json:"transaction_id"`
	UserReceivedAccount string        `json:"user_received_account"`
	Code                string        `json:"code"`
	Message             string        `json:"message"`
}

type StWxV3RefundCb struct {
	ID           string    `json:"id"`
	CreateTime   time.Time `json:"create_time"`
	ResourceType string    `json:"resource_type"`
	EventType    string    `json:"event_type"`
	Summary      string    `json:"summary"`
	Resource     struct {
		OriginalType   string `json:"original_type"`
		Algorithm      string `json:"algorithm"`
		Ciphertext     string `json:"ciphertext"`
		AssociatedData string `json:"associated_data"`
		Nonce          string `json:"nonce"`
	} `json:"resource"`
}

type StWxV3RefundCbContent struct {
	Mchid         string    `json:"mchid"`
	OutTradeNo    string    `json:"out_trade_no"`
	TransactionID string    `json:"transaction_id"`
	OutRefundNo   string    `json:"out_refund_no"`
	RefundID      string    `json:"refund_id"`
	RefundStatus  string    `json:"refund_status"`
	SuccessTime   time.Time `json:"success_time"`
	Amount        struct {
		Total       int `json:"total"`
		Refund      int `json:"refund"`
		PayerTotal  int `json:"payer_total"`
		PayerRefund int `json:"payer_refund"`
	} `json:"amount"`
	UserReceivedAccount string `json:"user_received_account"`
}

// RsaSign 签名
func RsaSign(signContent string, privateKey *rsa.PrivateKey, hash crypto.Hash) (string, error) {
	shaNew := hash.New()
	shaNew.Write([]byte(signContent))
	hashed := shaNew.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, hash, hashed)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// WxPayV3SignStr 获取签名结果
func WxPayV3SignStr(key *rsa.PrivateKey, cols []string) (string, error) {
	var buf bytes.Buffer
	for _, col := range cols {
		buf.WriteString(col)
		buf.WriteString("\n")
	}
	sign, err := RsaSign(buf.String(), key, crypto.SHA256)
	if err != nil {
		return "", err
	}
	return sign, nil
}

// WxPayV3Sign v3签名
func WxPayV3Sign(mchid, keySerial string, key *rsa.PrivateKey, req *gorequest.SuperAgent) (*gorequest.SuperAgent, error) {
	timestamp := time.Now().Unix()
	nonce := GetUUIDStr()

	uri, err := url.Parse(req.Url)
	if err != nil {
		return nil, err
	}

	var bodyBytes []byte
	if req.Method == "POST" {
		request, err := req.MakeRequest()
		if err != nil {
			return nil, err
		}
		bodyReader, err := request.GetBody()
		if err != nil {
			return nil, err
		}
		bodyBytes, err = ioutil.ReadAll(bodyReader)
		if err != nil {
			return nil, err
		}
	}
	sign, err := WxPayV3SignStr(key, []string{
		req.Method,
		uri.Path,
		strconv.FormatInt(timestamp, 10),
		nonce,
		string(bodyBytes),
	})
	if err != nil {
		return nil, err
	}
	auth := fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%d",serial_no="%s"`,
		mchid,
		nonce,
		sign,
		timestamp,
		keySerial,
	)
	req = req.
		Set("Authorization", auth).
		Set("Accept", "application/json").
		Set("User-Agent", "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50")
	return req, nil
}

// WxPayV3Decrype 解密
func WxPayV3Decrype(key string, cipherStr, nonce, associatedData string) (string, error) {
	keyBytes := []byte(key)
	nonceBytes := []byte(nonce)
	associatedDataBytes := []byte(associatedData)
	ciphertext, err := base64.StdEncoding.DecodeString(cipherStr)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plaintext, err := aesgcm.Open(nil, nonceBytes, ciphertext, associatedDataBytes)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// WxPayV3CheckSign v3签名验证
func WxPayV3CheckSign(header map[string][]string, body []byte, cerStr string) error {
	if len(cerStr) == 0 {
		return fmt.Errorf("no cer")
	}
	timestamp, err := WxPayV3GetHeaderByKey(header, "Wechatpay-Timestamp")
	if err != nil {
		return err
	}
	nonce, err := WxPayV3GetHeaderByKey(header, "Wechatpay-Nonce")
	if err != nil {
		return err
	}
	signature, err := WxPayV3GetHeaderByKey(header, "Wechatpay-Signature")
	if err != nil {
		return err
	}
	checkStr := timestamp + "\n" + nonce + "\n" + string(body) + "\n"

	block, _ := pem.Decode([]byte(cerStr))
	var cert *x509.Certificate
	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)

	oldSign, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256([]byte(checkStr))
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], oldSign)
	return err
}

// WxPayV3GetHeaderByKey 获取头
func WxPayV3GetHeaderByKey(header map[string][]string, key string) (string, error) {
	v, ok := header[key]
	if !ok {
		return "", fmt.Errorf("no key %s", key)
	}
	if len(v) == 0 {
		return "", fmt.Errorf("key empty %s", key)
	}
	return v[0], nil
}

// WxPayV3GetPrepay 获取预支付信息
func WxPayV3GetPrepay(keySerial string, key *rsa.PrivateKey, appID, mchID, openID, payBody, outTradeNo, cbURL string, totalFee int64, expireAt time.Time) (gin.H, string, error) {
	req := gorequest.New().
		Post("https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi").
		Send(
			H{
				"appid":        appID,
				"mchid":        mchID,
				"description":  payBody,
				"out_trade_no": outTradeNo,
				"time_expire":  expireAt.Format(time.RFC3339),
				"notify_url":   cbURL,
				"amount": H{
					"total": totalFee,
				},
				"payer": H{
					"openid": openID,
				},
			},
		)
	req, err := WxPayV3Sign(
		mchID,
		keySerial,
		key,
		req,
	)
	if err != nil {
		return nil, "", err
	}
	_, body, errs := req.EndBytes()
	if errs != nil {
		return nil, "", errs[0]
	}
	var prepayResp struct {
		PrepayID string `json:"prepay_id"`
	}
	err = jsoniter.Unmarshal(body, &prepayResp)
	if err != nil {
		return nil, "", err
	}
	if len(prepayResp.PrepayID) == 0 {
		return nil, "", fmt.Errorf("get prepay id err: %s", body)
	}
	v, err := WxPayV3SignPrepayid(key, appID, prepayResp.PrepayID)
	if err != nil {
		return nil, "", err
	}
	return v, prepayResp.PrepayID, nil
}

// WxPayV3SignPrepayid 签名prepayid
func WxPayV3SignPrepayid(key *rsa.PrivateKey, appID, prepayid string) (gin.H, error) {
	objTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	objNonce := GetUUIDStr()
	objCol := fmt.Sprintf("prepay_id=%s", prepayid)
	objSign, err := WxPayV3SignStr(
		key,
		[]string{
			appID,
			objTimestamp,
			objNonce,
			objCol,
		},
	)
	if err != nil {
		return nil, err
	}
	v := gin.H{
		"timeStamp": objTimestamp,
		"nonceStr":  objNonce,
		"package":   objCol,
		"signType":  "RSA",
		"paySign":   objSign,
	}
	return v, nil
}

// WxPayV3DecodePayResp 解析支付回调
func WxPayV3DecodePayResp(v3Key string, body []byte, mchid, appid string) (*StWxPayResp, error) {
	var rawResp StWxPayRawResp
	err := jsoniter.Unmarshal(body, &rawResp)
	if err != nil {
		return nil, err
	}
	if rawResp.EventType != "TRANSACTION.SUCCESS" {
		return nil, fmt.Errorf("error event_type: %s", rawResp.EventType)
	}
	if rawResp.ResourceType != "encrypt-resource" {
		return nil, fmt.Errorf("error resource_type: %s", rawResp.ResourceType)
	}
	originalType := rawResp.Resource.OriginalType
	if originalType != "transaction" {
		return nil, fmt.Errorf("error original_type: %s", originalType)
	}
	algorithm := rawResp.Resource.Algorithm
	if algorithm != "AEAD_AES_256_GCM" {
		return nil, fmt.Errorf("error algorithm: %s", algorithm)
	}

	ciphertext := rawResp.Resource.Ciphertext
	associatedData := rawResp.Resource.AssociatedData
	nonce := rawResp.Resource.Nonce

	plain, err := WxPayV3Decrype(
		v3Key,
		ciphertext,
		nonce,
		associatedData,
	)
	if err != nil {
		return nil, err
	}
	var finalResp StWxPayResp
	err = jsoniter.Unmarshal([]byte(plain), &finalResp)
	if err != nil {
		return nil, err
	}
	if finalResp.Mchid != mchid {
		return nil, fmt.Errorf("mchid error")
	}
	if finalResp.Appid != appid {
		return nil, fmt.Errorf("appid error")
	}
	if finalResp.TradeState != "SUCCESS" {
		return nil, fmt.Errorf("error trade_state: %s", finalResp.TradeState)
	}
	return &finalResp, nil
}

// WxPayCheckRefundCb 验证回调
func WxPayCheckRefundCb(mchKey string, body []byte) (*StWxRefundCb, error) {
	mchKeyMd5 := fmt.Sprintf("%x", md5.Sum([]byte(mchKey)))
	bodyMap, err := XMLWalk(body)
	if err != nil {
		// 返回数据
		return nil, err
	}
	reqInfo, ok := bodyMap["req_info"]
	if !ok {
		return nil, fmt.Errorf("no key req_info %s", body)
	}
	reqInfoStr, ok := reqInfo.(string)
	if !ok {
		return nil, fmt.Errorf("error format req_info: %s", body)
	}
	reqInfoBytes, err := base64.StdEncoding.DecodeString(reqInfoStr)
	if err != nil {
		return nil, err
	}
	reqInfoFull, err := DecryptAesEcb(reqInfoBytes, []byte(mchKeyMd5))
	if err != nil {
		return nil, err
	}
	var bodyXML StWxRefundCb
	err = xml.Unmarshal(reqInfoFull, &bodyXML)
	if err != nil {
		return nil, err
	}
	return &bodyXML, nil
}

// WxPayV3Refunds 退款
func WxPayV3Refunds(keySerial string, key *rsa.PrivateKey, mchID, transactionID, outRefundNo, cbURL string, totalFee, refundFee int64) (*StWxV3RefundResp, error) {
	req := gorequest.New().
		Post("https://api.mch.weixin.qq.com/v3/refund/domestic/refunds").
		Send(
			H{
				"transaction_id": transactionID,
				"out_refund_no":  outRefundNo,
				"notify_url":     cbURL,
				"amount": H{
					"refund":   refundFee,
					"total":    totalFee,
					"currency": "CNY",
				},
			},
		)
	req, err := WxPayV3Sign(
		mchID,
		keySerial,
		key,
		req,
	)
	if err != nil {
		return nil, err
	}
	_, body, errs := req.EndBytes()
	if errs != nil {
		return nil, errs[0]
	}
	Log.Debugf("body: %s", body)
	var resp StWxV3RefundResp
	err = jsoniter.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Code != "" {
		return nil, fmt.Errorf("refund err: %s", body)
	}
	return &resp, nil
}

// WxPayV3DecodeRefundsCb 解析退款回调
func WxPayV3DecodeRefundsCb(v3Key string, body []byte) (*StWxV3RefundCbContent, error) {
	var rawResp StWxV3RefundCb
	err := jsoniter.Unmarshal(body, &rawResp)
	if err != nil {
		return nil, err
	}
	if rawResp.EventType != "REFUND.SUCCESS" {
		return nil, fmt.Errorf("error event_type: %s", rawResp.EventType)
	}
	if rawResp.ResourceType != "encrypt-resource" {
		return nil, fmt.Errorf("error resource_type: %s", rawResp.ResourceType)
	}
	originalType := rawResp.Resource.OriginalType
	if originalType != "refund" {
		return nil, fmt.Errorf("error original_type: %s", originalType)
	}
	algorithm := rawResp.Resource.Algorithm
	if algorithm != "AEAD_AES_256_GCM" {
		return nil, fmt.Errorf("error algorithm: %s", algorithm)
	}

	ciphertext := rawResp.Resource.Ciphertext
	associatedData := rawResp.Resource.AssociatedData
	nonce := rawResp.Resource.Nonce

	plain, err := WxPayV3Decrype(
		v3Key,
		ciphertext,
		nonce,
		associatedData,
	)
	if err != nil {
		return nil, err
	}
	Log.Debugf("plain: %s", plain)
	var content StWxV3RefundCbContent
	err = jsoniter.Unmarshal([]byte(plain), &content)
	if err != nil {
		return nil, err
	}
	return &content, nil
}
