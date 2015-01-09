package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
)

func waitForPort(addr string) bool {
	// Wait until addr is ready.
	timeout := time.Now().Add(5 * time.Second)
	for time.Now().Before(timeout) {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func runSquirrels(t testing.TB, addrs []string) *exec.Cmd {
	var remotes []string
	for i, addr := range addrs {
		remotes = append(remotes, fmt.Sprintf("%s=srv%d", addr, i))
	}
	cmd := exec.Command(
		"./squirrels",
		"-remotes", strings.Join(remotes, ","),
	)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("can't start squirrels: %s", err)
	}
	if !waitForPort("localhost:7000") {
		t.Fatalf("squirrels don't listen")
	}
	return cmd
}

// setup runs a squirrels server and i in-memory redisses.
// Returns a connection and a callback which you have to call on exit.
func setup(t testing.TB, i int) (redis.Conn, func()) {
	var toClose []func()
	var addrs []string
	for i := 0; i < 5; i++ {
		r, a := Redis()
		toClose = append(toClose, func() { r.Close() })
		addrs = append(addrs, a)
	}
	// fmt.Printf("tmp redisses: %s\n", addrs)

	s := runSquirrels(t, addrs)
	toClose = append(toClose, func() {
		s.Process.Kill()
		s.Wait() // don't check the error, we killed it
		time.Sleep(10 * time.Millisecond)
	})

	conn, err := redis.Dial("tcp", "localhost:7000")
	if err != nil {
		t.Fatalf("can't connect to squirrels: %s", err)
	}
	return conn, func() {
		for _, c := range toClose {
			c()
		}
	}
}
