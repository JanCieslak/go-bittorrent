package bencode

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
)

func Encode(v interface{}) string {
	switch value := v.(type) {
	case string:
		return fmt.Sprintf("%d:%s", len(value), value)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("i%de", value)
	case []interface{}:
		buf := new(bytes.Buffer)
		for _, item := range value {
			buf.WriteString(Encode(item))
		}
		return fmt.Sprintf("l%se", buf.String())
	case map[string]interface{}:
		buf := new(bytes.Buffer)
		inSortedOrder(value, func(k string, v interface{}) {
			_, _ = buf.WriteString(Encode(k))
			_, _ = buf.WriteString(Encode(v))
		})
		return fmt.Sprintf("d%se", buf.String())
	default:
		log.Println("unsupported case of", value)
		os.Exit(1)
		return ""
	}
}

func inSortedOrder(m map[string]interface{}, fn func(k string, v interface{})) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fn(key, m[key])
	}
}
