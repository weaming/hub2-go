package main

import "log"

func fatalErr(err error, prefix string) {
	if err != nil {
		log.Fatal(prefix+":", err)
	}
}
