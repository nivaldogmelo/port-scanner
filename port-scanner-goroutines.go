package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const maxPort = 65535
const timeout = 300 * time.Millisecond

type Job struct {
	server string
	port   int
}

var available []int

func createWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	jobs := make(chan Job, noOfWorkers)

	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go worker(&wg, jobs)
	}

	for i := 1; i <= maxPort; i++ {
		job := Job{server: os.Args[1], port: i}
		jobs <- job
	}

	close(jobs)
	wg.Wait()
}

func worker(wg *sync.WaitGroup, jobs <-chan Job) {
	defer wg.Done()

	for job := range jobs {
		ip := fmt.Sprintf("%s:%d", job.server, job.port)

		conn, err := net.DialTimeout("tcp", ip, timeout)
		if err == nil {
			conn.Close()
			available = append(available, job.port)
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./portscanner <hostname>")
		os.Exit(1)
	}

	fmt.Println("Checking for available ports...")

	noOfWorkers := 100
	createWorkerPool(noOfWorkers)

	fmt.Println("Ports available:", available)
}
