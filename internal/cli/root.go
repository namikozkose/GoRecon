package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/namikozkose/GoRecon/internal/network"
	"github.com/spf13/cobra"
)

var (
	targetIP string
	ports    string
	workers  int
)

// parsePorts, terminalden gelen "80,443" veya "1-100" gibi metinleri tam sayı listesine çevirir
func parsePorts(portStr string) []int {
	var result []int

	// Virgülle ayrılmış değerleri parçala
	parts := strings.Split(portStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Eğer içinde "-" varsa (Örn: 1-100) aralığı hesapla
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				continue
			}

			// Metni tam sayıya (Integer) çevir
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])

			if err1 == nil && err2 == nil && start <= end {
				for i := start; i <= end; i++ {
					result = append(result, i)
				}
			}
		} else {
			// Tekil port numarasıysa doğrudan ekle
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
	Short: "Yüksek hızlı TCP Port Tarayıcı",
	Run: func(cmd *cobra.Command, args []string) {

		// 1. Terminalden gelen string metni ayrıştır
		targetPorts := parsePorts(ports)
		if len(targetPorts) == 0 {
			fmt.Println("[-] Hata: Geçerli bir port aralığı girilmedi!")
			return
		}

		fmt.Printf("[+] Hedef: %s | Taranacak Port Sayısı: %d | Worker: %d\n", targetIP, len(targetPorts), workers)

		// 2. Dinamik değerleri motora gönder
		openPorts := network.StartConcurrentScan(targetIP, targetPorts, workers)

		fmt.Printf("\n[!] TARAMA TAMAMLANDI. Açık Portlar:\n")
		for _, port := range openPorts {
			fmt.Printf(" -> %d AÇIK\n", port)
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
	// Flag'lerimizi CLI'a bağlıyoruz ve varsayılan (default) değerler atıyoruz
	rootCmd.Flags().StringVarP(&targetIP, "target", "t", "", "Hedef IP adresi (Zorunlu)")
	rootCmd.Flags().StringVarP(&ports, "ports", "p", "1-1024", "Taranacak portlar (Örn: 80,443 veya 1-100)")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 100, "Eşzamanlı iş parçacığı sayısı")

	rootCmd.MarkFlagRequired("target")
}
