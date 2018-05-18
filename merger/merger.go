// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

// Package merger recursively merge two data structures into new one. Only maps
// and structures are recursively merging. Values of other kinds (e.g. slices)
// do not merging. Non-zero value from the right side has precedence.
package merger

import "reflect"

// Merge method performs recursive merge of two data structures into new one.
func Merge(a, b interface{}) interface{} {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	cv := merge(av, bv)

	if !cv.IsValid() {
		return nil
	}

	return cv.Interface()
}

func merge(av, bv reflect.Value) reflect.Value {
	ak := av.Kind()
	bk := bv.Kind()

	if ak == reflect.Interface {
		av = av.Elem()
		ak = av.Kind()
	}
	if bk == reflect.Interface {
		bv = bv.Elem()
		bk = bv.Kind()
	}

	if !av.IsValid() {
		return bv
	}
	if !bv.IsValid() {
		return av
	}

	if ak == reflect.Map && bk == reflect.Map {
		return mergeMap(av, bv)
	}
	if ak == reflect.Struct && bk == reflect.Struct {
		return mergeStruct(av, bv)
	}
	if ak == reflect.Ptr && bk == reflect.Ptr {
		ae := av.Elem()
		be := bv.Elem()

		aek := ae.Kind()
		bek := be.Kind()

		if aek == reflect.Struct && bek == reflect.Struct {
			return mergeStruct(ae, be).Addr()
		}
		if aek == reflect.Map && bek == reflect.Map {
			return mergeMap(ae, be).Addr()
		}
	}

	if isZero(bv) {
		return av
	}

	return bv
}

func mergeMap(av, bv reflect.Value) reflect.Value {
	bt := bv.Type()

	cv := reflect.New(bt).Elem()
	cv.Set(reflect.MakeMap(bt))

	for _, kv := range av.MapKeys() {
		cv.SetMapIndex(kv, av.MapIndex(kv))
	}
	for _, kv := range bv.MapKeys() {
		cv.SetMapIndex(kv, merge(cv.MapIndex(kv), bv.MapIndex(kv)))
	}

	return cv
}

func mergeStruct(av, bv reflect.Value) reflect.Value {
	at := av.Type()
	bt := bv.Type()

	if at.Name() != bt.Name() ||
		at.PkgPath() != bt.PkgPath() {

		return bv
	}

	cv := reflect.New(bt).Elem()

	for i := 0; i < bt.NumField(); i++ {
		afv := av.Field(i)
		bfv := bv.Field(i)
		cfv := cv.Field(i)

		if cfv.Kind() == reflect.Interface &&
			isZero(afv) && isZero(bfv) {

			continue
		}

		if cfv.CanSet() {
			cfv.Set(merge(afv, bfv))
		}
	}

	return cv
}

func isZero(v reflect.Value) bool {
	zv := reflect.Zero(v.Type())
	return reflect.DeepEqual(zv.Interface(), v.Interface())
}
