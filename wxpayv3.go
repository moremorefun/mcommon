package mcommon

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"time"

	"github.com/parnurzeal/gorequest"
)

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
	cert, _ = x509.ParseCertificate(block.Bytes)
	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)

	oldSign, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256([]byte(checkStr))
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], oldSign)
	return err
}

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
func WxPayV3GetPrepay(keySerial string, key *rsa.PrivateKey, appID, mchID, openID, payBody, outTradeNo, cbURL string, totalFee int64) (H, error) {
	req := gorequest.New().
		Post("https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi").
		Send(
			H{
				"appid":        appID,
				"mchid":        mchID,
				"description":  payBody,
				"out_trade_no": outTradeNo,
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
		return nil, err
	}
	_, body, errs := req.EndBytes()
	if errs != nil {
		return nil, errs[0]
	}
	Log.Debugf("body: %s", body)
	var prepayResp struct {
		PrepayID string `json:"prepay_id"`
	}
	err = json.Unmarshal(body, &prepayResp)
	if len(prepayResp.PrepayID) == 0 {
		return nil, fmt.Errorf("get prepay id err: %s", body)
	}

	objTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	objNonce := GetUUIDStr()
	objCol := fmt.Sprintf("prepay_id=%s", prepayResp.PrepayID)
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
	v := H{
		"timeStamp": objTimestamp,
		"nonceStr":  objNonce,
		"package":   objCol,
		"signType":  "RSA",
		"paySign":   objSign,
	}
	return v, nil
}
