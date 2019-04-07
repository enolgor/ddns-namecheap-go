package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var host string

const (
	hostFlag      = "host"
	hostFlagShort = "h"
	hostUsage     = "Hostname to update, not domain"
)

var domain string

const (
	domainFlag      = "domain"
	domainFlagShort = "d"
	domainUsage     = "Domain name"
)

var password string

const (
	passwordFlag      = "password"
	passwordFlagShort = "p"
	passwordUsage     = "Dynamic dns password of domain"
)

var interval int

const (
	intervalFlag      = "interval"
	intervalFlagShort = "i"
	intervalUsage     = "Update interval in seconds"
	intervalDefault   = 5 * 60
)

var logfile string

const (
	logfileFlag      = "log"
	logfileFlagShort = "l"
	logfileUsage     = "Log file path"
	logfileDefault   = "/ddns-namecheap-go.log"
)

func init() {
	flag.StringVar(&host, hostFlag, "", hostUsage)
	flag.StringVar(&host, hostFlagShort, "", hostUsage+" (shortcut)")
	flag.StringVar(&domain, domainFlag, "", domainUsage)
	flag.StringVar(&domain, domainFlagShort, "", domainUsage+" (shortcut)")
	flag.StringVar(&password, passwordFlag, "", passwordUsage)
	flag.StringVar(&password, passwordFlagShort, "", passwordUsage+" (shortcut)")
	flag.IntVar(&interval, intervalFlag, intervalDefault, intervalUsage)
	flag.IntVar(&interval, intervalFlagShort, intervalDefault, intervalUsage+"( shortcut)")
	flag.StringVar(&logfile, logfileFlag, logfileDefault, logfileUsage)
	flag.StringVar(&logfile, logfileFlagShort, logfileDefault, logfileUsage+" (shortcut)")
	flag.Parse()
	if host == "" || domain == "" || password == "" {
		flag.PrintDefaults()
		panic("You must specify host, domain and password")
	}
	if interval <= 0 {
		panic("Interval must be greater than 0")
	}
}

func main() {
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	url := fmt.Sprintf("https://dynamicdns.park-your-domain.com/update?host=%s&domain=%s&password=%s", host, domain, password)
	for {
		b, err := fetch(url)
		if err != nil {
			log.Fatal(err)
		}
		res := result{}
		err = xml.Unmarshal(b, &res)
		if err != nil {
			log.Fatal(err)
		}
		log.Print(res)
		time.Sleep(time.Duration(interval) * time.Second)
	}

}

type result struct {
	XMLName                 xml.Name `xml:"interface-response"`
	Command, Language, IP   string
	ErrCount, ResponseCount int
	Done                    bool
}

func (r result) String() string {
	return fmt.Sprintf("Command: %s, Language: %s, IP: %s, ErrCount: %d, ResponseCount: %d, Done: %t", r.Command, r.Language, r.IP, r.ErrCount, r.ResponseCount, r.Done)
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}
