package schema

import (
	"fmt"
	"reflect"
)

func (schema *Schema) Getter(value reflect.Value, keys ...any) any {
	if !value.IsValid() {
		return nil
	}
	if len(keys) == 0 {
		return value.Interface()
	}
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Struct:
		if sch, err := Parse(value); err == nil {
			return sch.GetValue(value, keys...)
		} else {
			fmt.Printf("schema.Getter:%v\n", err)
			return nil
		}
	case reflect.Map:
		return schema.GetMapValue(value, keys...)
	case reflect.Slice, reflect.Array:
		return schema.GetArrayValue(value, keys...)
	default:
		fmt.Printf("schema.Getter:子类型不正确无法继续递归\n")
		return nil
	}
}

func (schema *Schema) GetMapValue(value reflect.Value, keys ...any) any {
	if !value.IsValid() || value.IsNil() {
		return nil
	}
	v := value.MapIndex(reflect.ValueOf(keys[0]))
	return schema.Getter(v, keys[1:]...)

}
func (schema *Schema) GetArrayValue(value reflect.Value, keys ...any) any {
	n := len(keys)
	if n == 0 || !value.IsValid() {
		return nil
	}

	i := int(ToInt(keys[0]))
	if i >= value.Len() {
		return nil
	}
	v := value.Index(i)
	return schema.Getter(v, keys[1:]...)

}
