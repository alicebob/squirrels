package main

import (
	"bufio"
	"bytes"
	"errors"
	// "log"
)

var (
	ErrProtocol = errors.New("invalid request")
)

// readResponse reads a single command from the reader and returns the
// bytes. Return raw and the cmd (or nil for arrays)
func readResponse(rd *bufio.Reader) ([]byte, []byte, error) {
	// a basic command is `<type byte><cmd>[ <arguments>]\r\n`
	line, err := rd.ReadBytes('\n')
	if err != nil {
		return nil, nil, err
	}
	if len(line) < 3 {
		return nil, nil, ErrProtocol
	}

	// log.Printf("readResponse for %s", token)

	switch line[0] {
	default:
		return nil, nil, ErrProtocol
	case '+', '-', ':':
		// +: simple string
		// -: errors
		// :: integer
		// Simple line based replies.
		return line, line[1 : len(line)-2], nil
	case '$':
		// bulk strings
		// These are: `$5\r\nhello\r\n`
		length, err := parseDec(line[1 : len(line)-2])
		if err != nil {
			return nil, nil, err
		}
		if length < 0 {
			// -1 is a nil response
			return line, nil, nil
		}
		buf := make([]byte, length+2)
		// TODO: check response
		if _, err = rd.Read(buf); err != nil {
			return nil, nil, err
		}
		return append(line, buf...), buf[:len(buf)-2], nil
	case '*':
		// arrays
		// These are: `*2\r\n+hello\r\n+world\r\n`
		l, err := parseDec(line[1 : len(line)-2])
		if err != nil {
			return nil, nil, err
		}
		// l can be -1
		buf := bytes.NewBuffer(line)
		for ; l > 0; l-- {
			// log.Printf("sub readResponse %d", l)
			sub, _, err := readResponse(rd)
			if err != nil {
				return nil, nil, err
			}
			buf.Write(sub)
		}
		return buf.Bytes(), nil, nil
	}
}

// client always talks in arrays.
func readArray(rd *bufio.Reader) ([]byte, [][]byte, error) {
	line, err := rd.ReadBytes('\n')
	if err != nil {
		return nil, nil, err
	}
	if len(line) < 3 {
		return nil, nil, ErrProtocol
	}

	switch line[0] {
	default:
		return nil, nil, ErrProtocol
	case '*':
		// arrays
		// These are: `*2\r\n+hello\r\n+world\r\n`
		l, err := parseDec(line[1 : len(line)-2])
		if err != nil {
			return nil, nil, err
		}
		// l can be -1
		buf := bytes.NewBuffer(line)
		var fields [][]byte
		for ; l > 0; l-- {
			// log.Printf("sub readResponse %d", l)
			sub, cmd, err := readResponse(rd)
			if err != nil {
				return nil, nil, err
			}
			buf.Write(sub)
			fields = append(fields, cmd)
		}
		return buf.Bytes(), fields, nil
	}
}
