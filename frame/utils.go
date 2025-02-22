package frame

import (
	"strings"
	"unicode"
	"unsafe"
)

// SubStringLast 返回字符串 str 中最后一个出现的子串 substr 后面的部分。
// 如果子串 substr 不在 str 中，则返回空字符串。
func SubStringLast(str string, substr string) string {
	index := strings.Index(str, substr)
	if index < 0 {
		return ""
	}
	return str[index+len(substr):]
}

// isASCII 检查字符串 s 是否都是 ASCII 字符。
// 遍历字符串中的每个字节，如果字节值大于 unicode.MaxASCII，则说明不是 ASCII 字符。
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// StringToBytes 将字符串 s 转换为字节切片，而不进行内存复制。
// 利用 unsafe 包进行类型转换，将字符串转换为字节切片，该操作不安全，需要谨慎使用。
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}
