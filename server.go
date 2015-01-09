package main

import (
	"bufio"
	"log"
	"net"

	"github.com/ngerakines/ketama"
)

type RemoteReq struct {
	req []byte
	res chan []byte
}

type Server struct {
	remotes map[string]chan RemoteReq // named servers
	ring    ketama.HashRing
}

// NewServer expects a map with servername -> addr
func NewServer(remotes map[string]string) *Server {
	s := Server{
		remotes: make(map[string]chan RemoteReq, len(remotes)),
		ring:    ketama.NewRing(10000), // TODO: number?
	}
	for n, addr := range remotes {
		cmds := make(chan RemoteReq, 1000)
		s.remotes[n] = cmds
		s.ring.Add(n, 40)
		go startRemote(addr, cmds)
	}
	return &s
}

// We listen to things send to us on cmds, and write them back.
func startRemote(addr string, cmds chan RemoteReq) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("dial err: %s", err)
		// TODO: loop
		return
	}
	buff := bufio.NewReader(conn)
	wr := bufio.NewWriter(conn)
	// We read as many commands as are available, and then read the replies.
	var replies []RemoteReq
	for c := range cmds {
		replies = replies[:0]
	loop:
		for {
			// log.Printf("cmd for remote: %s", string(c.req))
			n, err := wr.Write(c.req)
			if err != nil {
				log.Printf("client write err: %s", err)
				// TODO: loop
				return
			}
			if n != len(c.req) {
				log.Printf("client write length err: %d/%d", n, len(c.req))
				// TODO: loop
				return
			}
			replies = append(replies, c)
			if len(replies) > 1000 {
				// That'll do.
				break loop
			}
			// Any more command waiting Right Now?
			select {
			case c = <-cmds:
			default:
				break loop
			}
		}
		if err := wr.Flush(); err != nil {
			log.Printf("server write err: %s", err)
			// TODO: loop
			return
		}

		for _, r := range replies {
			response, _, err := readResponse(buff)
			// log.Printf("response from remote: %s (%s)", string(response), err)
			if err != nil {
				log.Printf("client response err: %s", err)
				// TODO: loop
				return
			}
			r.res <- response
		}
	}
}

func (s *Server) ListenAndServe(listenAddr string) error {
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.serveClient(conn)
	}
}

// serveClient starts a new session, using `conn` as a transport.
func (s *Server) serveClient(conn net.Conn) {

	// log.Printf("accepted %s", conn.RemoteAddr())

	buffer := bufio.NewReader(conn)

	// We read requests as fast as possible, and put for every request a
	// channel to get the response from in resp. This way we read all requests
	// the moment they come in, and still reply in order, as fast as we can.
	resp := make(chan chan []byte, 100)
	go func() {
		defer conn.Close()
		// defer log.Printf("closing %s", conn.RemoteAddr())
		var (
			wr  = bufio.NewWriter(conn)
			res []byte
		)
		r := <-resp
		for {
			// log.Printf("waiting for a response...")
			select {
			case res = <-r:
			default:
				// nothing available right away
				if err := wr.Flush(); err != nil {
					log.Printf("client write err: %s", err)
					return
				}
				res = <-r
			}
			// log.Printf("about to send: %s", string(res))
			n, err := wr.Write(res)
			if err != nil {
				log.Printf("client write resp err: %s", err)
				return
			}
			if n != len(res) {
				log.Printf("client write length: %d/%d", n, len(res))
				return
			}
			select {
			case r = <-resp:
				// TODO return on ! ok
			default:
				// nothing available right away
				if err := wr.Flush(); err != nil {
					log.Printf("client write err: %s", err)
					return
				}
				r = <-resp
				// TODO return on ! ok
			}
		}
	}()

	for {
		res := make(chan []byte, 1)
		resp <- res

		var key string

		// we only expect 'new style' (array) requests.
		raw, fields, err := readArray(buffer)
		if err != nil {
			// log.Printf("client req err: %s", err)
			res <- []byte("-ERR syntax error\v\n")
			close(resp)
			return
		}
		if len(fields) < 1 {
			// log.Printf("client proto err: no fields")
			res <- []byte("-ERR syntax error\v\n")
			close(resp)
			return
		}
		cmd := string(fields[0])
		switch cmd {
		case "GET", "SET":
			// supported command
			if len(fields) < 2 {
				res <- []byte("-ERR syntax error\v\n")
				close(resp)
				return
			}
			key = string(fields[1])
		default:
			res <- []byte("-ERR unsupported command\v\n")
			close(resp)
			return
		}

		// log.Printf("client req: %s", string(req.Raw))
		// log.Printf("server for %q: %s", key, s.ring.Hash(key))
		s.remotes[s.ring.Hash(key)] <- RemoteReq{
			req: raw,
			res: res,
		}
	}
}
