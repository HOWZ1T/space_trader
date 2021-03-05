package assert

import "testing"

type T testing.T

func (t *T) Equals(a interface{}, b interface{}) {
	res := a == b
	if !res {
		t.Fatalf("Assert: %v == %v -> %t", a, b, res)
	}
}

func (t *T) NotEquals(a interface{}, b interface{}) {
	res := a != b
	if !res {
		t.Fatalf("Assert: %v != %v -> %t", a, b, res)
	}
}

func (t *T) Nil(a interface{}) {
	res := a == nil
	if !res {
		t.Fatalf("Assert: %v is nil -> %t", a, res)
	}
}

func (t *T) NotNil(a interface{}) {
	res := a != nil
	if !res {
		t.Fatalf("Assert: %v is not nil -> %t", a, res)
	}
}
