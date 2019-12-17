package main

import "testing"

func Test_addNumbers(t *testing.T) {
	expectedSum := 8

	sum := addNumbers(3, 5)
	if sum != expectedSum {
		t.Errorf("addNumbers() = %v, want %v", sum, expectedSum)
	}
}
