package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"service-health-check/config"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Services struct {
	Stats []Service
	MQ    chan Message
}

type Service struct {
	Lock          *sync.RWMutex
	Name          string
	URL           string
	Status        int
	StatusHistory []ServiceStatus
}

type Message struct {
	URL   string
	index int
}

type ServiceStatus struct {
	Status int
	Time   time.Time
}

var timeout time.Duration
var client http.Client

func makeService(name, url string) Service {
	return Service{
		Name:          name,
		URL:           url,
		StatusHistory: []ServiceStatus{},
		Lock:          &sync.RWMutex{},
	}
}

// import -- scan a file and import the data
func (s *Services) Import(filePath string) error {

	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	i := 0

	for scanner.Scan() {
		i++
		fields := strings.Split(scanner.Text(), ",")
		if len(fields) < 2 {
			return errors.New("CSV: Column " + string(i) + " is not in format")
		}

		// Check CSV format
		if i == 1 {

			if fields[0] != "name" {
				return errors.New("CSV: cannot find field 'name'")
			}

			if fields[1] != "url" {
				return errors.New("CSV: cannot find field 'url'")
			}

			continue
		}
		s.Stats = append(s.Stats, makeService(fields[0], fields[1]))
	}

	if err = scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (s *Services) SpawnWorkerThreads(threads int) {

	for i := 0; i < threads; i++ {
		go Listener(s)
	}

	return
}

func (s *Services) Summary() string {
	fmt.Print("\033[H\033[2J")
	resp200 := 0
	resp404 := 0
	resp500 := 0
	remaining := 0
	var result string
	for _, v := range s.Stats {
		result += v.Name + "    \t|\t\t" + strconv.Itoa(v.Status) + "\t\t"
		for i := 0; i < len(v.StatusHistory); i++ {
			v.Lock.RLock()
			result += "|" + strconv.Itoa(v.StatusHistory[i].Status) + ""
			v.Lock.RUnlock()
		}
		result += "\n"
		switch int(v.Status / 100) {
		case 0:
			// Not pinged yet
			continue
		case 2:
			resp200++
			continue
		case 4:
			resp404++
			continue
		case 5:
			resp500++
			continue
		default:
			remaining++
			continue

		}
	}
	fmt.Printf("SERVICE MONITOR\n\n")
	fmt.Printf("Total Services: %d\n", len(s.Stats))
	fmt.Printf("Total Healthy Services(200): %d\n", resp200)
	fmt.Printf("Total Missing Services(400): %d\n", resp404)
	fmt.Printf("Total Broken Services(500): %d\n", resp500)
	fmt.Printf("Total Unresponsive Services: %d\n", remaining)
	fmt.Printf("---------------------------------------------------------\n")
	fmt.Printf("SERVICE \t|\t CURRENT STATUS \t|\t HISTORY(Last 6 Responses)\n")
	fmt.Printf(result)
	return result

}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func Listener(s *Services) {
	var message Message
	for {
		message = <-s.MQ
		resp, err := client.Get(message.URL)
		if err != nil {
			resp = &http.Response{
				StatusCode: 503,
			}
		} else {
			defer resp.Body.Close()
		}
		s.Stats[message.index].Lock.Lock()
		s.Stats[message.index].Status = resp.StatusCode
		s.Stats[message.index].StatusHistory = append(s.Stats[message.index].StatusHistory, ServiceStatus{Status: resp.StatusCode, Time: time.Now()})
		if len(s.Stats[message.index].StatusHistory) > 6 {
			s.Stats[message.index].StatusHistory = s.Stats[message.index].StatusHistory[1:]
		}
		s.Stats[message.index].Lock.Unlock()
	}
}

func main() {
	var S Services
	settings := config.Settings
	timeout = time.Duration(settings.Timeout) * time.Second

	err := S.Import(settings.Source)
	if err != nil {
		log.Fatal(err)
	}
	S.MQ = make(chan Message, 10*settings.MaxConcurrentThreads)
	S.SpawnWorkerThreads(settings.MaxConcurrentThreads)
	ticker := time.Tick(time.Duration(settings.HealthCheckFrequency) * time.Second)
	transport := http.Transport{
		Dial: dialTimeout,
	}

	client = http.Client{
		Transport: &transport,
	}
	go func() {
		for ; true; <-ticker {
			for i, v := range S.Stats {
				S.MQ <- Message{
					URL:   v.URL,
					index: i,
				}
			}
			S.Summary()
		}
	}()
	log.Fatal(http.ListenAndServe(settings.Port, nil))
}
