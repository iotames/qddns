package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/iotames/miniutils"
)

func CreateClientAli(accessKeyId *string, accessKeySecret *string) (_result *alidns20150109.Client, _err error) {
	config := &openapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: accessKeyId,
		// 必填，您的 AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Alidns
	config.Endpoint = tea.String("alidns.cn-hangzhou.aliyuncs.com")
	_result = &alidns20150109.Client{}
	_result, _err = alidns20150109.NewClient(config)
	return _result, _err
}

func GetClientAli() *alidns20150109.Client {
	client, err := CreateClientAli(tea.String(AliKey), tea.String(AliSecret))
	if err != nil {
		panic(err)
	}
	return client
}

func CheckAliDNS(h func(rr alidns20150109.DescribeDomainRecordsResponseBodyDomainRecordsRecord)) {
	logger := miniutils.GetLogger("")
	if time.Since(sipInfo.DnsIpUpdatedAt) < time.Second*7200 {
		realIP := GetRealIP()
		if realIP != "" && realIP == sipInfo.DnsIp {
			logger.Info("---Skip--CheckAliDNS--sipInfo.DnsIpUpdatedAt---In(3600s)(realIP == sipInfo.DnsIp)--")
			return
		}
	}
	describeDomainRecordsRequest := &alidns20150109.DescribeDomainRecordsRequest{DomainName: &DomainName}
	runtime := &util.RuntimeOptions{}
	client := GetClientAli()
	result, err := client.DescribeDomainRecordsWithOptions(describeDomainRecordsRequest, runtime)
	if err != nil {
		panic(err)
	}
	bd := result.Body
	statusCode := *result.StatusCode
	if statusCode != 200 {
		logger.Error(fmt.Sprintf("---CheckAliDNS--Err--statusCode = %d---bd(%+v)", statusCode, *bd))
		return
	}
	// fmt.Printf("---statusCode(%d)--Total(%d)--PageNumber(%d)--PageSize(%d)--\n", statusCode, *bd.TotalCount, *bd.PageNumber, *bd.PageSize)
	sipInfo.DnsIpUpdatedAt = time.Now()
	rrs := bd.DomainRecords.Record
	for _, r := range rrs {
		h(*r)
	}
}

func UpdateAliDNS(rr alidns20150109.DescribeDomainRecordsResponseBodyDomainRecordsRecord) {
	logger := miniutils.GetLogger("")
	updateDomainRecordRequest := &alidns20150109.UpdateDomainRecordRequest{
		RecordId: rr.RecordId,
		RR:       rr.RR, Type: rr.Type, Value: rr.Value}
	runtime := &util.RuntimeOptions{}
	client := GetClientAli()
	result, err := client.UpdateDomainRecordWithOptions(updateDomainRecordRequest, runtime)
	if err != nil {
		panic(err)
	}
	bd := result.Body
	statusCode := *result.StatusCode
	if statusCode != 200 {
		logger.Error(fmt.Sprintf("---Fail--statusCode = %d---bd(%+v)", statusCode, *bd))
		return
	}
	logger.Info(fmt.Sprintf("---Success--UpdateAliDNS--Domain(%s.%s)->Value(%s)--", *rr.RR, *rr.DomainName, *rr.Value))
}

type HttpBinResult struct {
	Origin string
}

type ServerIpInfo struct {
	DnsIpUpdatedAt  time.Time
	DnsIp           string
	RealIpUpdatedAt time.Time
	RealIp          string
}

var sipInfo *ServerIpInfo

func GetRealIP() string {
	if time.Since(sipInfo.RealIpUpdatedAt) < time.Second*60 && sipInfo.RealIp != "" {
		fmt.Printf("---UseCache---RealIP(%+v)", sipInfo)
		return sipInfo.RealIp
	}
	logger := miniutils.GetLogger("")
	c := http.DefaultClient
	hreq, _ := http.NewRequest("GET", "https://httpbin.org/get", nil)
	resp, err := c.Do(hreq)
	if err != nil {
		logger.Warn("request https://httpbin.org/get err:", err)
		return ""
	}
	defer resp.Body.Close()
	bd, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("response https://httpbin.org/get err:", err)
		return ""
	}
	hres := HttpBinResult{}
	err = json.Unmarshal(bd, &hres)
	if err != nil {
		logger.Warn("response https://httpbin.org/get json.Unmarshal err:", err)
		return ""
	}
	sipInfo.RealIpUpdatedAt = time.Now()
	sipInfo.RealIp = hres.Origin
	return sipInfo.RealIp
}

func handleAliDNS(rr alidns20150109.DescribeDomainRecordsResponseBodyDomainRecordsRecord) {
	logger := miniutils.GetLogger("")
	si := miniutils.GetIndexOf[string](*rr.RR, SubDomains)
	if si > -1 {
		logger.Debug(fmt.Sprintf("---match--(%+v)---", rr))
		oldValue := *rr.Value
		realip := GetRealIP()
		sipInfo.DnsIp = oldValue
		if realip == "" {
			logger.Warn("---realIP is empty-----")
			return
		}
		if realip == oldValue {
			logger.Warn("----realip == oldValue")
			return
		}
		// logger.Warn(fmt.Sprintf("---response(%s)---OldValue(%s)--origin(%s)-", string(bd), oldValue, hres.Origin))
		*rr.Value = realip
		UpdateAliDNS(rr)
	}
}
