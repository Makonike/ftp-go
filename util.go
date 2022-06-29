package main

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
)

func stringInList(search string, data []string) bool {
	for _, d := range data {
		if search == d {
			return true
		}
	}
	return false
}

// parses a port arg and return an ip and port
// ex: 10,0,0,1,192,127 return 10.0.0.0:49279
func parsePortArgs(arg string) string {
	parts := strings.Split(arg, ",")
	ip := strings.Join(parts[:4], ".")
	p1, _ := strconv.Atoi(parts[4])
	p2, _ := strconv.Atoi(parts[5])
	port := p1*256 + p2
	return fmt.Sprintf("%s:%d", ip, port)
}

func stripDirectory(remote string) string {
	_, filename := path.Split(remote)
	return filename
}

func parseCommand(in string) (string, string, error) {
	var command, args string

	if len(in) < 3 {
		return command, args, errors.New(SyntaxErr)
	}
	resp := strings.SplitAfterN(in, " ", 2)

	switch {
	case len(resp) == 2:
		command = strings.TrimSpace(resp[0])
		args = strings.TrimSpace(resp[1])
	case len(resp) == 1:
		command = resp[0]
	}
	return command, args, nil
}
