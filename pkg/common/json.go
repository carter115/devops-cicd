package common

import "encoding/json"

func ParseJsonStr(obj interface{}) string {
	bs, _ := json.Marshal(obj)
	return string(bs)
}
