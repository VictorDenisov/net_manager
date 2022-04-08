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
	defer f.Close()
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

func readCheckins(netLog string) (r chan string, err error) {
	r = make(chan string)
	f, err := os.Open(netLog)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	lineReader := bufio.NewReader(f)
	go func() {
		defer close(r)
		for {
			line, _, err := lineReader.ReadLine()
			if err != nil {
				break
			}
			s := strings.ToUpper(string(line))
			r <- strings.TrimSpace(s)
		}
	}()
	return r, nil
}

func countCheckins(callSigns map[string]struct{}, netLog <-chan string) {
	confirmedMembers := make(map[string]struct{})
	sectionMembers := make(map[string]struct{})
	for v := range netLog {
		if _, ok := callSigns[v]; ok {
			if _, ok := confirmedMembers[v]; ok {
				fmt.Printf("%v = \n", v)
			} else {
				fmt.Printf("%v\n", v)
				sectionMembers[v] = struct{}{}
			}
			confirmedMembers[v] = struct{}{}
		} else {
			if v == "" {
				if len(sectionMembers) > 0 {
					fmt.Printf("Section count: %v\n", len(sectionMembers))
					sectionMembers = make(map[string]struct{})
				}
				fmt.Printf("\n")
			} else {
				fmt.Printf("%v - \n", v)
			}
		}
	}
	fmt.Printf("Section count: %v\n", len(sectionMembers))
	fmt.Printf("Confirmed members: %v\n", len(confirmedMembers))
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
	countCheckins(callSigns, netLog)
}
