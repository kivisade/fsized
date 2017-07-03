package main

import (
	"testing"
	"fmt"
)

func TestP2(t *testing.T) {
	type TestCase struct {
		input uint64
		expected uint
	}

	var testSuite []TestCase = []TestCase{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 1},
		{7, 2},
		{14, 3},
		{30, 4},
		{33, 5},
		{70, 6},
		{202, 7},
		{511, 8},
		{513, 9},
		{1024, 10},
		{1025, 10},
	}

	for _, test := range testSuite {
		if actual := p2(test.input); actual != test.expected {
			t.Errorf("Failed on input %d: expected %d, got %d", test.input, test.expected, actual)
		} else {
			fmt.Printf("Passed on input %d: result is %d\n", test.input, actual)
		}
	}
}
