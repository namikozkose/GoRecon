package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/namikozkose/GoRecon/internal/network"
	"github.com/spf13/cobra"
)

var (
	targetIP   string
	ports      string
	workers    int
	outputFile string
)

func parsePorts(portStr string) []int {
	var result []int
	parts := strings.Split(portStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				continue
			}
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])
			if err1 == nil && err2 == nil && start <= end {
				for i := start; i <= end; i++ {
					result = append(result, i)
				}
			}
		} else {
			p, err := strconv.Atoi(part)
			if err == nil {
				result = append(result, p)
			}
		}
	}
	return result
}

var rootCmd = &cobra.Command{
	Use:   "scanner",
	Short: "Yüksek Hızlı Dağıtık TCP Port Tarayıcı",
	Run: func(cmd *cobra.Command, args []string) {

		targetPorts := parsePorts(ports)
		if len(targetPorts) == 0 {
			fmt.Println("[-] HATA: Geçerli bir port aralığı girilmedi!")
			return
		}

		targets, err := network.GetTargets(targetIP)
		if err != nil {
			fmt.Printf("[-] HATA: Hedef ayrıştırılamadı: %v\n", err)
			return
		}

		fmt.Printf("\n[+] HEDEF KAPSAMI: %d Adet IP Adresi\n", len(targets))
		fmt.Printf("[+] GÜÇ: %d Worker | CEPHANE: %d Port (Her IP için)\n", workers, len(targetPorts))
		fmt.Println("[~] AĞ TARAMASI BAŞLATILDI! Lütfen bekleyin...")

		allResults := make(map[string][]network.PortResult)

		// YENİ SİLAH: Host seviyesinde eşzamanlılık için WaitGroup ve Mutex
		var wg sync.WaitGroup
		var mu sync.Mutex

		startTime := time.Now()

		// Artık IP'leri sırayla beklemek yok, hepsine aynı anda saldırıyoruz!
		for _, ip := range targets {
			wg.Add(1)

			// Her bir IP için ayrı bir Goroutine başlatıyoruz
			go func(targetIP string) {
				defer wg.Done()

				// O IP için port taramasını başlat
				openPorts := network.StartConcurrentScan(targetIP, targetPorts, workers)

				// Eğer açık port bulunduysa, sonuçları güvenle kaydet
				if len(openPorts) > 0 {
					mu.Lock() // Aynı anda haritaya yazmayı engellemek için kilitliyoruz (Thread-Safe)
					allResults[targetIP] = openPorts
					mu.Unlock() // Kilidi aç
				}
			}(ip)
		}

		// Tüm IP Goroutine'lerinin işini bitirmesini bekle
		wg.Wait()

		duration := time.Since(startTime)
		fmt.Printf("\n[!] TÜM AĞ %v İÇİNDE YOK EDİLDİ.\n", duration)
		fmt.Println("---------------------------------------------------")

		for ip, resList := range allResults {
			fmt.Printf("[*] %s üzerindeki açık portlar:\n", ip)
			for _, result := range resList {
				cleanBanner := strings.TrimSpace(result.Banner)
				if cleanBanner != "" {
					lines := strings.Split(cleanBanner, "\n")
					fmt.Printf(" -> %d/tcp \tAÇIK \t| %s\n", result.Port, strings.TrimSpace(lines[0]))
				} else {
					fmt.Printf(" -> %d/tcp \tAÇIK \t| (Filtreli / Yanıt Yok)\n", result.Port)
				}
			}
			fmt.Println("---------------------------------------------------")
		}

		if outputFile != "" {
			jsonData, err := json.MarshalIndent(allResults, "", "  ")
			if err != nil {
				fmt.Printf("[-] JSON hatası: %v\n", err)
				return
			}
			os.WriteFile(outputFile, jsonData, 0644)
			fmt.Printf("[+] Rapor başarıyla '%s' dosyasına kaydedildi!\n", outputFile)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&targetIP, "target", "t", "", "Hedef IP, Domain veya CIDR (Zorunlu)")
	rootCmd.Flags().StringVarP(&ports, "ports", "p", "1-1024", "Taranacak portlar (Örn: 80,443 veya 1-100)")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 100, "Goroutine Sayısı")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Sonuçları JSON formatında kaydet")
	rootCmd.MarkFlagRequired("target")
}
