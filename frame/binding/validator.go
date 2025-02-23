package binding

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
	"sync"
)

// StructValidator 是一个接口，定义了验证结构体和获取验证器引擎的方法
type StructValidator interface {
	// ValidateStruct 验证给定的结构体，如果验证失败则返回错误
	ValidateStruct(any) error
	// Engine 返回当前使用的验证器引擎实例
	Engine() any
}

// Validator 是 StructValidator 接口的全局实现，使用默认验证器
var Validator StructValidator = &defaultValidator{}

// defaultValidator 是 StructValidator 接口的具体实现，使用 sync.Once 确保验证器只被初始化一次
type defaultValidator struct {
	one      sync.Once
	validate *validator.Validate
}

// SliceValidationError 代表一个错误切片，用于存储多个验证错误
type SliceValidationError []error

// Error 实现了 error 接口，以友好的格式输出所有错误信息
func (err SliceValidationError) Error() string {
	// 根据错误数量构建并返回错误信息字符串
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
				}
			}
		}
		return b.String()
	}
}

// ValidateStruct 验证给定的对象，支持指针、结构体、切片和数组类型
func (d *defaultValidator) ValidateStruct(obj any) error {
	of := reflect.ValueOf(obj)
	switch of.Kind() {
	case reflect.Pointer:
		// 如果是指针类型，获取其指向的值并进行验证
		return d.ValidateStruct(of.Elem().Interface())
	case reflect.Struct:
		// 如果是结构体类型，调用 validateStruct 进行验证
		return d.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		// 如果是切片或数组类型，遍历每个元素进行验证
		count := of.Len()
		sliceValidationError := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := d.validateStruct(of.Index(i).Interface()); err != nil {
				sliceValidationError = append(sliceValidationError, err)
			}
		}
		// 如果有验证错误，返回包含所有错误的 SliceValidationError
		if len(sliceValidationError) == 0 {
			return nil
		}
		return sliceValidationError
	}
	// 对于不支持的类型，返回 nil 表示无错误
	return nil
}

// Engine 返回当前使用的验证器引擎实例
func (d *defaultValidator) Engine() any {
	d.lazyInit()
	return d.validate
}

// lazyInit 懒惰初始化验证器引擎，确保只被初始化一次
func (d *defaultValidator) lazyInit() {
	d.one.Do(func() {
		d.validate = validator.New()
	})
}

// validateStruct 验证给定的结构体，实际调用 validator.Validate 的 Struct 方法
func (d *defaultValidator) validateStruct(obj any) error {
	d.lazyInit()
	return d.validate.Struct(obj)
}

// validate 是 ValidateStruct 方法的包装，提供更简单的调用接口
func validate(obj any) error {
	// 调用全局的 Validator 进行验证
	return Validator.ValidateStruct(obj)
}
