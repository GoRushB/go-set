package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func AsString(item interface{}) string {
	switch item.(type) {
	case byte:
		return string(item.(byte))
	case []byte:
		return string(item.([]byte))
	case []rune:
		return string(item.([]rune))
	case *strings.Builder:
		return item.(*strings.Builder).String()
	case string:
		return item.(string)
	case int:
		return strconv.Itoa(item.(int))
	case int32:
		return strconv.Itoa(int(item.(int32)))
	case int64:
		return strconv.FormatInt(item.(int64), 10)
	case float32:
		return strconv.FormatFloat(item.(float64), 'f', 2, 32)
	case float64:
		return strconv.FormatFloat(item.(float64), 'f', 2, 64)
	case bool:
		return strconv.FormatBool(item.(bool))
	default:
		if stringer, ok := item.(fmt.Stringer); ok {
			return stringer.String()
		}
		return fmt.Sprint(item)
	}
}

func AsInt(item interface{}) int {
	switch item.(type) {
	case byte:
		return int(item.(byte))
	case string:
		v, _ := strconv.Atoi(item.(string))
		return v
	case int:
		return item.(int)
	case int32:
		return int(item.(int32))
	case int64:
		return int(item.(int64))
	case float32:
		return int(math.Floor(float64(item.(float32))))
	case float64:
		return int(math.Floor(item.(float64)))
	case bool:
		if item.(bool) {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func AsInt64(item interface{}) int64 {
	switch item.(type) {
	case byte:
		return int64(item.(byte))
	case string:
		v, _ := strconv.ParseInt(item.(string), 10, 64)
		return v
	case int:
		return int64(item.(int))
	case int32:
		return int64(item.(int32))
	case int64:
		return item.(int64)
	case float32:
		return int64(math.Floor(float64(item.(float32))))
	case float64:
		return int64(math.Floor(item.(float64)))
	case bool:
		if item.(bool) {
			return 1
		}
		return 0
	default:
		return 0
	}
}
