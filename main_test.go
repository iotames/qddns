package main

import (
	"testing"
)

func TestTryGetRealIP(t *testing.T) {
	getip := TryGetRealIP()
	t.Logf("-----TryGetRealIP(%s)----\n", getip)
}
