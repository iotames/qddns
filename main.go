package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	sipInfo = &ServerIpInfo{}
	CheckAliDNS(handleAliDNS)
	for range time.Tick(time.Minute * time.Duration(CheckTTL)) {
		log.Printf("---Begin---Check----ServerIpInfo(%+v)--", sipInfo)
		CheckAliDNS(handleAliDNS)
	}
}

func init() {
	err := godotenv.Load(".env", "env.default")
	if err != nil {
		panic("godotenv Error: " + err.Error())
	}
	LoadEnvArgs()
}
