package main

import (
	"fmt"
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
	if time.Since(sipInfo.DnsIpUpdatedAt) < time.Hour*48 {
		realIP := GetRealIP()
		if realIP == "" {
			realIP = GetRealIP()
		}
		if realIP != "" && realIP == sipInfo.DnsIp {
			logger.Info("---Skip--CheckAliDNS--sipInfo.DnsIpUpdatedAt---In(48h)(realIP == sipInfo.DnsIp)--")
			return
		}
	}
	describeDomainRecordsRequest := &alidns20150109.DescribeDomainRecordsRequest{DomainName: &DomainName}
	runtime := &util.RuntimeOptions{}
	client := GetClientAli()
	result, err := client.DescribeDomainRecordsWithOptions(describeDomainRecordsRequest, runtime)
	if err != nil {
		logger.Error(fmt.Sprintf("-----err--CheckAliDNS--client.DescribeDomainRecordsWithOptions--err(%v)", err))
		return
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
	sipInfo.RealIpUpdatedAt = time.Now()
	sipInfo.RealIp = TryGetRealIP()
	logger.Debug(fmt.Sprintf("-----GetRealIP--SUCCESS(%s)---", sipInfo.RealIp))
	return sipInfo.RealIp
}

func handleAliDNS(rr alidns20150109.DescribeDomainRecordsResponseBodyDomainRecordsRecord) {
	logger := miniutils.GetLogger("")
	si := miniutils.GetIndexOf[string](*rr.RR, SubDomains)
	if si > -1 {
		logger.Debug(fmt.Sprintf("---handleAliDNS--Match--SUB_DOMAINS(%+v)--rr(%+v)---", SubDomains, rr))
		oldValue := *rr.Value
		realip := GetRealIP()
		if realip == "" {
			realip = GetRealIP()
		}
		sipInfo.DnsIp = oldValue
		if realip == "" {
			logger.Warn("---handleAliDNS--Skip--UpdateAliDNS-realIP is empty-----")
			return
		}
		if realip == oldValue {
			logger.Debug("---handleAliDNS--Skip--UpdateAliDNS--(realip == oldValue)", oldValue)
			return
		}
		// logger.Warn(fmt.Sprintf("---response(%s)---OldValue(%s)--origin(%s)-", string(bd), oldValue, hres.Origin))
		*rr.Value = realip
		UpdateAliDNS(rr)
	}
}
