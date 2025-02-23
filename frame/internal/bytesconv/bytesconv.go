package bytesconv

import "unsafe"

// StringToBytes 将字符串转换为字节切片，而不进行内存复制。
// 这个函数利用了unsafe包来避免字符串到字节切片的常规转换中涉及的内存分配和复制，
// 从而提高了性能。但是，使用这种方法时需要谨慎，因为生成的字节切片共享字符串的内存，
// 对其进行修改可能会导致意外的副作用。
// 参数:
//
//	s - 需要转换的字符串。
//
// 返回值:
//
//	字节切片，与输入字符串共享内存。
func StringToBytes(s string) []byte {
	// 通过unsafe.Pointer直接操作内存，将字符串转换为字节切片。
	// 这个转换是通过将字符串和其长度包装在一个结构体中，然后将该结构体的内存地址
	// 解释为一个指向字节切片的指针来实现的。
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}
