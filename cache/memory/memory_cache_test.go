package memory_cache

import (
	"reflect"
	"testing"
)

func TestCache_Get(t *testing.T) {
	c := New()
	err := c.Set("1", 1)
	if err != nil {
		t.Error(err)
	}
	err = c.Set("2", "zwei")
	if err != nil {
		t.Error(err)
	}
	a1, err := c.Get("1")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a1, 1) {
		t.Error("should equal")
	}
	a2, err := c.Get("2")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a2, "zwei") {
		t.Error("should equal")
	}
	a3, err := c.Get("not found")
	if err != nil || a3 != nil {
		t.Error("error:", err, "a3", a3)
	}
	err = c.Del("2")
	if err != nil {
		t.Error(err)
	}
	a2, err = c.Get("2")
	if err != nil || a2 != nil {
		t.Error(err)
	}
	err = c.Empty()
	if err != nil {
		t.Error(err)
	}
	a1, err = c.Get("1")
	if err != nil || a1 != nil {
		t.Error(err)
	}
}
