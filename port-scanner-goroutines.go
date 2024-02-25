package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const timeout = 300 * time.Millisecond

type Job struct {
	server string
	port   int
}

var available []int
var verbose int
var numThreads int

func scanPorts(jobs []Job, minPacketRate int, delay time.Duration, wg *sync.WaitGroup, jobsChan chan Job) {
	packetRate := time.Duration(1000000000 / minPacketRate)

	for job := range jobsChan {
		ip := fmt.Sprintf("%s:%d", job.server, job.port)

		conn, err := net.DialTimeout("tcp", ip, timeout)
		if err == nil {
			conn.Close()
			available = append(available, job.port)
			if verbose >= 3 {
				fmt.Printf("Port %d is open\n", job.port)
			}
		}
		
		time.Sleep(packetRate)
		
		time.Sleep(delay)

		wg.Done()
	}
}

func parsePorts(portArg string, scanAll bool) ([]int, error) {
	if scanAll {
		var portRange []int
		for i := 1; i <= 65535; i++ {
			portRange = append(portRange, i)
		}
		return portRange, nil
	}

	ports := strings.Split(portArg, ",")
	var portRange []int

	for _, port := range ports {
		if strings.Contains(port, "-") {
			rangeParts := strings.Split(port, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid port range format: %s", port)
			}

			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid start of port range: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid end of port range: %s", rangeParts[1])
			}

			for i := start; i <= end; i++ {
				portRange = append(portRange, i)
			}
		} else {
			portNum, err := strconv.Atoi(port)
			if err != nil {
				return nil, fmt.Errorf("invalid port number: %s", port)
			}
			portRange = append(portRange, portNum)
		}
	}

	return portRange, nil
}

func main() {
	minPacketRate := 10000 // Default minimum packet rate
	var portArg string
	var delay time.Duration

	for i, arg := range os.Args {
		if strings.HasPrefix(arg, "--min-packet-rate=") {
			var err error
			minPacketRate, err = strconv.Atoi(arg[len("--min-packet-rate="):])
			if err != nil {
				fmt.Println("Error: invalid minimum packet rate")
				os.Exit(1)
			}
		} else if strings.HasPrefix(arg, "--delay=") {
			var err error
			delay, err = time.ParseDuration(arg[len("--delay="):])
			if err != nil {
				fmt.Println("Error: invalid delay duration")
				os.Exit(1)
			}
		} else if arg == "-p" && i+1 < len(os.Args) {
			portArg = os.Args[i+1]
		} else if arg == "-vvv" { // If vvv used then verbose will be set to 3
			verbose = 3
		} else if strings.HasPrefix(arg, "--threads=") {
			var err error
			numThreads, err = strconv.Atoi(arg[len("--threads="):])
			if err != nil || numThreads <= 0 {
				fmt.Println("Error: invalid number of threads")
				os.Exit(1)
			}
		}
	}

	if numThreads == 0 {
		numThreads = runtime.NumCPU()
	}

	hostname := os.Args[1]

	// If no ports are specified, scan all ports
	if portArg == "" {
		portArg = "1-65535"
	}

	
	ports, err := parsePorts(portArg, true)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("Checking for available ports...")

	jobs := make([]Job, len(ports))
	for i, port := range ports {
		jobs[i] = Job{server: hostname, port: port}
	}

	var wg sync.WaitGroup
	jobsChan := make(chan Job, numThreads)

	for i := 0; i < numThreads; i++ {
		go scanPorts(jobs, minPacketRate, delay, &wg, jobsChan)
	}

	for _, job := range jobs {
		wg.Add(1)
		jobsChan <- job
	}

	wg.Wait()
	close(jobsChan)

	if verbose < 3 {
		fmt.Println("Ports available:", available)
	}
}
