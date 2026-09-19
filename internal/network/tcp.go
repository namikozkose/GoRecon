package network

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// PortResult, hem port numarasını hem de oradan dönen servis bilgisini (banner) tutacak
type PortResult struct {
	Port   int
	Banner string
}

func worker(ports <-chan int, results chan<- PortResult, hostname string, wg *sync.WaitGroup) {
	defer wg.Done()

	for p := range ports {
		address := fmt.Sprintf("%s:%d", hostname, p)
		conn, err := net.DialTimeout("tcp", address, time.Second*1)
		if err != nil {
			continue // Kapı hiç açılmadıysa geç
		}

		// BANNER GRABBING MANTIĞI:
		conn.SetReadDeadline(time.Now().Add(time.Second * 2))
		buffer := make([]byte, 1024)
		n, _ := conn.Read(buffer) // Karşıdan gelen veriyi oku

		banner := ""
		if n > 0 {
			banner = string(buffer[:n]) // Konuşkan servisler (SSH, FTP) buraya girer
		} else {
			// ACTIVE PROBING: Karşı taraf sessiz kaldıysa utangaç bir HTTP servisi olabilir, onu dürtelim:
			conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))

			// Cevabı dinlemek için tekrar okuma yapalım (1 saniye yeterli)
			conn.SetReadDeadline(time.Now().Add(time.Second * 1))
			n2, _ := conn.Read(buffer)
			if n2 > 0 {
				banner = string(buffer[:n2]) // Web sunucusunun yanıtını alıyoruz
			}
		}

		conn.Close()

		// Sonucu kanala (PortResult struct'ı olarak) gönder
		results <- PortResult{Port: p, Banner: banner}
	}
}

func StartConcurrentScan(hostname string, targetPorts []int, workerCount int) []PortResult {
	ports := make(chan int, workerCount)
	results := make(chan PortResult)
	var openPorts []PortResult
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ports, results, hostname, &wg)
	}

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

	for res := range results {
		openPorts = append(openPorts, res)
	}

	// Port numarasına göre listeyi küçükten büyüğe sırala
	sort.Slice(openPorts, func(i, j int) bool {
		return openPorts[i].Port < openPorts[j].Port
	})

	return openPorts
}
