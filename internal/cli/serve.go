package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/namikozkose/GoRecon/internal/network"
	"github.com/spf13/cobra"
)

var servePort string

// serveCmd, uygulamamızı bir REST API sunucusu olarak ayağa kaldırır
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "GoRecon REST API (C2) Sunucusunu başlatır",
	Run: func(cmd *cobra.Command, args []string) {

		// /scan uç noktasını (endpoint) tanımlıyoruz
		http.HandleFunc("/scan", scanHandler)

		fmt.Println("---------------------------------------------------")
		fmt.Printf("[+] C2 KONTROL MERKEZİ AKTİF\n")
		fmt.Printf("[+] Sunucu dinleniyor: http://localhost:%s\n", servePort)
		fmt.Println("[+] Örnek İstek: GET /scan?target=scanme.nmap.org&ports=22,80,443")
		fmt.Println("---------------------------------------------------")

		// Sunucuyu başlat
		if err := http.ListenAndServe(":"+servePort, nil); err != nil {
			fmt.Printf("[-] Sunucu hatası: %v\n", err)
		}
	},
}

func init() {
	// API'nin çalışacağı portu belirleyen flag (-P)
	serveCmd.Flags().StringVarP(&servePort, "port", "P", "8080", "API'nin dinleyeceği port (Örn: 8080)")

	// 'serve' komutunu ana programa (rootCmd) bağlıyoruz
	rootCmd.AddCommand(serveCmd)
}

// API'ye gelen istekleri işleyen ve taramayı başlatan fonksiyon
func scanHandler(w http.ResponseWriter, r *http.Request) {
	// URL'den parametreleri al (?target=...&ports=...)
	targetParam := r.URL.Query().Get("target")
	portsParam := r.URL.Query().Get("ports")
	workersParam := r.URL.Query().Get("workers")

	// Güvenlik: Zorunlu parametreler yoksa 400 Bad Request dön
	if targetParam == "" || portsParam == "" {
		http.Error(w, `{"error": "target ve ports parametreleri zorunludur"}`, http.StatusBadRequest)
		return
	}

	wks := 100 // Varsayılan worker
	if workersParam != "" {
		if val, err := strconv.Atoi(workersParam); err == nil {
			wks = val
		}
	}

	targetPorts := parsePorts(portsParam)
	targets, err := network.GetTargets(targetParam)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz hedef IP veya CIDR"}`, http.StatusBadRequest)
		return
	}

	allResults := make(map[string][]network.PortResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Tarama Motorunu Ateşle
	for _, ip := range targets {
		wg.Add(1)
		go func(targetIP string) {
			defer wg.Done()
			openPorts := network.StartConcurrentScan(targetIP, targetPorts, wks)
			if len(openPorts) > 0 {
				mu.Lock()
				allResults[targetIP] = openPorts
				mu.Unlock()
			}
		}(ip)
	}
	wg.Wait()

	// Sonuçları JSON olarak tarayıcıya/istemciye fırlat
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allResults)
}
