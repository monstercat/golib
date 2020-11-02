package redis

import (
	"testing"
	"time"
)

const connStr = "tcp://localhost:6379"

func TestNewRedis(t *testing.T) {
	r, err := NewRedis(connStr)
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
}

func TestRedisSetGetDel(t *testing.T) {
	r, err := NewRedis(connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	sec := time.Second
	val := "hey there!"
	err = r.Set("hey", val, sec)
	if err != nil {
		t.Fatal(err)
	}

	buf, ok, err := r.Get("hey")
	if err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatalf("Getting 'hey' should be OK")
	} else if string(buf) != val {
		t.Fatalf("Should have gotten '%s', got '%s'", val, string(buf))
	}

	time.Sleep(sec + time.Millisecond * 100)
	buf, ok, err = r.Get("hey")
	if err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatalf("Key 'hey' should not be set - should have expired.")
	} else if string(buf) == val {
		t.Fatalf("Value returned should have been empty.")
	}

	// Setting with 0 should work!
	err = r.Set("key", "value", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Del("key"); err != nil {
		t.Fatal(err)
	}
}

func TestRedis_Scan(t *testing.T) {
	r, err := NewRedis(connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	err = r.Set("key", "value", 0)
	if err != nil {
		t.Fatal(err)
	}

	xs, cursor, err := r.Scan(0, "key")
	if err != nil {
		t.Fatal(err)
	} else if cursor != 0 {
		t.Fatalf("Cursor should have reached end of line and be 0.")
	} else if len(xs) != 1 {
		t.Fatalf("Should have gotten 1 result for key search, got %d.", len(xs))
	} else if xs[0] != "key" {
		t.Fatalf("Key did not match.")
	}

	r.Del("key")
}

func TestRedis_DeleteKeyMatch(t *testing.T) {
	r, err := NewRedis(connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	keys := map[string]string{
		"catA:1": "1",
		"catA:2": "2",
		"catB:1": "3",
	}
	for key, val := range keys {
		err := r.Set(key, val, 0)
		if err != nil {
			t.Fatal(err)
		}
	}

	count, err := r.DeleteKeyMatch("catA:*")
	if err != nil {
		t.Fatal(err)
	} else if count != 2 {
		t.Fatalf("Should have deleted 2 keys, got %d", count)
	}
}