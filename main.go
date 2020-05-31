package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var host string

const (
	hostFlag      = "host"
	hostFlagShort = "h"
	hostUsage     = "Hostnames to update (comma separated), not domain"
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

func init() {
	log.SetOutput(os.Stdout)
	envHost := os.Getenv("HOST")
	envDomain := os.Getenv("DOMAIN")
	envPassword := os.Getenv("PASSWORD")
	var envInterval int
	var err error
	if envInterval, err = strconv.Atoi(os.Getenv("INTERVAL")); err != nil {
		envInterval = intervalDefault
	}
	flag.StringVar(&host, hostFlag, envHost, hostUsage)
	flag.StringVar(&host, hostFlagShort, envHost, hostUsage+" (shortcut)")
	flag.StringVar(&domain, domainFlag, envDomain, domainUsage)
	flag.StringVar(&domain, domainFlagShort, envDomain, domainUsage+" (shortcut)")
	flag.StringVar(&password, passwordFlag, envPassword, passwordUsage)
	flag.StringVar(&password, passwordFlagShort, envPassword, passwordUsage+" (shortcut)")
	flag.IntVar(&interval, intervalFlag, envInterval, intervalUsage)
	flag.IntVar(&interval, intervalFlagShort, envInterval, intervalUsage+"( shortcut)")
	flag.Parse()
	if host == "" || domain == "" || password == "" {
		flag.PrintDefaults()
		panic("You must specify host, domain and password")
	}
	if interval <= 0 {
		panic("Interval must be greater than 0")
	}
	log.Printf("Host=%s; Domain=%s; Interval=%d\n", host, domain, interval)
}

func main() {
	hosts := strings.Split(host, ",")
	urls := make(map[string]string)
	for _, hostname := range hosts {
		urls[hostname] = fmt.Sprintf("https://dynamicdns.park-your-domain.com/update?host=%s&domain=%s&password=%s", hostname, domain, password)
	}
	for {
		for hostname, url := range urls {
			b, err := fetch(url)
			if err != nil {
				log.Fatal(err)
			}
			res := result{}
			err = xml.Unmarshal(b, &res)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Host: %s - %s\n", hostname, res)
		}
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
