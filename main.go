package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

const (
	CityResponsiblityScheduleFileName = "city_responsibility_schedule.txt"
	callsignDB                        = "ContactListByName.csv"
	NetcontrolScheduleFileName        = "netcontrol_schedule.txt"
)

func main() {
	count := flag.Bool("count", false, "Count checkin numbers")
	sort := flag.Bool("sort", false, "Sort and print member checkins")
	timeSheet := flag.Bool("time-sheet", false, "Calculate time sheet for the specified month")
	sendEmails := flag.Bool("send-emails", false, "Check if it's time to send emails")
	monthPrefix := flag.String("month-prefix", "", "Month prefix in the format year-mo for drawing time sheet")
	netLogFile := flag.String("net-log", "net_log.txt", "File with net log")
	logLevelString := flag.String("debug-level", "info", "Debug level of the application")
	flag.Parse()

	config := readConfig()

	fmt.Printf("Parsed config: %v\n", config)

	logLevel, err := log.ParseLevel(*logLevelString)
	if err != nil {
		fmt.Printf("Failed to parse log level: %v", *logLevelString)
	}
	log.SetLevel(logLevel)

	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to retrieve working directory: %v", err)
		os.Exit(1)
	}
	log.Tracef("Working directory: %v", workingDirectory)

	log.Tracef("Parsed command line args:")
	log.Tracef("Count: %v", *count)
	log.Tracef("Sort: %v", *sort)
	log.Tracef("Time Sheet: %v", timeSheet)

	callSigns, err := readCallsignDB()
	if err != nil {
		fmt.Printf("Failed to read call signs: %v", err)
		os.Exit(1)
	}
	if *sort || *count {
		netLog, err := readCheckins(*netLogFile)
		if err != nil {
			fmt.Printf("Failed to read net log: %v", err)
			os.Exit(1)
		}
		if *sort {
			sortCheckins(callSigns, netLog)
		} else if *count {
			countCheckins(callSigns, netLog)
		}
	} else if *timeSheet {
		if !validMonthPrefixFormat(monthPrefix) {
			fmt.Printf("Month prefix is invalid")
			os.Exit(1)
		}
		drawTimeSheet(*monthPrefix, workingDirectory, callSigns)
	} else if *sendEmails {
		log.Trace("Checking if emails should be sent")
		dispatchEmails(callSigns, config)
	}
}

type CityResponsibilityRecord struct {
	Date time.Time
	City string
}

func weekdayNumber(t time.Time) int {
	return (t.Day()-1)/7 + 1
}

func dispatchEmails(callsignDB map[string]Member, config *Config) {
	ncSchedule, err := readNetcontrolSchedule()
	if err != nil {
		fmt.Printf("Failed to parse net control schedule: %v\n", err)
		os.Exit(1)
	}

	callForSignups(ncSchedule, config)

	now := time.Now()
	if now.Weekday() == time.Sunday {
		err := notifyNetControl(callsignDB, config, ncSchedule)
		if err != nil {
			fmt.Printf("Failed to notify net control: %v\n", err)
		}
	}
	if now.Day() == 1 {
		log.Trace("Sending time sheet\n")
		sendReport(config, callsignDB)
	}
}

