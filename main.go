package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func readCallsigns() (r map[string]struct{}, err error) {
	f, err := os.Open("dbCallSigns.txt")
	if err != nil {
		return nil, err
	}
	r = make(map[string]struct{})
	lineReader := bufio.NewReader(f)
	for {
		line, _, err := lineReader.ReadLine()
		if err != nil {
			break
		}
		r[string(line)] = struct{}{}
	}
	return
}

func readCheckins(netLog string) (r []string, err error) {
	f, err := os.Open(netLog)
	if err != nil {
		return nil, err
	}
	lineReader := bufio.NewReader(f)
	for {
		line, _, err := lineReader.ReadLine()
		if err != nil {
			break
		}
		r = append(r, strings.ToUpper(string(line)))
	}
	return
}

func main() {
	netLogFile := flag.String("net-log", "net_log.txt", "File with net log")
	flag.Parse()

	callSigns, err := readCallsigns()
	if err != nil {
		fmt.Printf("Failed to read call signs: %v", err)
		os.Exit(1)
	}
	netLog, err := readCheckins(*netLogFile)
	if err != nil {
		fmt.Printf("Failed to read net log: %v", err)
		os.Exit(1)
	}
	confirmedMembers := make(map[string]struct{})
	for _, v := range netLog {
		if _, ok := callSigns[v]; ok {
			fmt.Printf("%v\n", v)
			confirmedMembers[v] = struct{}{}
		}
	}
	fmt.Printf("Confirmed members: %v\n", len(confirmedMembers))
}
