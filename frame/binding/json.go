package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

// jsonBinding 结构体定义了JSON绑定的属性和行为。
// 它允许配置是否允许未知字段以及是否在绑定时进行验证。
type jsonBinding struct {
	DisallowUnknownFields bool
	IsValidate            bool
}

// Name 方法返回绑定的名称，这里是"json"。
func (jsonBinding) Name() string {
	return "json"
}

// Bind 方法处理HTTP请求的JSON数据绑定到指定的对象。
// 它根据配置处理未知字段和参数验证。
func (b jsonBinding) Bind(r *http.Request, obj any) error {
	body := r.Body
	//post传参的内容 是放在 body中的
	if body == nil {
		return errors.New("invalid request")
	}
	// 创建一个JSON解码器，并设置是否允许未知字段。
	decoder := json.NewDecoder(body)
	// 如果不允许未知字段，则设置不允许未知字段。
	if b.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	// 如果启用了参数验证，则调用validateParam函数对解码后的数据进行验证。
	if b.IsValidate {
		// 调用validateParam函数对解码后的数据进行验证。
		err := validateParam(obj, decoder)
		if err != nil {
			return err
		}
	} else {
		// 如果不启用参数验证，则直接解码JSON数据到obj对象中。
		err := decoder.Decode(obj)
		if err != nil {
			return err
		}
	}
	return validate(obj)
}

// validateParam 函数根据对象的反射类型进行参数验证。
// 它支持指针、结构体、切片和数组类型的参数验证。
func validateParam(obj any, decoder *json.Decoder) error {
	// 获取对象的反射值，并检查是否是指针类型。
	valueOf := reflect.ValueOf(obj)
	// 检查对象的反射值是否是指针类型，如果不是，则返回错误。
	if valueOf.Kind() != reflect.Pointer {
		return errors.New("This argument must have a pointer type")
	}
	// 获取对象的反射值，并检查是否是指针类型。
	elem := valueOf.Elem().Interface()
	of := reflect.ValueOf(elem)
	// 根据对象的反射值类型进行不同的处理。
	switch of.Kind() {
	case reflect.Struct:
		return checkParam(of, obj, decoder)
	case reflect.Slice, reflect.Array:
		elem := of.Type().Elem()
		if elem.Kind() == reflect.Struct {
			return checkParamSlice(elem, obj, decoder)
		}
	default:
		_ = decoder.Decode(obj)
	}
	return nil
}

// checkParamSlice 函数用于验证切片类型的参数。
// 它解码JSON数据为map切片，然后检查每个结构体字段的必需性。
func checkParamSlice(of reflect.Type, obj any, decoder *json.Decoder) error {
	mapValue := make([]map[string]interface{}, 0)
	_ = decoder.Decode(&mapValue)
	for i := 0; i < of.NumField(); i++ {
		field := of.Field(i)
		name := field.Name
		jsonName := field.Tag.Get("json")
		if jsonName != "" {
			name = jsonName
		}
		required := field.Tag.Get("binding")
		for _, v := range mapValue {
			value := v[name]
			if value == nil && required == "required" {
				return errors.New(fmt.Sprintf("filed [%s] is not exist,because [%s] is required", jsonName, jsonName))
			}
		}
	}
	b, _ := json.Marshal(mapValue)
	_ = json.Unmarshal(b, obj)
	return nil
}

// checkParam 函数用于验证结构体类型的参数。
// 它解码JSON数据为map，然后检查每个字段的必需性。
func checkParam(of reflect.Value, obj any, decoder *json.Decoder) error {
	mapValue := make(map[string]interface{})
	_ = decoder.Decode(&mapValue)
	for i := 0; i < of.NumField(); i++ {
		field := of.Type().Field(i)
		name := field.Name
		jsonName := field.Tag.Get("json")
		if jsonName != "" {
			name = jsonName
		}
		required := field.Tag.Get("binding")
		value := mapValue[name]
		if value == nil && required == "required" {
			return errors.New(fmt.Sprintf("filed [%s] is not exist,because [%s] is required", jsonName, jsonName))
		}
	}
	b, _ := json.Marshal(mapValue)
	_ = json.Unmarshal(b, obj)
	return nil
}
