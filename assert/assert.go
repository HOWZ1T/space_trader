// Assert provides useful utilities for writing unit tests.
//
// Assert wraps the standard testing.T type to implement its utilities on that type.
package assert

import "testing"

type T testing.T

// Asserts that A and B are Equal.
//
// Example:
//  (*assert.T)(t).Equals(10, 10)
func (t *T) Equals(a interface{}, b interface{}) {
	res := a == b
	if !res {
		t.Fatalf("Assert: %v == %v -> %t", a, b, res)
	}
}

// Asserts that A and B are Not Equal.
//
// Example:
//  (*assert.T)(t).NotEquals(10, 5)
func (t *T) NotEquals(a interface{}, b interface{}) {
	res := a != b
	if !res {
		t.Fatalf("Assert: %v != %v -> %t", a, b, res)
	}
}

// Asserts that A is Nil
//
// Example:
//  (*assert.T)(t).Nil(nil)
func (t *T) Nil(a interface{}) {
	res := a == nil
	if !res {
		t.Fatalf("Assert: %v is nil -> %t", a, res)
	}
}

// Asserts that A is not Nil
//
// Example:
//  (*assert.T)(t).NotNil([]string{"hello", "world"})
func (t *T) NotNil(a interface{}) {
	res := a != nil
	if !res {
		t.Fatalf("Assert: %v is not nil -> %t", a, res)
	}
}
