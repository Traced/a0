package utils

import (
	"io/ioutil"
	"log"

	jsoniter "github.com/json-iterator/go"
)

func JsonBytes(v interface{}) []byte {
	r, _ := jsoniter.Marshal(v)
	return r
}

func JsonString(v interface{}) string {
	return BytesToString(JsonBytes(v))
}

func Marshal(v interface{}) []byte {
	return MarshalIndentBytes(v, "", "    ")
}

func MarshalString(v interface{}) string {
	return MarshalIndentString(v, "", "    ")
}

func MarshalIndentBytes(v interface{}, prefix, indent string) []byte {
	b, _ := jsoniter.MarshalIndent(v, prefix, indent)
	return b
}

func MarshalIndentString(v interface{}, prefix, indent string) string {
	return BytesToString(MarshalIndentBytes(v, prefix, indent))
}

func DumpToFile(fn string, data interface{}) {
	if err := ioutil.WriteFile(fn, Marshal(data), 0755); err != nil {
		log.Println(err)
	}
}
