package schema

import (
	"fmt"
	"go/ast"
	"reflect"
)

// Parse get data type from dialector
func Parse(dest interface{}) (*Schema, error) {
	return ParseWithSpecialTableName(dest, "", config)
}
func GetOrParse(dest interface{}, opts *Options) (*Schema, error) {
	return ParseWithSpecialTableName(dest, "", opts)
}

// ParseWithSpecialTableName get data type from dialector with extra schema table
func ParseWithSpecialTableName(dest interface{}, specialTableName string, opts *Options) (*Schema, error) {
	if dest == nil {
		return nil, fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
	}
	modelType := Kind(dest)
	if modelType.Kind() != reflect.Struct {
		if modelType.PkgPath() == "" {
			return nil, fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
		}
		return nil, fmt.Errorf("%w: %s.%s", ErrUnsupportedDataType, modelType.PkgPath(), modelType.Name())
	}

	// Cache the schema for performance,
	// Use the modelType or modelType + schemaTable (if it present) as cache key.
	var schemaCacheKey interface{}
	if specialTableName != "" {
		schemaCacheKey = fmt.Sprintf("%p-%s", modelType, specialTableName)
	} else {
		schemaCacheKey = modelType
	}

	schema := &Schema{
		init:    make(chan struct{}),
		options: opts,
	}
	// Player exist schmema cache, return if exists
	if v, loaded := opts.Store.LoadOrStore(schemaCacheKey, schema); loaded {
		s := v.(*Schema)
		if s.init != nil {
			<-s.init
		}
		return s, s.err
	} else {
		defer schema.initialized()
	}
	defer func() {
		if schema.err != nil {
			opts.Store.Delete(modelType)
		}
	}()

	modelValue := reflect.New(modelType)
	var tableName string
	if tabler, ok := modelValue.Interface().(Tabler); ok {
		tableName = tabler.TableName()
	} else {
		tableName = opts.TableName(modelType.Name())
	}
	if specialTableName != "" && specialTableName != tableName {
		tableName = specialTableName
	}

	schema.Name = modelType.Name()
	schema.ModelType = modelType
	schema.Table = tableName
	schema.FieldsByName = map[string]*Field{}
	schema.FieldsByDBName = map[string]*Field{}

	//var embeddedField []*Field

	for i := 0; i < modelType.NumField(); i++ {
		fieldStruct := modelType.Field(i)
		if ast.IsExported(fieldStruct.Name) {
			field := schema.ParseField(fieldStruct)

			if field.StructField.Anonymous {
				schema.Embedded = append(schema.Embedded, field)
				//if field.Embedded != nil {
				//	embeddedField = append(embeddedField, field)
				//}
			} else {
				schema.Fields = append(schema.Fields, field)
				schema.FieldsByName[field.Name] = field
			}
		}
		if schema.err != nil {
			return nil, schema.err
		}
	}
	//所有Anonymous字段
	//ignore := map[string]*Field{} //忽略字段,在多个继承对象中出现
	//anonymous := map[string]*Field{}

	for _, v := range schema.Embedded {
		for _, field := range v.GetEmbeddedFields() {
			if _, ok := schema.FieldsByName[field.Name]; !ok {
				schema.FieldsByName[field.Name] = field
			}
			//if _, ok := ignore[field.Name]; !ok {
			//	if _, ok2 := anonymous[field.Name]; !ok2 {
			//		anonymous[field.Name] = field
			//	} else {
			//		ignore[field.Name] = field
			//		delete(anonymous, field.Name)
			//	}
			//}
		}
	}

	//for _, field := range anonymous {
	//	if _, ok := schema.FieldsByName[field.Name]; !ok {
	//		schema.FieldsByName[field.Name] = field
	//	}
	//
	//}

	for _, field := range schema.FieldsByName {
		if field.DBName != "" {
			if f, ok := schema.FieldsByDBName[field.DBName]; !ok {
				schema.FieldsByDBName[field.DBName] = field
			} else {
				return nil, fmt.Errorf("struct(%v) DBName repeat: %v,%v", schema.Name, f.Name, field.Name)
			}
		}
	}

	return schema, schema.err
}
