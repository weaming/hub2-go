package main

import (
	"fmt"
	"log"
	"reflect"
)

func fatalErr(err error, prefix string) {
	if err != nil {
		log.Fatal(prefix+":", err)
	}
}

func str(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func sliceDel(xs []interface{}, i int) []interface{} {
	if i > len(xs)-1 {
		return xs
	}
	if i == len(xs)-1 {
		goto tripTail
	}

	xs[i] = xs[len(xs)-1] // Copy last element to index i

tripTail:
	xs[len(xs)-1] = reflect.Zero(reflect.TypeOf(xs[len(xs)-1])) // Erase last element (write zero value).
	xs = xs[:len(xs)-1]
	return xs
}

type Set map[string]uint8

func NewSet(xs ...string) Set {
	s := Set{}
	for _, x := range xs {
		s.Add(x)
	}
	return s
}

func (s Set) Add(k string) {
	s[k] = 1
}

func (s Set) Has(k string) bool {
	_, ok := s[k]
	return ok
}

func (s Set) Del(k string) bool {
	if _, ok := s[k]; ok {
		delete(s, k)
		return true
	}
	return false
}

func (s Set) Arr() (rv []string) {
	for k := range s {
		rv = append(rv, k)
	}
	return
}
