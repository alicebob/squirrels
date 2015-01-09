package main

import (
	"errors"
)

var (
	ErrInvalidNumber = errors.New("invalid number")
)

func parseDec(s []byte) (int, error) {
	n := 0
	sign := 1
	if s[0] == '-' {
		sign = -1
		s = s[1:]
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, ErrInvalidNumber
		}
		n *= 10
		n += int(c - '0')
	}
	return n * sign, nil
}
