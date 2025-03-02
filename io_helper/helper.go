package io_helper

import (
	"bufio"
	"os"
	"unicode/utf8"
)

func ReadPart(filePath string, n int) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buffer := make([]byte, n)
	bufReader := bufio.NewReader(f)
	numBytes, err := bufReader.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer[:numBytes], nil
}

func IsTextFile(filePath string) (bool, error) {
	content, err := ReadPart(filePath, 1024)
	if err != nil {
		return false, err
	}
	if utf8.Valid(content) {
		return true, nil
	} else {
		return false, nil
	}
}
