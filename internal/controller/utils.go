package controller

import (
	"reflect"
	"strings"

	"github.com/go-xmlfmt/xmlfmt"
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

func map2Xml(properties map[string]string) string {
	var res string
	for key, value := range properties {
		property := `<property>
	<name>` + key + `</name>
	<value>` + value + `</value>
</property>`
		res = res + property
	}

	res = xmlfmt.FormatXML(res, "", "  ")

	return res
}

func int32Ptr(i int32) *int32 { return &i }
