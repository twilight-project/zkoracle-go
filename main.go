package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

var oracleAddr string
var accountName string
var WsHub *Hub

var txCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "total_tx_zkoracle",
	Help: "Total number of tx processed.",
})

func initialize() {
	viper.AddConfigPath("./config")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("json")   // Look for specific type
	viper.ReadInConfig()

	accountName = fmt.Sprintf("%v", viper.Get("accountName"))
	command := fmt.Sprintf("nyksd keys show %s -a --keyring-backend test", accountName)
	args := strings.Fields(command)
	cmd := exec.Command(args[0], args[1:]...)

	oracleAddr_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	oracleAddr = string(oracleAddr_)
	oracleAddr = strings.ReplaceAll(oracleAddr, "\n", "")

	fmt.Println("Oracle Address: ", oracleAddr)

	// oracleAddr = "twilight1qq36nw2wk27zhedzu9ks54g2077u4wu6aau80x"
}

func main() {
	initialize()
	go server()
	go pubsubServer()
	prometheus_server()
}

func prometheus_server() {
	// Create a new instance of a registry
	reg := prometheus.NewRegistry()

	// Optional: Add Go module build info.
	reg.MustRegister(
		txCounter,
	)

	// Register the promhttp handler with the registry
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	// Simple health check endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is running"))
	})

	// Start the server
	fmt.Println("Starting prometheis server on :2550")
	if err := http.ListenAndServe(":2550", nil); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
