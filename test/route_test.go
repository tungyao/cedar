package test

import (
	"../../cedar"
	"testing"
)

func TestR(t *testing.T) {
	r := cedar.NewRestRouter(cedar.RestConfig{
		"wechat",
		"api",
	})
	r.GetR()
}
