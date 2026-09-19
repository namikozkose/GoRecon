package network

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// worker fonksiyonunda değişiklik yok, aynı kalıyor
func worker(ports <-chan int, results chan<- int, hostname string, wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range ports {
		address := fmt.Sprintf("%s:%d", hostname, p)
		conn, err := net.DialTimeout("tcp", address, time.Second*1)
		if err != nil {
			continue
		}
		conn.Close()
		results <- p
	}
}

// maxPort yerine targetPorts []int (port listesi) alıyoruz
func StartConcurrentScan(hostname string, targetPorts []int, workerCount int) []int {
	ports := make(chan int, workerCount)
	results := make(chan int)
	var openPorts []int
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ports, results, hostname, &wg)
	}

	// 2. Adım Güncellendi: Artık sadece listedeki portları kanala gönderiyoruz
	go func() {
		for _, p := range targetPorts {
			ports <- p
		}
		close(ports)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for port := range results {
		openPorts = append(openPorts, port)
	}

	sort.Ints(openPorts)
	return openPorts
}
