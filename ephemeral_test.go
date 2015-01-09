package main

// Start a redis server in memory-only mode on a random port.

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
)

const (
	executable = "redis-server"
)

type ephemeral exec.Cmd

// Redis starts a memory-only redis on a random port. Will panic if that
// doesn't work.
// Returns something which you'll have to Close(), and a string to give to Dial()
func Redis() (*ephemeral, string) {
	port := arbitraryPort()

	c := exec.Command(executable, "-")
	stdin, err := c.StdinPipe()
	if err != nil {
		panic(err)
	}
	stdin.Write([]byte(fmt.Sprintf("port %d\nbind 127.0.0.1\n", port)))
	stdin.Close()
	err = c.Start()
	if err != nil {
		panic(err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	if waitForPort(addr) {
		e := ephemeral(*c)
		return &e, addr
	}

	panic(fmt.Sprintf("No connection on port %d", port))
}

func (e *ephemeral) Close() {
	((*exec.Cmd)(e)).Process.Kill()
	((*exec.Cmd)(e)).Wait()
}

// arbitraryPort returns a non-used port.
func arbitraryPort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	defer l.Close()
	addr := l.Addr().String()
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	p, _ := strconv.Atoi(port)
	return p
}