func sendReport(config *Config, callsigns map[string]Member) {
	now := time.Now()
	previousMonthTime := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	monthPrefix := fmt.Sprintf("%d-%02d", previousMonthTime.Year(), previousMonthTime.Month())
	netString, netHours, err := drawTimeSheetString(monthPrefix, config.NetDir, callsigns)
	hospitalHours, err := hospitalHoursCount(monthPrefix, config.HospitalDir, callsigns)
	log.Tracef("Report to be sent: \n%v\n, %v\n", netString, err)
	log.Tracef("Hospital Net: %0.3f, %v\n", hospitalHours, err)
	log.Tracef("Total Hours: %0.3f, %v\n", hospitalHours+netHours, err)

	d := gomail.NewDialer(config.Station.Mail.SmtpHost, config.Station.Mail.Port, config.Station.Mail.Email, config.Station.Mail.Password)

	m := gomail.NewMessage()
	m.SetHeader("From", config.Station.Mail.Email)
	m.SetHeader("To", config.TimeReport.MainMail)
	if config.TimeReport.CcMail != "" {
		m.SetHeader("Cc", config.TimeReport.CcMail)
	}
	monthString := previousMonthTime.Format("Jan 2006")
	m.SetHeader("Bcc", config.Station.Mail.Email)
	m.SetHeader("Subject", fmt.Sprintf("[SJ-RACES] Net report for %v", monthString))
	bodyText := ""
	bodyText += "Hi folks,\n\n"
	bodyText += fmt.Sprintf("Here is net control statistics for %v:\n\n", monthString)
	bodyText += netString
	bodyText += "\n"
	bodyText += fmt.Sprintf("Hospital Net: %0.3f\n\n", hospitalHours)
	bodyText += fmt.Sprintf("Total Hours: %0.3f\n", hospitalHours+netHours)
	bodyText += fmt.Sprintf("\n\n%v", config.Station.Signature)

	m.SetBody("text/plain", bodyText)

	if err := d.DialAndSend(m); err != nil {
		log.Errorf("Failed to send email: %v", err)
		os.Exit(1)
	}
}

func callForSignups(ncSchedule []NetcontrolScheduleRecord, config *Config) {
	citySchedule, err := readCityResponsibilitySchedule()
	if err != nil {
		fmt.Printf("Failed to read city responsibility schedule: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed city responsibility schedule: %v\n", citySchedule)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var nextMonthStart time.Time
	if now.Day() < 15 {
		nextMonthStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	} else {
		nextMonthStart = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	}
	distance := nextMonthStart.Sub(today)

	fmt.Printf("Distance %v\n", distance)
	fmt.Printf("NextMonthStart %v\n", nextMonthStart)
	if !monthCityComplete(nextMonthStart, citySchedule) {
		fmt.Printf("Next month city schedule is incomplete. Add more records.\n")
		os.Exit(1)
	}
	monthFull, ms := monthSchedule(nextMonthStart, ncSchedule, citySchedule)
	fmt.Printf("Month schedule: %v\n", ms)
	fmt.Printf("Month full: %v\n", monthFull)
	if distance < time.Hour*24*10 && !monthFull {
		fmt.Printf("Hi,\n")
		fmt.Printf("Net control positions are open.\n")
		fmt.Printf("Here is the schedule right now:\n")
		for _, nc := range ms {
			fmt.Printf("%v\t%v\t%v\n", nc.Date.Format("1/2/2006"), nc.City, nc.Callsign)
		}

		d := gomail.NewDialer(config.Station.Mail.SmtpHost, config.Station.Mail.Port, config.Station.Mail.Email, config.Station.Mail.Password)

		m := gomail.NewMessage()
		m.SetHeader("From", config.Station.Mail.Email)
		m.SetHeader("To", "Main@SJ-RACES.groups.io")
		m.SetHeader("Subject", fmt.Sprintf("[SJ-RACES] SJ RACES Net Control for %v", nextMonthStart.Format("Jan 2006")))
		bodyText := ""
		bodyText += "Hi,\n\n"
		bodyText += "Net control positions are open.\n\n"
		bodyText += "Here is the schedule right now:\n"
		for _, nc := range ms {
			bodyText += fmt.Sprintf("%v\t%v\t%v\n", nc.Date.Format("1/2/2006"), nc.City, nc.Callsign)
		}
		m.SetBody("text/plain", bodyText)

		if err := d.DialAndSend(m); err != nil {
			fmt.Printf("Failed to send email: %v", err)
			os.Exit(1)
		}

	}

}

func monthCityComplete(monthStart time.Time, citySchedule []CityResponsibilityRecord) bool {
	// TODO check if city responsiblity schedule covers the whole month
	return true
}

type ScheduleRecord struct {
	Date     time.Time
	City     string
	Callsign string
}

func monthSchedule(monthStart time.Time, ncSchedule []NetcontrolScheduleRecord, citySchedule []CityResponsibilityRecord) (monthFull bool, schedule []ScheduleRecord) {
	ncMonth := make([]NetcontrolScheduleRecord, 0)
	for _, cr := range ncSchedule {
		if equalByMonth(cr.Date, monthStart) {
			ncMonth = append(ncMonth, cr)
		}
	}
	schedule = make([]ScheduleRecord, 0)
	monthFull = true
	for _, cr := range citySchedule {
		if equalByMonth(cr.Date, monthStart) {
			sr := ScheduleRecord{Date: cr.Date, City: cr.City}
			dateMatch := false
			for _, nr := range ncMonth {
				if equalByDate(cr.Date, nr.Date) {
					sr.Callsign = nr.Callsign
					dateMatch = true
				}
			}
			if !dateMatch {
				monthFull = false
			}
			schedule = append(schedule, sr)
		}
	}

	return monthFull, schedule
}

type NetcontrolScheduleRecord struct {
	Date     time.Time
	Callsign string
}

func equalByMonth(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month()
}

func equalByDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func notifyNetControl(callsignDB map[string]Member, config *Config, netcontrolSchedule []NetcontrolScheduleRecord) error {
	now := time.Now()
	inTwoDaysFromNow := now.Add(48 * time.Hour)
	var upcomingNc NetcontrolScheduleRecord
	for _, ncRecord := range netcontrolSchedule {
		if equalByDate(inTwoDaysFromNow, ncRecord.Date) {
			upcomingNc = ncRecord
			break
		}
	}

	ncCallsign := strings.ToUpper(upcomingNc.Callsign)

	fmt.Printf("Chosen nc record: %v\n", upcomingNc)
	ncEmail := callsignDB[ncCallsign].Email
	if ncEmail == "" {
		return fmt.Errorf("Net control %v has empty email", strings.ToUpper(upcomingNc.Callsign))
	}
	fmt.Printf("Sending email to: %v\n", ncEmail)

	d := gomail.NewDialer(config.Station.Mail.SmtpHost, config.Station.Mail.Port, config.Station.Mail.Email, config.Station.Mail.Password)

	m := gomail.NewMessage()
	m.SetHeader("From", config.Station.Mail.Email)
	m.SetHeader("To", ncEmail)
	dateString := upcomingNc.Date.Format("1/2/2006")
	m.SetHeader("Bcc", config.Station.Mail.Email)
	m.SetHeader("Subject", fmt.Sprintf("Net control %v", dateString))
	m.SetBody("text/plain", fmt.Sprintf("Hi %s,\n\nThank you for volunteering. Could you please confirm that you are still comfortable running the net on %v\n\nThanks, Victor.", callsignDB[ncCallsign].Name, dateString))

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("Failed to send email: %w", err)
	}

	return nil
}

