package controller

import (
	"reflect"
	"strings"
)

func compareConf(srcConf map[string]string, dstConf map[string]string) bool {
	return reflect.DeepEqual(srcConf, dstConf)
}

func map2String(kv map[string]string) string {
	var sb strings.Builder
	for key, value := range kv {
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(value)
		sb.WriteString("\n")
	}
	return sb.String()
}

func int32Ptr(i int32) *int32 { return &i }
