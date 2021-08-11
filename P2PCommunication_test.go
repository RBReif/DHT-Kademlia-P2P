package main

import (
	"testing"
)

func TestDistance(t *testing.T) {
	id1 := id{}
	id2 := id{}

	if distance(id1, id2) != id1 {
		t.Errorf("Distance of two empty ids has to be empty id")
	}
}