func readNetcontrolSchedule() ([]NetcontrolScheduleRecord, error) {
	f, err := openFile(NetcontrolScheduleFileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to open netcontrol schedule: %w", err)
	}
	defer f.Close()

	records := make([]NetcontrolScheduleRecord, 0)
	lineReader := bufio.NewReader(f)
	for {
		line, _, err := lineReader.ReadLine()
		if err != nil {
			break
		}
		tokens := bytes.Split(line, []byte("\t"))
		date, err := time.Parse("1/2/2006", string(bytes.TrimSpace(tokens[0])))
		if err != nil {
			return nil, fmt.Errorf("Failed to parse netcontrol schedule: %w", err)
		}
		callsign := string(bytes.TrimSpace(tokens[1]))
		records = append(records, NetcontrolScheduleRecord{date, callsign})
	}
	return records, nil
}

func readCityResponsibilitySchedule() (records []CityResponsibilityRecord, err error) {
	f, err := openFile(CityResponsiblityScheduleFileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to open city responsibility schedule: %w", err)
	}
	defer f.Close()
	records = make([]CityResponsibilityRecord, 0)
	lineReader := bufio.NewReader(f)
	for {
		line, _, err := lineReader.ReadLine()
		if err != nil {
			break
		}
		whiteSpace := bytes.IndexAny(line, "\t ")
		time, err := time.Parse("1/2/2006", string(line[0:whiteSpace]))
		if err != nil {
			return nil, fmt.Errorf("Failed to parse city responsibility schedule: %w", err)
		}
		cityName := string(bytes.TrimSpace(line[whiteSpace:]))
		records = append(records, CityResponsibilityRecord{time, cityName})
	}
	return
}

type Member struct {
	Name     string
	Callsign string
	Email    string
}

