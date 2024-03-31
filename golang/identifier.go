package main

import "strconv"

type identifier string

func NewIdentifier(idString string) (identifier, error) {
	_, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return "", err
	}
	return identifier(idString), nil
}

func NewCustomIdentifier(customId string) identifier {
	return identifier(customId)
}
