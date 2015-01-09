package main

import (
	"flag"
	"log"
	"strings"
)

var (
	addrs  = flag.String("remotes", "localhost:6379=me", "list of redisses to connect to. format: 1.2.3.4:6379=name1,1.2.3.5:6379=name2")
	listen = flag.String("listen", "localhost:7000", "listen address")
)

func main() {
	flag.Parse()

	remotes := map[string]string{}
	for _, a := range strings.Split(*addrs, ",") {
		t := strings.SplitN(a, "=", 2)
		remotes[t[1]] = t[0]
	}
	log.Printf("listening on %s\n", *listen)
	log.Printf("connecting to: %v\n", remotes)
	s := NewServer(remotes)
	log.Print(s.ListenAndServe(*listen))
}
