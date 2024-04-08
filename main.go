package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Report struct {
	USDCBalance      float64 `json:"usdcBalance"`
	MKLPSupply       float64 `json:"mklpSupply"`
	AdjustMKLPSupply float64 `json:"adjustMklpSupply"`
	APR              float64 `json:"apr"`
	MKLPPrice        float64 `json:"mklpPrice"`
}

func main() {
	graphiteAddress := os.Getenv("GRAPHITE_ADDRESS")

	statsdClient := NewStatsDClient(
		context.Background(),
		"stats.gauges.merklewatch.",
		graphiteAddress,
		5*time.Second,
	)
	mklp_price := statsdClient.NewGauge("mklp_price")
	usdcBalance := statsdClient.NewGauge("usdc_balance")
	mklpSupply := statsdClient.NewGauge("mklp_supply")
	adjustMKLPSupply := statsdClient.NewGauge("adjust_mklp_supply")
	apr := statsdClient.NewGauge("apr")

	fmt.Println("Merklewatch started")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigs:
			fmt.Println("Signal received, exiting...")
			return
		case <-ticker.C:
			fmt.Println("Checking data...")

			data := reportData()
			if data == nil {
				continue
			}

			mklp_price.Set(data.MKLPPrice)
			usdcBalance.Set(data.USDCBalance)
			mklpSupply.Set(data.MKLPSupply)
			adjustMKLPSupply.Set(data.AdjustMKLPSupply)
			apr.Set(data.APR)
		}
	}
}

func reportData() *Report {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://app.merkle.trade/api/v1/mklp/stats?p=30d")
	if err != nil {
		fmt.Println("Error fetching data", err)
		return nil
	}

	defer resp.Body.Close()

	var report Report
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		fmt.Println("Error decoding data")
		return nil
	}

	return &report
}
