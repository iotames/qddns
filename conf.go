package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var AliKey string
var AliSecret string
var DomainName string
var SubDomains []string
var CheckTTL int

func LoadEnvArgs() {
	var err error
	AliKey = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	AliSecret = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	DomainName = os.Getenv("DOMAIN_NAME")
	if DomainName == "" {
		panic("DOMAIN_NAME is empty")
	}
	SubDomains = strings.Split(os.Getenv("SUB_DOMAINS"), ",")
	CheckTTL, err = strconv.Atoi(os.Getenv("CHECK_TTL"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("---AliKey(%s)--AliSecret(%s)--DomainName(%s)--SubDomains(%+v)--CheckTTL(%d)--\n", AliKey, AliSecret, DomainName, SubDomains, CheckTTL)
}
