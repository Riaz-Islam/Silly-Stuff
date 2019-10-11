package "lib"

import (
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

// ptr wraps the given value with pointer: V => *V, *V => **V, etc.
func ptr(v reflect.Value) reflect.Value {
	pt := reflect.PtrTo(v.Type()) // create a *T type.
	pv := reflect.New(pt.Elem())  // create a reflect.Value of type *T.
	pv.Elem().Set(v)              // sets pv to point to underlying value of v.
	return pv
}

// A value is addressable if it is
// an element of a slice, an element of an addressable array,
// a field of an addressable struct, or the result of dereferencing a pointer.
func DeepKeepDiff(v1, v2 reflect.Value, depth int) error {
	if depth > 10 {
		return nil
	}
	if depth == 0 && v1.Type() != v2.Type() {
		err := errors.New("type mismatch")
		return err
	}
	if v1.Kind() == reflect.Ptr {
		fmt.Printf("type: %v $$ val1: %v $$ val2 : %v\n", v1.Type(), v1.Elem().Interface(), v2.Elem().Interface())
	} else {
		fmt.Printf("type: %v $$ val1: %v $$ val2 : %v\n", v1.Type(), v1.Interface(), v2.Interface())
	}
	if !v1.IsValid() || !v2.IsValid() {
		return nil
	}
	//if v1.Kind() == reflect.Ptr {
	//	if v1.IsNil() || v2.IsNil() {
	//		return nil
	//	}
	//	v1 = reflect.Indirect(v1)
	//	v2 = reflect.Indirect(v2)
	//}
	if v1.CanInterface() && v2.CanInterface() && reflect.DeepEqual(v1.Interface(), v2.Interface()) {
		if v1.CanSet() && v2.CanSet() {
			v1.Set(reflect.Zero(v1.Type()))
			v2.Set(reflect.Zero(v2.Type()))
			return nil
		}
	}
	switch v1.Kind() {
	case reflect.Array:
		for i := 0; i < v1.Len(); i++ {
			if v1.Index(i).CanAddr() && v2.Index(i).CanAddr() {
				if err := DeepKeepDiff(v1.Index(i).Addr(), v2.Index(i).Addr(), depth+1); err != nil {
					return err
				}
			}
		}
		return nil

	case reflect.Slice:
		if v1.IsNil() || v2.IsNil() {
			return nil
		}
		minlen := v1.Len()
		if v2.Len() < v1.Len() {
			minlen = v2.Len()
		}
		for i := 0; i < minlen; i++ {
			if v1.Index(i).CanAddr() && v2.Index(i).CanAddr() {
				if err := DeepKeepDiff(v1.Index(i).Addr(), v2.Index(i).Addr(), depth+1); err != nil {
					return err
				}
			}
		}
		return nil

	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return nil
		}
		return DeepKeepDiff(v1.Elem(), v2.Elem(), depth+1)

	case reflect.Ptr:
		//if v1/v2 is nil pointer, the Elem() will return zero value, which will be discarded on next level of function call
		return DeepKeepDiff(v1.Elem(), v2.Elem(), depth+1)

	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			fieldName := v1.Type().Field(i).Name
			if strings.HasPrefix(fieldName, "XXX_") { //protobuf generated fields that begin with XXX_
				continue
			}
			f1 := v1.Field(i)
			f2 := v2.Field(i)
			if f1.CanAddr() && f2.CanAddr() {
				if err := DeepKeepDiff(f1.Addr(), f2.Addr(), depth+1); err != nil {
					return err
				}
			} else {
				return errors.New("cannot get address for field: " + fieldName)
			}
		}
		return nil

	case reflect.Map:
		if v1.IsNil() || v2.IsNil() {
			return nil
		}
		commonKeys := []reflect.Value{}
		for _, k := range v1.MapKeys() {
			val1 := v1.MapIndex(k)
			val2 := v2.MapIndex(k)
			if !val1.IsValid() || !val2.IsValid() {
				continue
			}
			if val1.CanInterface() && val2.CanInterface() && reflect.DeepEqual(val1.Interface(), val2.Interface()) {
				commonKeys = append(commonKeys, k)
			} else {
				if err := DeepKeepDiff(ptr(val1), ptr(val2), depth+1); err != nil {
					return err
				}
			}
		}
		for _, k := range commonKeys {
			v1.SetMapIndex(k, reflect.Value{})
			v2.SetMapIndex(k, reflect.Value{})
		}
		return nil

	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return nil
		}
		// Can't do better than this:
		return errors.New("can't get diff for function")
	case reflect.String:
		//we want to see if it is some marshalled json,
		//tryChangeMap := func(mapType reflect.Value) {
		//	ok := false
		//	map1 := mapType.Interface()
		//	map2 := mapType.Interface()
		//	fmt.Printf("begin=%s\nbegin=%s\n", v1.String(), v2.String())
		//	if err := json.Unmarshal([]byte(v1.String()), &map1); err == nil {
		//		if err := json.Unmarshal([]byte(v2.String()), &map2); err == nil {
		//			fmt.Printf("map1=%+v\n", map1)
		//			fmt.Printf("map2=%+v\n", map2)
		//			if err := DeepKeepDiff(reflect.ValueOf(&map1), reflect.ValueOf(&map2), depth+1); err == nil {
		//				ok = true
		//			}
		//		}
		//	}
		//	fmt.Printf("finally cansets: %v %v, ok: %v\n", v1.CanSet(), v2.CanSet(), ok)
		//	if ok {
		//		if v1.CanSet() && v2.CanSet() {
		//			if b1, err := json.Marshal(map1); err == nil {
		//				if b2, err := json.Marshal(map2); err == nil {
		//					v1.Set(reflect.ValueOf(string(b1)))
		//					v2.Set(reflect.ValueOf(string(b2)))
		//				}
		//			}
		//		}
		//	}
		//}
		//tryChangeMap(reflect.ValueOf([]map[string]interface{}{}))
		//tryChangeMap(reflect.ValueOf(map[string]interface{}{}))

		{
			ok := false
			map1 := map[string]interface{}{}
			map2 := map[string]interface{}{}
			fmt.Printf("begin=%s\nbegin=%s\n", v1.String(), v2.String())
			if err := json.Unmarshal([]byte(v1.String()), &map1); err == nil {
				if err := json.Unmarshal([]byte(v2.String()), &map2); err == nil {
					fmt.Printf("map1=%+v\n", map1)
					fmt.Printf("map2=%+v\n", map2)
					if err := DeepKeepDiff(reflect.ValueOf(&map1), reflect.ValueOf(&map2), depth+1); err != nil {
						return err
					} else {
						ok = true
					}
				}
			}
			fmt.Printf("finally cansets: %v %v, ok: %v\n", v1.CanSet(), v2.CanSet(), ok)
			if ok {
				if v1.CanSet() && v2.CanSet() {
					if b1, err := json.Marshal(map1); err == nil {
						if b2, err := json.Marshal(map2); err == nil {
							fmt.Printf("final1=%+v\n", string(b1))
							fmt.Printf("final2=%+v\n", string(b2))
							v1.Set(reflect.ValueOf(string(b1)))
							v2.Set(reflect.ValueOf(string(b2)))
						}
					}
				}
			}
		}

		{
			ok := false
			map1 := []map[string]interface{}{}
			map2 := []map[string]interface{}{}
			fmt.Printf("[]begin=%s\n[]begin=%s\n", v1.String(), v2.String())
			if err := json.Unmarshal([]byte(v1.String()), &map1); err == nil {
				if err := json.Unmarshal([]byte(v2.String()), &map2); err == nil {
					fmt.Printf("[]map1=%+v\n", map1)
					fmt.Printf("[]map2=%+v\n", map2)
					if err := DeepKeepDiff(reflect.ValueOf(&map1), reflect.ValueOf(&map2), depth+1); err != nil {
						return err
					} else {
						ok = true
					}

				}
			}
			fmt.Printf("[]finally cansets: %v %v, ok: %v\n", v1.CanSet(), v2.CanSet(), ok)
			if ok {
				if v1.CanSet() && v2.CanSet() {
					if b1, err := json.Marshal(map1); err == nil {
						if b2, err := json.Marshal(map2); err == nil {
							fmt.Printf("[]final1=%+v\n", string(b1))
							fmt.Printf("[]final2=%+v\n", string(b2))
							v1.Set(reflect.ValueOf(string(b1)))
							v2.Set(reflect.ValueOf(string(b2)))
						}
					}
				}
			}
		}

		return nil
	default:
		// Normal equality suffices
		return nil
	}
}
