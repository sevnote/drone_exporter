package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	//_ "go.uber.org/automaxprocs"
	"drone_exporter"
	"log"
)

func main() {
	http.Handle("/metrics", promhttp.HandlerFor(drone_exporter.Reg, promhttp.HandlerOpts{}))
	log.Print("Started")
	log.Fatal(http.ListenAndServe(":9125", nil))
}
