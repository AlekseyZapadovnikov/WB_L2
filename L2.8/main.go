package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/beevik/ntp"
)

func main() {
	for {
		ntpTime, err := ntp.Time("ntp0.ntp-servers.net")
		if err != nil {
			slog.Error("error", "err", err)
		} else {
			fmt.Println(ntpTime)
		}
		time.Sleep(3 * time.Second)
	}
}