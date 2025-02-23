package msstrings

import (
	"fmt"
	"reflect"
	"strings"
)

// JoinStrings 将多个字符串或其他类型的值连接成一个字符串。
// 它接受可变数量的参数，参数类型可以是任意类型。
// 该函数使用 strings.Builder 来高效地构建最终的字符串。
func JoinStrings(data ...any) string {
	var sb strings.Builder
	for _, v := range data {
		sb.WriteString(check(v))
	}
	return sb.String()
}

// check 检查并转换传入的值为字符串。
// 如果值是字符串类型，则直接返回。
// 否则，将值格式化为字符串并返回。
func check(v any) string {
	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.String:
		return v.(string)
	default:
		return fmt.Sprintf("%v", v)
	}
}
