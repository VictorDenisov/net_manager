package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
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
	lineReader := bufio.NewReader(f)
	go func() {
		defer f.Close()
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

type CheckinChans struct {
	dupCallSign     <-chan string
	memberCallSign  <-chan string
	sectionMarker   <-chan struct{}
	unknownCallSign <-chan string
}

func distributeCheckins(callSigns map[string]struct{}, netLog <-chan string) (r *CheckinChans) {
	confirmedMembers := make(map[string]struct{})
	sectionMembers := make(map[string]struct{})

	dupCallSign := make(chan string)
	memberCallSign := make(chan string)
	sectionMarker := make(chan struct{})
	unknownCallSign := make(chan string)

	r = &CheckinChans{
		dupCallSign:     dupCallSign,
		memberCallSign:  memberCallSign,
		sectionMarker:   sectionMarker,
		unknownCallSign: unknownCallSign,
	}
	go func() {
		for v := range netLog {
			if _, ok := callSigns[v]; ok {
				if _, ok := confirmedMembers[v]; ok {
					dupCallSign <- v
				} else {
					memberCallSign <- v
					sectionMembers[v] = struct{}{}
				}
				confirmedMembers[v] = struct{}{}
			} else {
				if v == "" {
					sectionMarker <- struct{}{}
				} else {
					unknownCallSign <- v
				}
			}
		}
		sectionMarker <- struct{}{}
		close(dupCallSign)
		close(memberCallSign)
		close(sectionMarker)
		close(unknownCallSign)
	}()
	return r
}

func countCheckins(callSigns map[string]struct{}, netLog <-chan string) {
	checkinChans := distributeCheckins(callSigns, netLog)

	sectionCount := 0
	totalCount := 0
loop:
	for {
		select {
		case dup, ok := <-checkinChans.dupCallSign:
			if !ok {
				break loop
			}
			fmt.Printf("%v = \n", dup)
		case member, ok := <-checkinChans.memberCallSign:
			if !ok {
				break loop
			}
			fmt.Printf("%v\n", member)
			sectionCount++
			totalCount++
		case _, ok := <-checkinChans.sectionMarker:
			if !ok {
				break loop
			}
			if sectionCount > 0 {
				fmt.Printf("Section count: %v\n", sectionCount)
				sectionCount = 0
			}
			fmt.Printf("\n")
		case unknown, ok := <-checkinChans.unknownCallSign:
			if !ok {
				break loop
			}
			fmt.Printf("%v - \n", unknown)
		}
	}

	fmt.Printf("Confirmed members: %v\n", totalCount)
}

func sortCheckins(callSigns map[string]struct{}, netLog <-chan string) {
	confirmedMembers := make(map[string]struct{})
	for v := range netLog {
		if _, ok := callSigns[v]; ok {
			confirmedMembers[v] = struct{}{}
		}
	}
	ls := make([]string, 0, len(confirmedMembers))
	for v := range confirmedMembers {
		ls = append(ls, v)
	}
	sort.Strings(ls)
	for _, v := range ls {
		fmt.Printf("%v\n", v)
	}
}

func main() {
	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to retrieve working directory: %v", err)
		os.Exit(1)
	}
	log.Tracef("Working directory: %v", workingDirectory)

	count := flag.Bool("count", true, "Count checkin numbers")
	sort := flag.Bool("sort", false, "Sort and print member checkins")
	timeSheet := flag.Bool("time-sheet", false, "Calculate time sheet for the specified month")
	monthPrefix := flag.String("month-prefix", "", "Month prefix in the format year-mo for drawing time sheet")
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
	if *sort {
		sortCheckins(callSigns, netLog)
	} else if *timeSheet {
		if !validMonthPrefixFormat(monthPrefix) {
			fmt.Printf("Month prefix is invalid")
			os.Exit(1)
		}
		drawTimeSheet(*monthPrefix, workingDirectory)
	} else if *count {
		countCheckins(callSigns, netLog)
	}
}

func validMonthPrefixFormat(monthPrefix *string) bool {
	// TODO improve month prefix validation
	return monthPrefix != nil
}

func drawTimeSheet(monthPrefix string, workingDir string) error {
	list, err := filepath.Glob(monthPrefix + "*")
	if err != nil {
		return err
	}
	for _, f := range list {
		checkins, err := readCheckins(f)
		if err != nil {
			return err
		}
		fmt.Printf("Checkins len: %v\n", checkins)
	}
	return nil
}
