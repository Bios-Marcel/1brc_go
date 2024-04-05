package main

import "testing"

func Test_parseNumber(t *testing.T) {
	if r := parseNumber([]byte(`12.5`)); r != 125 {
		t.Fatal(r)
	}
}
