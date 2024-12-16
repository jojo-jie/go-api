package service

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestDes(t *testing.T) {
	AesDecrypt("1111==", "111")
}

func AesDecrypt(msgSecret, appSecret string) string {
	var appSecretArr = []byte(strings.ReplaceAll(appSecret, "-", ""))
	bytesPass, err := base64.StdEncoding.DecodeString(msgSecret)
	if err != nil {
		fmt.Println(err)
		return "解密失败！！！"
	}
	sourceMsg, err := DoAesDecrypt(bytesPass, appSecretArr)
	if err != nil {
		fmt.Println(err)
		return "解密失败！！！"
	}
	fmt.Printf("解密后:%s\n", sourceMsg)
	return string(sourceMsg)
}
func DoAesDecrypt(encryptedMsg, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	fmt.Println("blockSize:", key[:blockSize])
	fmt.Println(string(key[:blockSize]))
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(encryptedMsg))
	blockMode.CryptBlocks(origData, encryptedMsg)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

// 去除填充数据
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unfilledNum := int(origData[length-1])
	return origData[:(length - unfilledNum)]
}
