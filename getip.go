package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/iotames/miniutils"
)

func GetRealIpBy1(requrl string) string {
	// https://ip.3322.net
	logger := miniutils.GetLogger("")
	return GetRealIpByUrl(requrl, func(b []byte) string {
		resp := string(b)
		logger.Info(fmt.Sprintf("----TryGetRealIP(%s)--resp(%s)---", requrl, resp))
		result := strings.TrimSpace(resp)
		resplit := strings.Split(result, `.`)
		if len(resplit) != 4 {
			result = ""
		}
		// logger.Debug(fmt.Sprintf("----TryGetRealIP(%s)--Result(%s)----", requrl, result))
		return result
	})
}

func GetRealIpBy2(requrl string) string {
	// {"rs":1,"code":0,"address":"中国  福建省 泉州市 电信","ip":"117.24.81.51","isDomain":0}
	// requrl = "https://www.ip.cn/api/index?ip&type=0"
	logger := miniutils.GetLogger("")
	return GetRealIpByUrl(requrl, func(b []byte) string {
		btext := string(b)
		logger.Info(fmt.Sprintf("----TryGetRealIP(%s)--resp(%s)---", requrl, btext))
		vv, err := GetRealIpJson(b, func(dt map[string]interface{}) string {
			var v string
			val, ok := dt["ip"]
			if ok {
				v = val.(string)
			}
			return v
		})
		if err != nil {
			logger.Warn(fmt.Sprintf("----TryGetRealIP(%s)--json.Unmarshal--err(%v)--bd(%s)--", requrl, err, btext))
		}
		// logger.Debug(fmt.Sprintf("----TryGetRealIP(%s)--Result(%s)----", requrl, vv))
		return vv
	})
}

func GetRealIpBy3(requrl string) string {
	// {"code":200,"msg":"success","data":{"address":"中国 福建 泉州 电信","ip":"117.24.81.51"}}
	// requrl = "https://searchplugin.csdn.net/api/v1/ip/get"
	logger := miniutils.GetLogger("")
	return GetRealIpByUrl(requrl, func(b []byte) string {
		btext := string(b)
		logger.Info(fmt.Sprintf("----TryGetRealIP(%s)--resp(%s)---", requrl, btext))
		vv, err := GetRealIpJson(b, func(dt map[string]interface{}) string {
			var v string
			val, ok := dt["code"]
			if ok {
				code := val.(float64)
				if code == 200 {
					val, ok = dt["data"]
					if ok {
						data := val.(map[string]interface{})
						val, ok = data["ip"]
						if ok {
							v = val.(string)
						}
					}
				}
			}
			return v
		})
		if err != nil {
			logger.Warn(fmt.Sprintf("----TryGetRealIP(%s)--json.Unmarshal--err(%v)--bd(%s)--", requrl, err, btext))
		}
		// logger.Debug(fmt.Sprintf("----TryGetRealIP(%s)--Result(%s)----", requrl, vv))
		return vv
	})
}

// TryGetRealIP 通过多种方式尝试获取本机IP
// https://blog.csdn.net/qq_43762932/article/details/129408876
func TryGetRealIP() string {
	var requrl string
	var ip string
	requrl = "https://ip.3322.net"
	ip = GetRealIpBy1(requrl)

	// if ip == "" {
	// 	// {"rs":1,"code":0,"address":"中国  福建省 泉州市 电信","ip":"117.24.81.51","isDomain":0}
	// 	requrl = "https://www.ip.cn/api/index?ip&type=0" 不准
	// 	ip = GetRealIpBy2(requrl)
	// }
	if ip == "" {
		// {"ret":"ok","ip":"117.24.81.51","data":["中国","福建","泉州","鲤城","电信","362000","0595"]}
		requrl = "https://2023.ipchaxun.com"
		ip = GetRealIpBy2(requrl)
	}

	if ip == "" {
		// {"code":200,"msg":"success","data":{"address":"中国 福建 泉州 电信","ip":"117.24.81.51"}}
		requrl = "https://searchplugin.csdn.net/api/v1/ip/get"
		ip = GetRealIpBy3(requrl)
	}
	return ip
}

func GetRealIpJson(b []byte, h func(dt map[string]interface{}) string) (result string, err error) {
	hres := make(map[string]interface{})
	err = json.Unmarshal(b, &hres)
	if err != nil {
		return
	}
	result = h(hres)
	return
}

func GetRealIpByUrl(reqUrl string, getip func(b []byte) string) string {
	logger := miniutils.GetLogger("")
	c := http.DefaultClient
	hreq, _ := http.NewRequest("GET", reqUrl, nil)
	hreq.Header.Set("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36`)
	hreq.Header.Set("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7`)
	resp, err := c.Do(hreq)
	if err != nil {
		logger.Error(fmt.Sprintf("---GetRealIpByUrl--c.Do--err(%v)---", err))
		return ""
	}
	defer resp.Body.Close()
	bd, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn(fmt.Sprintf("----GetRealIpByUrl--io.ReadAll--err(%v)---", err))
		return ""
	}
	return getip(bd)
}
