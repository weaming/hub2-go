package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

func fatalErr(err error, prefix string) {
	if err != nil {
		log.Fatal(prefix+":", err)
	}
}

func str(v interface{}) string {
	return fmt.Sprintf("%v", v)
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

func prepareDir(filePath string, forceDir bool) {
	if !strings.HasSuffix(filePath, "/") || forceDir {
		filePath = path.Dir(filePath)
	}
	if err := os.MkdirAll(filePath, os.FileMode(0755)); err != nil {
		log.Fatal(err)
	}
}

func writeJson(filePath string, v interface{}) {
	bytes, err := json.Marshal(v)
	fatalErr(err, "marshal")

	err = ioutil.WriteFile(filePath, bytes, 0644)
	fatalErr(err, "write")
}

func readJson(filePath string, v interface{}) {
	log.Printf("reading json %s\n", filePath)
	jsonFile, err := os.Open(filePath)
	fatalErr(err, "open")
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	fatalErr(err, "read")

	err = json.Unmarshal(bytes, v)
	fatalErr(err, "unmarshal")
}
