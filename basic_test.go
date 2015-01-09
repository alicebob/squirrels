package main

import (
	"testing"

	"github.com/garyburd/redigo/redis"
)

func TestBasic(t *testing.T) {
	conn, cb := setup(t, 2)
	defer cb()

	if _, err := conn.Do("SET", "foo", "bar"); err != nil {
		t.Fatalf("SET err: %s", err)
	}

	v, err := redis.String(conn.Do("GET", "foo"))
	if err != nil {
		t.Fatalf("GET err: %s", err)
	}
	if have, want := v, "bar"; have != want {
		t.Fatalf("have: %s, want: %s", have, want)
	}
}

func TestMany(t *testing.T) {
	// Run many GET commands.
	conn, cb := setup(t, 3)
	defer cb()

	if _, err := conn.Do("SET", "foo", "bar"); err != nil {
		t.Fatalf("SET err: %s", err)
	}

	for i := 0; i < 10000; i++ {
		v, err := redis.String(conn.Do("GET", "foo"))
		if err != nil {
			t.Fatalf("GET err: %s", err)
		}
		if have, want := v, "bar"; have != want {
			t.Fatalf("have: %s, want: %s", have, want)
		}
	}
}

func TestUnknown(t *testing.T) {
	// Run a command we don't support.
	conn, cb := setup(t, 3)
	defer cb()

	_, err := conn.Do("MGET", "foo", "bar")
	if err == nil {
		t.Fatalf("no MGET err")
	}
}
