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

var DnsIpUpdatedAt, RealIpUpdatedAt time.Time
var RealIp string

func CheckAliDNS(h func(rr alidns20150109.DescribeDomainRecordsResponseBodyDomainRecordsRecord)) {
	logger := miniutils.GetLogger("")
	if time.Since(DnsIpUpdatedAt) < time.Hour*12 {
		realIP := GetRealIP()
		if realIP == "" {
			realIP = GetRealIP()
		}
		if realIP != "" {
			canNext := false
			for _, v := range IpInfoMap {
				logger.Debugf("----CheckAliDNS--DnsIpUpdatedAt--In(12h)--realIP(%s)-vs-DnsIp(%s)--", realIP, v.DnsIp)
				if v.DnsIp != realIP {
					canNext = true
				}
			}
			if !canNext {
				logger.Infof("---Skip--CheckAliDNS--DnsIpUpdatedAt--In(12h)--realIP(%s) == DnsIp--", realIP)
				return
			}
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
	logger.Debugf("---ApiResponse--statusCode(%d)--Total(%d)--PageNumber(%d)--PageSize(%d)--", statusCode, *bd.TotalCount, *bd.PageNumber, *bd.PageSize)

	/**
	  2023/09/05 22:20:02 ---CheckAliDNS--Before--handleAliDNS--SUB_DOMAINS([openwrt])--rr({
	     "DomainName": "catmes.com",
	     "Line": "default",
	     "Locked": false,
	     "RR": "openwrt",
	     "RecordId": "830067725280195584",
	     "Status": "ENABLE",
	     "TTL": 600,
	     "Type": "A",
	     "Value": "117.26.40.37",
	     "Weight": 1
	  })---
	  **/
	rrs := bd.DomainRecords.Record
	for _, r := range rrs {
		logger.Debugf("---CheckAliDNS--Before--handleAliDNS--SUB_DOMAINS(%+v)--record(%+v)---", SubDomains, *r)
		h(*r)
	}
}

func UpdateAliDNS(rr alidns20150109.DescribeDomainRecordsResponseBodyDomainRecordsRecord) int {
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
		return int(statusCode)
	}
	logger.Infof("---Success--UpdateAliDNS--Domain(%s.%s)->Value(%s)--", *rr.RR, *rr.DomainName, *rr.Value)
	return int(statusCode)
}

type ServerIpInfo struct {
	DnsIpUpdatedAt  time.Time
	DnsIp           string
	RealIpUpdatedAt time.Time
	RealIp          string
}

var IpInfoMap map[string]*ServerIpInfo

func GetRealIP() string {
	if time.Since(RealIpUpdatedAt) < time.Second*60 && RealIp != "" {
		fmt.Printf("---UseCache---RealIP(%s)----\n", RealIp)
		return RealIp
	}
	logger := miniutils.GetLogger("")
	RealIpUpdatedAt = time.Now()
	RealIp = TryGetRealIP()
	for _, v := range IpInfoMap {
		v.RealIp = RealIp
		v.RealIpUpdatedAt = RealIpUpdatedAt
	}
	logger.Debugf("-----GetRealIP--SUCCESS(%s)---", RealIp)
	return RealIp
}

func handleAliDNS(rr alidns20150109.DescribeDomainRecordsResponseBodyDomainRecordsRecord) {
	logger := miniutils.GetLogger("")
	ipinfo, ok := IpInfoMap[*rr.RR]
	if ok {
		realip := GetRealIP()
		if realip == "" {
			realip = GetRealIP()
		}
		DnsIpUpdatedAt = time.Now()
		ipinfo.DnsIpUpdatedAt = DnsIpUpdatedAt
		ipinfo.RealIpUpdatedAt = DnsIpUpdatedAt
		ipinfo.RealIp = realip
		ipinfo.DnsIp = *rr.Value
		if realip == "" {
			logger.Warn("---handleAliDNS--Skip--UpdateAliDNS-realIP is empty-----")
			return
		}
		if ipinfo.DnsIp == ipinfo.RealIp {
			logger.Debugf("---handleAliDNS--Skip-UpdateAliDNS-(ipinfo.DnsIp == ipinfo.RealIp)[%s]", ipinfo.DnsIp)
			return
		}
		*rr.Value = GetRealIP()
		if UpdateAliDNS(rr) == 200 {
			ipinfo.DnsIp = *rr.Value
		}
	}
}
