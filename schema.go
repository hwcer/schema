package schema

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrUnsupportedDataType unsupported data type
var ErrUnsupportedDataType = errors.New("unsupported data type")

type Schema struct {
	err            error
	init           chan struct{}
	options        *Options
	Name           string
	Table          string
	Fields         []*Field
	Embedded       []*Field //嵌入字段
	ModelType      reflect.Type
	FieldsByName   map[string]*Field
	FieldsByDBName map[string]*Field
}

func (schema *Schema) initialized() {
	close(schema.init)
	schema.init = nil
}

func (schema *Schema) String() string {
	if schema.ModelType.Name() == "" {
		return fmt.Sprintf("%s(%s)", schema.Name, schema.Table)
	}
	return fmt.Sprintf("%s.%s", schema.ModelType.PkgPath(), schema.ModelType.Name())
}

func (schema *Schema) New() reflect.Value {
	results := reflect.New(schema.ModelType)
	return results
}

// Make Make a Slice
func (schema *Schema) Make() reflect.Value {
	slice := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(schema.ModelType)), 0, 20)
	results := reflect.New(slice.Type())
	results.Elem().Set(slice)
	return results
}

func (schema *Schema) LookUpField(name string) *Field {
	if field, ok := schema.FieldsByDBName[name]; ok {
		return field
	}
	if field, ok := schema.FieldsByName[name]; ok {
		return field
	}
	return nil
}

// ReflectValueOf 获取i中的一个字段
// key 字段名,可以使用(a.b.c)模式递归查找
//
//	func (schema *Schema) ReflectValueOf(v reflect.Value, key string) (r reflect.Value, field *Field) {
//		index := strings.Index(key, ".")
//		if index < 0 {
//			if field = schema.LookUpField(key); field != nil {
//				r = field.ReflectValueOf(v)
//			}
//			return
//		}
//		if field = schema.LookUpField(key[0:index]); field == nil {
//			return
//		}
//
//		fmt.Printf("================================\nFieldName:%v\n", field.Name)
//		fmt.Printf("FieldType:%v\n", field.FieldType.Kind())
//		fmt.Printf("IndirectFieldType:%v\n", field.IndirectFieldType.Kind())
//		switch field.IndirectFieldType.Kind() {
//		case reflect.Struct:
//			r = field.Get(v)
//			return field.Embedded.ReflectValueOf(r, key[index+1:])
//		default:
//			field = nil
//			fmt.Printf("子类型不正确无法继续递归\n")
//			return
//		}
//
//		return
//	}

func (schema *Schema) GetValue(obj any, keys ...any) (r any) {
	vf := ValueOf(obj)
	n := len(keys)
	if n == 0 {
		return
	}
	k := ToString(keys[0])
	field := schema.LookUpField(k)
	if field == nil {
		return
	}
	v := field.Get(vf)
	return schema.Getter(v, keys[1:]...)
}

func (schema *Schema) SetValue(obj any, val any, keys ...any) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	vf := ValueOf(obj)
	var sch *Schema
	var field *Field
	l := len(keys)
	n := l - 1
	for i := 0; i < l; i++ {
		sk := ToString(keys[i])
		if field != nil {
			sch = field.Embedded
		} else {
			sch = schema
		}
		field = sch.LookUpField(sk)
		if field == nil {
			return fmt.Errorf("field not exist:%v", sk)
		}
		if i < n {
			if field.Embedded == nil {
				return fmt.Errorf("field not object:%v", field.Name)
			}
			vf = field.Get(vf)

			//if v := field.Get(vf); v.IsZero() {
			//	v = reflect.New(field.FieldType)
			//	if err = field.Set(vf, v.Interface()); err != nil {
			//		return
			//	}
			//	vf = v
			//} else {
			//	vf = v
			//}

		}
	}
	return field.Set(vf, val)
	//if len(keys) == 1 {
	//	k := ToString(keys[0])
	//	field := schema.LookUpField(k)
	//	if field == nil {
	//		return fmt.Errorf("field not exist:%v", k)
	//	}
	//	return field.Set(ValueOf(obj), val)
	//}
	//b := schema.convert(val, keys...)
	//logger.Trace("convert:%v", string(b))
	//err = json.Unmarshal(b, obj)
	return
}

//
//// a.b.c = v 转换成{"a":{"b":{"c":v}}}
//func (schema *Schema) convert(val any, keys ...any) []byte {
//	data := map[string]any{}
//	var node map[string]any
//	n := len(keys) - 1
//	for i, k := range keys {
//		sk := ToString(k)
//		if i == n {
//			node[sk] = val
//		} else {
//			node = map[string]any{}
//			data[sk] = node
//		}
//	}
//	b, _ := json.Marshal(data)
//	return b
//
//}

func (schema *Schema) ParseField(fieldStruct reflect.StructField) *Field {
	//var err error

	field := &Field{
		Name:              fieldStruct.Name,
		FieldType:         fieldStruct.Type,
		IndirectFieldType: fieldStruct.Type,
		StructField:       fieldStruct,
		Schema:            schema,
	}

	for field.IndirectFieldType.Kind() == reflect.Ptr {
		field.IndirectFieldType = field.IndirectFieldType.Elem()
	}

	fieldValue := reflect.New(field.IndirectFieldType)
	if dbName := fieldStruct.Tag.Get("bson"); dbName != "" && dbName != "inline" {
		field.DBName = dbName
	} else {
		field.DBName = schema.options.ColumnName(schema.Table, field.Name)
	}
	field.Index = field.StructField.Index
	kind := reflect.Indirect(fieldValue).Kind()
	switch kind {
	case reflect.Struct:
		field.Embedded, schema.err = GetOrParse(fieldValue.Interface(), schema.options)
	case reflect.Map, reflect.Slice, reflect.Array:
		//初始化子结构
	case reflect.Invalid, reflect.Uintptr, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
		schema.err = fmt.Errorf("invalid embedded struct for %s's field %s, should be struct, but got %v", field.Schema.Name, field.Name, field.FieldType)
	}

	//if fieldStruct.Anonymous {
	//
	//}
	return field
}
