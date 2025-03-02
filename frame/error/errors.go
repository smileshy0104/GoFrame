// Mserror包用于定义错误处理的结构体和方法
package error

// MsError是错误处理的结构体，包含一个error类型的err和一个错误处理函数ErrFuc
type MsError struct {
	err    error
	ErrFuc ErrorFuc
}

// Default方法返回MsError结构体的默认实例
func Default() *MsError {
	return &MsError{}
}

// Error方法实现了error接口，返回当前错误的字符串表示
func (e *MsError) Error() string {
	return e.err.Error()
}

// Put方法用于记录错误，通过调用内部的check方法进行错误处理
func (e *MsError) Put(err error) {
	e.check(err)
}

// check方法用于检查错误，如果错误不为空，则将错误赋值给err字段并抛出panic
func (e *MsError) check(err error) {
	if err != nil {
		e.err = err
		panic(e)
	}
}

// ErrorFuc定义了错误处理函数的类型，它接受一个指向MsError实例的指针
type ErrorFuc func(msError *MsError)

// Result方法用于设置错误处理函数ErrFuc
func (e *MsError) Result(errFuc ErrorFuc) {
	e.ErrFuc = errFuc
}

// ExecResult方法用于执行错误处理函数ErrFuc，它接受当前的MsError实例作为参数
func (e *MsError) ExecResult() {
	e.ErrFuc(e)
}
