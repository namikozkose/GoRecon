package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/namikozkose/GoRecon/internal/network"
	"github.com/spf13/cobra"
)

var (
	targetIP   string
	ports      string
	workers    int
	outputFile string // Yeni silahımız: JSON Export
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

		fmt.Printf("\n[+] HEDEF KİLİTLENDİ: %s\n", targetIP)
		fmt.Printf("[+] GÜÇ: %d Worker | CEPHANE: %d Port\n", workers, len(targetPorts))
		fmt.Println("[~] Tarama başlatılıyor, ağ trafiği izleniyor...")

		// KRONOMETREYİ BAŞLAT
		startTime := time.Now()

		openPorts := network.StartConcurrentScan(targetIP, targetPorts, workers)

		// KRONOMETREYİ DURDUR
		duration := time.Since(startTime)

		fmt.Printf("\n[!] TARAMA %v İÇİNDE YOK EDİLDİ.\n", duration)
		fmt.Println("---------------------------------------------------")

		for _, result := range openPorts {
			cleanBanner := strings.TrimSpace(result.Banner)
			if cleanBanner != "" {
				lines := strings.Split(cleanBanner, "\n")
				fmt.Printf(" -> %d/tcp \tAÇIK \t| %s\n", result.Port, strings.TrimSpace(lines[0]))
			} else {
				fmt.Printf(" -> %d/tcp \tAÇIK \t| (Filtreli / Yanıt Yok)\n", result.Port)
			}
		}
		fmt.Println("---------------------------------------------------")

		// EĞER KULLANICI -o FLAG'İ GİRDİYSE JSON'A DÖK
		if outputFile != "" {
			fmt.Printf("[+] JSON Raporu hazırlanıyor: %s\n", outputFile)

			// JSON formatını güzelleştirerek (Indent) oluştur
			jsonData, err := json.MarshalIndent(openPorts, "", "  ")
			if err != nil {
				fmt.Printf("[-] JSON oluşturulurken hata: %v\n", err)
				return
			}

			// Dosyaya yazdır
			err = os.WriteFile(outputFile, jsonData, 0644)
			if err != nil {
				fmt.Printf("[-] Dosya yazılamadı: %v\n", err)
				return
			}
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
	rootCmd.Flags().StringVarP(&targetIP, "target", "t", "", "Hedef IP veya Hostname (Zorunlu)")
	rootCmd.Flags().StringVarP(&ports, "ports", "p", "1-1024", "Taranacak portlar (Örn: 80,443 veya 1-100)")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 100, "Goroutine Sayısı (Eşzamanlı iş parçacığı)")

	// YENİ FLAG: Output
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Sonuçları JSON formatında kaydet (Örn: scan_result.json)")

	rootCmd.MarkFlagRequired("target")
}
