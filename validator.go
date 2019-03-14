package jsonunknown

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ValidateUnknownFields checks that provided json
// matches provided struct. If that is not the case
// list of unknown fields is returned.
func ValidateUnknownFields(jsn []byte, strct interface{}) ([]string, error) {
	var obj interface{}
	err := json.Unmarshal(jsn, &obj)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling json: %v", err)
	}
	return checkUnknownFields("", obj, reflect.ValueOf(strct)), nil
}

func checkUnknownFields(keyPref string, jsn interface{}, strct reflect.Value) []string {
	var uf []string
	switch concreteVal := jsn.(type) {
	case map[string]interface{}:
		// Iterate over map and check every value
		for field, val := range concreteVal {
			fullKey := fmt.Sprintf("%s.%s", keyPref, field)
			subStrct := getSubStruct(field, strct)
			if !subStrct.IsValid() {
				uf = append(uf, fullKey[1:])
			} else {
				subUf := checkUnknownFields(fullKey, val, subStrct)
				uf = append(uf, subUf...)
			}
		}
	case []interface{}:
		for i, val := range concreteVal {
			fullKey := fmt.Sprintf("%s[%v]", keyPref, i)
			subStrct := strct.Index(i)
			uf = append(uf, checkUnknownFields(fullKey, val, subStrct)...)
		}
	}
	return uf
}

func getSubStruct(field string, strct reflect.Value) reflect.Value {
	elem := strct
	if strct.Kind() == reflect.Interface || strct.Kind() == reflect.Ptr {
		elem = strct.Elem()
	}
	switch elem.Kind() {
	case reflect.Map:
		iter := elem.MapRange()
		for iter.Next() {
			key := iter.Key().Interface().(string)
			if key == field {
				return iter.Value()
			}
		}
	case reflect.Struct:
		for i := 0; i < elem.NumField(); i++ {
			structField := elem.Type().Field(i)
			// Check if field is embedded struct
			if structField.Anonymous {
				subStruct := getSubStruct(field, elem.Field(i))
				if subStruct.IsValid() {
					return subStruct
				}
			} else {
				fieldName := getJSONTagName(structField, i)
				if fieldName == field {
					return elem.Field(i)
				}
			}
		}
	}
	return reflect.Value{}
}

func getJSONTagName(field reflect.StructField, i int) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		if commaIdx := strings.Index(jsonTag, ","); commaIdx > 0 {
			return jsonTag[:commaIdx]
		}
		return jsonTag
	}
	return ""
}
