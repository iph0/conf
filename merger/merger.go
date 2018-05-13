// Package merger recursively merges separate configuration maps or structures
// to one configuration tree.
package merger

import "reflect"

// Merge method performs recursive merging data
func Merge(a, b interface{}) interface{} {
	aV := reflect.ValueOf(a)
	bV := reflect.ValueOf(b)

	cV := merge(aV, bV)

	if !cV.IsValid() {
		return nil
	}

	return cV.Interface()
}

func merge(aV, bV reflect.Value) reflect.Value {
	aK := aV.Kind()
	bK := bV.Kind()

	if aK == reflect.Interface {
		aV = aV.Elem()
		aK = aV.Kind()
	}
	if bK == reflect.Interface {
		bV = bV.Elem()
		bK = bV.Kind()
	}

	if !aV.IsValid() {
		return bV
	}
	if !bV.IsValid() {
		return aV
	}

	if aK == reflect.Map && bK == reflect.Map {
		return mergeMap(aV, bV)
	}
	if aK == reflect.Struct && bK == reflect.Struct {
		return mergeStruct(aV, bV)
	}
	if aK == reflect.Ptr && bK == reflect.Ptr {
		aE := aV.Elem()
		bE := bV.Elem()

		aEK := aE.Kind()
		bEK := bE.Kind()

		if aEK == reflect.Struct && bEK == reflect.Struct {
			return mergeStruct(aE, bE).Addr()
		}
		if aEK == reflect.Map && bEK == reflect.Map {
			return mergeMap(aE, bE).Addr()
		}
	}

	if isZero(bV) {
		return aV
	}

	return bV
}

func mergeMap(aV, bV reflect.Value) reflect.Value {
	bT := bV.Type()

	cV := reflect.New(bT).Elem()
	cV.Set(reflect.MakeMap(bT))

	for _, kV := range aV.MapKeys() {
		cV.SetMapIndex(kV, aV.MapIndex(kV))
	}
	for _, kV := range bV.MapKeys() {
		cV.SetMapIndex(kV, merge(cV.MapIndex(kV), bV.MapIndex(kV)))
	}

	return cV
}

func mergeStruct(aV, bV reflect.Value) reflect.Value {
	aT := aV.Type()
	bT := bV.Type()

	if aT.Name() != bT.Name() ||
		aT.PkgPath() != bT.PkgPath() {

		return bV
	}

	cV := reflect.New(bT).Elem()

	for i := 0; i < bT.NumField(); i++ {
		aFV := aV.Field(i)
		bFV := bV.Field(i)
		cFV := cV.Field(i)

		if cFV.Kind() == reflect.Interface &&
			isZero(aFV) && isZero(bFV) {

			continue
		}

		if cFV.CanSet() {
			cFV.Set(merge(aFV, bFV))
		}
	}

	return cV
}

func isZero(v reflect.Value) bool {
	zV := reflect.Zero(v.Type())
	return reflect.DeepEqual(zV.Interface(), v.Interface())
}