func readCallsignDB() (r map[string]Member, err error) {
	r = make(map[string]Member)
	f, err := openFile(callsignDB)
	if err != nil {
		return nil, fmt.Errorf("Failed to open call signdb: %v %w", callsignDB, err)
	}
	lineReader := bufio.NewReader(f)
	for {
		line, _, err := lineReader.ReadLine()
		if err != nil {
			break
		}
		fields := strings.Split(string(line), ",")
		name := strings.TrimSpace(fields[1])
		callsign := strings.TrimSpace(fields[2])
		email := strings.TrimSpace(fields[7])

		r[callsign] = Member{name, callsign, email}
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

func distributeCheckins(callSigns map[string]Member, netLog <-chan string) (r *CheckinChans) {
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

func countCheckins(callSigns map[string]Member, netLog <-chan string) {
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

func sortCheckins(callSigns map[string]Member, netLog <-chan string) {
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

func validMonthPrefixFormat(monthPrefix *string) bool {
	fmt.Printf("%v\n", *monthPrefix)
	if monthPrefix == nil {
		return false
	}
	if len(*monthPrefix) != 4 && len(*monthPrefix) != 7 {
		return false
	}
	for i := 0; i < 4; i++ {
		if !('0' <= (*monthPrefix)[i] && (*monthPrefix)[i] <= '9') {
			return false
		}
	}
	if len(*monthPrefix) == 7 {
		if !('0' <= (*monthPrefix)[5] && (*monthPrefix)[5] <= '9') {
			return false
		}
		if !('0' <= (*monthPrefix)[6] && (*monthPrefix)[6] <= '9') {
			return false
		}
	}
	return true
}

func drawTimeSheet(monthPrefix string, logDirectory string, callSigns map[string]Member) error {
	s, hours, err := drawTimeSheetString(monthPrefix, logDirectory, callSigns)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", s)
	fmt.Printf("Hours: %v\n", hours)
	return nil
}

func drawTimeSheetString(monthPrefix string, logDirectory string, callSigns map[string]Member) (string, float64, error) {
	var sb strings.Builder
	list, err := filepath.Glob(filepath.Join(logDirectory, monthPrefix) + "*")
	if err != nil {
		return "", 0, err
	}
	var totalHours float64
	for _, f := range list {
		checkins, err := readCheckins(f)
		if err != nil {
			return "", 0, err
		}
		totalCount := totalCheckins(callSigns, checkins)
		hours := float64(totalCount)/3 + 0.5 + 0.25
		fmt.Fprintf(&sb, "%v:\t%d\t%0.3f\t%0.3f\t%0.3f\t%0.3f\n", filepath.Base(f), totalCount, hours, 0.5, 0.25, hours)
		totalHours += hours
	}
	fmt.Fprintf(&sb, "Total hours: %0.3f\n", totalHours)
	return sb.String(), totalHours, nil
}

func hospitalHoursCount(monthPrefix string, logDirectory string, callSigns map[string]Member) (float64, error) {
	var totalHours float64
	list, err := filepath.Glob(filepath.Join(logDirectory, monthPrefix) + "*")
	if err != nil {
		return 0, err
	}
	log.Tracef("Doing hospital count")
	for i, f := range list {
		log.Tracef("Processing file: %v", f)
		if i > 0 {
			log.Errorf("More than one hospital net log")
			break
		}
		checkins, err := readCheckins(f)
		if err != nil {
			return 0, err
		}
		totalCount := totalCheckins(callSigns, checkins)
		totalHours = float64(totalCount)*0.5 + 0.25
	}
	return totalHours, nil
}

func totalCheckins(callSigns map[string]Member, netLog <-chan string) (r int) {
	checkinChans := distributeCheckins(callSigns, netLog)
loop:
	for {
		select {
		case _, ok := <-checkinChans.dupCallSign:
			if !ok {
				break loop
			}
		case _, ok := <-checkinChans.memberCallSign:
			if !ok {
				break loop
			}
			r++
		case _, ok := <-checkinChans.sectionMarker:
			if !ok {
				break loop
			}
		case _, ok := <-checkinChans.unknownCallSign:
			if !ok {
				break loop
			}
		}
	}
	return r
}
