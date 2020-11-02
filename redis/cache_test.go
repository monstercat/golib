package redis

import (
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	c, err := NewCache(connStr)
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
}

func TestCacheSetGetDel(t *testing.T) {
	c, err := NewCache(connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	val := "meow"
	err = c.Set("cat", val, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	buf, ok, err := c.Get("cat")
	if err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatalf("Should have been OK.")
	} else if val != buf {
		t.Fatalf("Get should have returned %s, got %s", val, buf)
	}

	time.Sleep(time.Millisecond * 1100)
	buf, ok, err = c.Get("cat")
	if err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatalf("Should have not been OK, should have expired.")
	} else if buf == val {
		t.Fatalf("Returned value should not have been set.")
	}

	err = c.Set("mix", val, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Del("mix")
	if err != nil {
		t.Fatal(err)
	}
	buf, ok, err = c.Get("mix")
	if err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatalf("Should not have gotten ok")
	}
}

// This should simulate a scenario where you load data from the database.
func TestCache_WipeLocal(t *testing.T) {
	c, err := NewCache(connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	err = c.Set("cat", "meow", 0)
	if err != nil {
		t.Fatal(err)
	}

	c.WipeLocal()
	if item, ok := c.store["cat"]; ok || item.Value == "meow" {
		t.Fatalf("Key 'cat' should not be set locally.")
	}

	val, ok, err := c.Get("cat")
	if err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatalf("Should have gotten OK.")
	} else if val != "meow" {
		t.Fatalf("Value should have been 'meow', got %s", val)
	}

	c.Del("cat")
}
