package main

import "testing"

func TestCheckerPrefix(t *testing.T) {
	badUUID := "string"
	goodUUID := checkerScheme + ":" + "field1-field2"
	badResult, failure := checkerPrefix(badUUID)
	goodResult, success := checkerPrefix(goodUUID)
	if badResult != "" || failure != false {
		t.Errorf("CheckerPrefix should not accept value: %s", badResult)
	}
	if goodResult != "field1" || success != true {
		t.Errorf("CheckerPrefix should accept value: %s", goodResult)
	}
}