package network

import (
	"net"
	"strings"
)

// GetTargets, tekil bir IP/Domain veya CIDR (örn: 192.168.1.0/24) alıp IP listesi döner
func GetTargets(target string) ([]string, error) {
	// Eğer içinde "/" yoksa, bu tek bir IP veya Domain'dir. Direkt liste olarak döndür.
	if !strings.Contains(target, "/") {
		return []string{target}, nil
	}

	ip, ipnet, err := net.ParseCIDR(target)
	if err != nil {
		return nil, err
	}

	var ips []string
	// Subnet içindeki tüm IP'leri tek tek hesapla
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// Network adresini (ilk IP) ve Broadcast adresini (son IP) çıkararak dön
	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}
	return ips, nil
}

// IP adresini bir artırır
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
