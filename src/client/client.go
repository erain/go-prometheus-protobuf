package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func main() {
	url := "http://localhost:8080/metrics"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching metrics: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	metrics, err := parseBinaryMetrics(body)
	if err != nil {
		fmt.Printf("Error parsing metrics: %v\n", err)
		return
	}

	for _, mf := range metrics {
		fmt.Printf("Metric: %s\n", mf.GetName())
		fmt.Printf("Help: %s\n", mf.GetHelp())
		fmt.Printf("Type: %s\n", mf.GetType())
		for _, m := range mf.GetMetric() {
			fmt.Printf("  Labels: %v\n", m.GetLabel())
			if m.Gauge != nil {
				fmt.Printf("  Gauge Value: %f\n", m.Gauge.GetValue())
			}
			if m.Counter != nil {
				fmt.Printf("  Counter Value: %f\n", m.Counter.GetValue())
			}
			// Add more metric types as needed
		}
		fmt.Println()
	}
}

func parseBinaryMetrics(data []byte) ([]*io_prometheus_client.MetricFamily, error) {
	var parser expfmt.TextParser
	metrics, err := parser.TextToMetricFamilies(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var result []*io_prometheus_client.MetricFamily
	for _, mf := range metrics {
		result = append(result, mf)
	}

	return result, nil
}
