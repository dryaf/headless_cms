package memory_cache

import (
	"context"
	"reflect"
	"testing"
)

func TestCache_Get(t *testing.T) {
	ctx := context.TODO()
	c := New()
	err := c.Set(ctx, "1", []byte("1"))
	if err != nil {
		t.Error(err)
	}
	err = c.Set(ctx, "2", []byte("zwei"))
	if err != nil {
		t.Error(err)
	}
	a1, err := c.Get(ctx, "1")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a1, []byte("1")) {
		t.Error("should equal")
	}
	a2, err := c.Get(ctx, "2")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a2, []byte("zwei")) {
		t.Error("should equal")
	}
	a3, err := c.Get(ctx, "not found")
	if err != nil || a3 != nil {
		t.Error("error:", err, "a3", a3)
	}
	err = c.Del(ctx, "2")
	if err != nil {
		t.Error(err)
	}
	a2, err = c.Get(ctx, "2")
	if err != nil || a2 != nil {
		t.Error(err)
	}
	err = c.Empty(ctx)
	if err != nil {
		t.Error(err)
	}
	a1, err = c.Get(ctx, "1")
	if err != nil || a1 != nil {
		t.Error(err)
	}
}
