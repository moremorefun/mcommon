package mcommon

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

// key 16
// AES-128
// CBC
// PKCS7Padding

// AesEncrypt 加密
func AesEncrypt(orig string, key string) (string, error) {
	// 转成字节数组
	origData := []byte(orig)
	k := []byte(key)
	// 16, 24, 32
	if len(k) < 16 {
		k = append(k, bytes.Repeat([]byte{0}, 16-len(k))...)
	} else if len(k) < 24 {
		k = append(k, bytes.Repeat([]byte{0}, 24-len(k))...)
	} else if len(k) < 32 {
		k = append(k, bytes.Repeat([]byte{0}, 32-len(k))...)
	} else {
		k = k[:32]
	}
	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)

	return base64.StdEncoding.EncodeToString(cryted), nil

}

// AesDecrypt aes解密
func AesDecrypt(cryted string, key string) (string, error) {
	// 转成字节数组
	crytedByte, _ := base64.StdEncoding.DecodeString(cryted)
	k := []byte(key)
	// 16, 24, 32
	if len(k) < 16 {
		k = append(k, bytes.Repeat([]byte{0}, 16-len(k))...)
	} else if len(k) < 24 {
		k = append(k, bytes.Repeat([]byte{0}, 24-len(k))...)
	} else if len(k) < 32 {
		k = append(k, bytes.Repeat([]byte{0}, 32-len(k))...)
	} else {
		k = k[:32]
	}

	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig = PKCS7UnPadding(orig)
	return string(orig), nil
}

// PKCS7Padding 补码
func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding 去码
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// DecryptAesEcb aes ecb 解密
func DecryptAesEcb(data, key []byte) ([]byte, error) {
	cip, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(data))
	size := cip.BlockSize()

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cip.Decrypt(decrypted[bs:be], data[bs:be])
	}
	decrypted = PKCS7UnPadding(decrypted)
	return decrypted, nil
}
