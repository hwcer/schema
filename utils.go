package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Kind(dest interface{}) reflect.Type {
	value := ValueOf(dest)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		value = reflect.New(value.Type().Elem())
	}
	modelType := reflect.Indirect(value).Type()

	if modelType.Kind() == reflect.Interface {
		modelType = reflect.Indirect(reflect.ValueOf(dest)).Elem().Type()
	}

	for modelType.Kind() == reflect.Slice || modelType.Kind() == reflect.Array || modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	return modelType
}

func ValueOf(i interface{}) reflect.Value {
	value, ok := i.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(i)
	}
	return value
}

func ToArray(v interface{}) (r []interface{}) {
	vf := reflect.Indirect(reflect.ValueOf(v))
	if vf.Kind() != reflect.Array && vf.Kind() != reflect.Slice {
		return []interface{}{v}
	}
	for i := 0; i < vf.Len(); i++ {
		r = append(r, vf.Index(i).Interface())
	}
	return
}

func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func ToInt(i any) (v int64) {
	switch d := i.(type) {
	case int:
		v = int64(d)
	case uint:
		v = int64(d)
	case int8:
		v = int64(d)
	case uint8:
		v = int64(d)
	case int16:
		v = int64(d)
	case uint16:
		v = int64(d)
	case int32:
		v = int64(d)
	case uint32:
		v = int64(d)
	case int64:
		v = d
	case uint64:
		v = int64(d)
	case float32:
		v = int64(d)
	case float64:
		v = int64(d)
	case string:
		v, _ = strconv.ParseInt(d, 10, 64)
	}
	return
}

func ParseTagSetting(str string, sep string) map[string]string {
	settings := map[string]string{}
	names := strings.Split(str, sep)

	for i := 0; i < len(names); i++ {
		j := i
		if len(names[j]) > 0 {
			for {
				if names[j][len(names[j])-1] == '\\' {
					i++
					names[j] = names[j][0:len(names[j])-1] + sep + names[i]
					names[i] = ""
				} else {
					break
				}
			}
		}

		values := strings.Split(names[j], ":")
		k := strings.TrimSpace(strings.ToUpper(values[0]))

		if len(values) >= 2 {
			settings[k] = strings.Join(values[1:], ":")
		} else if k != "" {
			settings[k] = k
		}
	}

	return settings
}
