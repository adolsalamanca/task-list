package main

import "strconv"

type identifier int64

func NewIdentifier(idString string) (identifier, error) {
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return -1, err
	}
	return identifier(id), nil
}
