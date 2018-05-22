// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

// Package merger recursively merge two data structures into new one. Only maps
// and structures are recursively merging. Values of other kinds (e.g. slices)
// do not merging. Non-zero value from the right side has precedence.
package merger

import "reflect"

// Merge method performs recursive merge of two data structures into new one.
func Merge(left, right interface{}) interface{} {
	result := merge(
		reflect.ValueOf(left),
		reflect.ValueOf(right),
	)

	if !result.IsValid() {
		return nil
	}

	return result.Interface()
}

func merge(left, right reflect.Value) reflect.Value {
	leftKind := left.Kind()
	rightKind := right.Kind()

	if leftKind == reflect.Interface {
		left = left.Elem()
		leftKind = left.Kind()
	}
	if rightKind == reflect.Interface {
		right = right.Elem()
		rightKind = right.Kind()
	}

	if !left.IsValid() {
		return right
	}
	if !right.IsValid() {
		return left
	}

	if leftKind == reflect.Map &&
		rightKind == reflect.Map {

		return mergeMap(left, right)
	}

	if leftKind == reflect.Struct &&
		rightKind == reflect.Struct {

		return mergeStruct(left, right)
	}

	if leftKind == reflect.Ptr &&
		rightKind == reflect.Ptr {

		left := left.Elem()
		leftKind := left.Kind()

		right := right.Elem()
		rightKind := right.Kind()

		if leftKind == reflect.Struct &&
			rightKind == reflect.Struct {

			return mergeStruct(left, right).Addr()
		}
		if leftKind == reflect.Map &&
			rightKind == reflect.Map {

			return mergeMap(left, right).Addr()
		}
	}

	if isZero(right) {
		return left
	}

	return right
}

func mergeMap(left, right reflect.Value) reflect.Value {
	bt := right.Type()
	result := reflect.MakeMap(bt)

	for _, key := range left.MapKeys() {
		result.SetMapIndex(key, left.MapIndex(key))
	}

	for _, key := range right.MapKeys() {
		value := merge(result.MapIndex(key), right.MapIndex(key))
		result.SetMapIndex(key, value)
	}

	return result
}

func mergeStruct(left, right reflect.Value) reflect.Value {
	leftType := left.Type()
	rightType := right.Type()

	if leftType != rightType {
		return right
	}

	result := reflect.New(rightType).Elem()

	for i := 0; i < rightType.NumField(); i++ {
		leftFVal := left.Field(i)
		rightFVal := right.Field(i)
		resFVal := result.Field(i)

		if resFVal.Kind() == reflect.Interface &&
			isZero(leftFVal) && isZero(rightFVal) {

			continue
		}

		if resFVal.CanSet() {
			resFVal.Set(merge(leftFVal, rightFVal))
		}
	}

	return result
}

func isZero(value reflect.Value) bool {
	zero := reflect.Zero(value.Type())
	return reflect.DeepEqual(zero.Interface(), value.Interface())
}
