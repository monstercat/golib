package redis

import (
	"fmt"
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

	i := 0
	for i < 5000 {
		err = r.Set(fmt.Sprintf("key:%d", i), fmt.Sprintf("value%d", i), 0)
		if err != nil {
			t.Fatal(err)
		}
		i++
	}

	xs, cursor, err := r.Scan(0, "key:*")
	if err != nil {
		t.Fatal(err)
	} else if len(xs) <= 0 {
		t.Fatalf("Should have gotten some results for key search, got 0.")
	} else if cursor == 0 {
		t.Fatalf("Should have gotten a new cursor.")
	}

	// iterate through the rest
	var count int
	for cursor != 0 {
		xs, cursor, err = r.Scan(cursor, "key:*")
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count < 1 {
		t.Errorf("Expected to have iterated via SCAN atleast once.")
	}

	r.DeleteKeyMatch("key:*")
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