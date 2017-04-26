package envar

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

func valueStore(v reflect.Value, envVal string) error {
	if !v.CanSet() {
		return fmt.Errorf("The value %+v is not settable", v)
	}

	switch v.Kind() {
	case reflect.Bool:
		switch envVal {
		case "true", "True", "TRUE":
			v.SetBool(true)
		case "false", "False", "FALSE":
			v.SetBool(false)
		default:
			return fmt.Errorf("Unknown boolean value: %s", envVal)
		}
	case reflect.String:
		v.SetString(envVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(envVal, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	default:
		return fmt.Errorf("Unknown type: %+v", v.Kind())
	}
	return nil
}

func Parse(config interface{}) error {
	v := reflect.ValueOf(config)
	elem := v.Elem()

	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("Element type: %s is not supported", elem.Kind())
	}

	for i := 0; i < elem.Type().NumField(); i++ {
		f := elem.Type().Field(i)
		fVal := elem.FieldByName(f.Name)
		if !fVal.CanSet() {
			continue
		}

		envKey := f.Tag.Get("envar")
		if envKey == "" {
			envKey = f.Name
		}
		envVal := os.Getenv(envKey)
		if envVal == "" {
			continue
		}

		switch fVal.Kind() {
		case reflect.Slice:
			offset := 0
			envSlice := make([]string, 0, 10)
			for i, c := range []byte(envVal) {
				if c == ',' {
					envSlice = append(envSlice, envVal[offset:i])
					offset = i + 1
				}
			}
			envSlice = append(envSlice, envVal[offset:])

			v := reflect.MakeSlice(fVal.Type(), len(envSlice), len(envSlice))
			for i, s := range envSlice {
				err := valueStore(v.Index(i), s)
				if err != nil {
					return err
				}
			}
			fVal.Set(v)
		default:
			err := valueStore(fVal, envVal)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
