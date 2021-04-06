package util

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"time"
)

const (
	TTL  = 120
	SIZE = 15 * 1024 * 1024
)

type OssInfo struct {
	AccessId  string `json:"access_id"`
	Host      string `json:"host"`
	TTL       int    `json:"ttl"`
	Policy    string `json:"policy"`
	Signature string `json:"signature"`
	Dir       string `json:"dir"`
}

func NewOss() *OssInfo {
	var buf strings.Builder
	buf.WriteString("https://")
	buf.WriteString(os.Getenv("OSS_BUCKET"))
	buf.WriteString(".")
	buf.WriteString(strings.TrimLeft(os.Getenv("OSS_END_POINT"), "https://"))
	o := &OssInfo{
		AccessId: os.Getenv("OSS_ACCESS_KEY_ID"),
		Host:     buf.String(),
		TTL:      SIZE,
	}
	return o
}

type policy struct {
	Expiration string `json:"expiration"`
	Conditions []interface{} `json:"conditions"`
}

func (o *OssInfo) Info(d string) (*OssInfo, error) {
	now := time.Now()
	expiration := now.Add(TTL * time.Second).UTC().Format(time.RFC3339)
	conditions := make([]interface{}, 2, 2)
	cond := []interface{}{"content-length-range",0,SIZE}
	conditions[0] = cond
	start := []interface{}{"starts-with", "$key", d}
	conditions[1] = start
	p := policy{
		Expiration: expiration,
		Conditions: conditions,
	}
	policyJson, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyJson)
	mac := hmac.New(sha1.New, []byte(os.Getenv("OSS_ACCESS_KEY_SECRET")))
	mac.Write([]byte(policyBase64))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	o.Policy = policyBase64
	o.Signature = signature
	o.Dir = d
	return o, nil
}
