package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	IpInfoMap = make(map[string]*ServerIpInfo, len(SubDomains))
	for _, v := range SubDomains {
		IpInfoMap[v] = new(ServerIpInfo) // &ServerIpInfo{}
	}
	CheckAliDNS(handleAliDNS)
	log.Printf("---End---CheckAliDNS-----\n")
	for range time.Tick(time.Minute * time.Duration(CheckTTL)) {
		log.Printf("---Begin---CheckAliDNS----IpInfoMap(%+v)--\n", IpInfoMap)
		CheckAliDNS(handleAliDNS)
		log.Printf("---End---CheckAliDNS-----\n")
	}
}

func init() {
	err := godotenv.Load(".env", "env.default")
	if err != nil {
		panic("godotenv Error: " + err.Error())
	}
	LoadEnvArgs()
}
