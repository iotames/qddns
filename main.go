package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	sipInfo = &ServerIpInfo{}
	CheckAliDNS(handleAliDNS)
	log.Printf("---End---CheckAliDNS-----\n")
	for range time.Tick(time.Minute * time.Duration(CheckTTL)) {
		log.Printf("---Begin---CheckAliDNS----ServerIpInfo(%+v)--\n", sipInfo)
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
