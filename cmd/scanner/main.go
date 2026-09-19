package main

import (
	"github.com/namikozkose/GoRecon/internal/cli"
)

func main() {
	// Sadece CLI'yi tetikliyoruz, geri kalan her şeyi CLI yönetecek.
	cli.Execute()
}
