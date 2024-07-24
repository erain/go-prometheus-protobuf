package main

import (
    "net/http"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Create a new registry
    registry := prometheus.NewRegistry()

    // Gauge metric
    gauge := prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "example_gauge",
        Help: "An example gauge metric",
    })
    registry.MustRegister(gauge)
    gauge.Set(42.0)

    // Counter metric
    counter := prometheus.NewCounter(prometheus.CounterOpts{
        Name: "example_counter",
        Help: "An example counter metric",
    })
    registry.MustRegister(counter)
    counter.Inc()

    // Summary metric
    summary := prometheus.NewSummary(prometheus.SummaryOpts{
        Name:       "example_summary",
        Help:       "An example summary metric",
        Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
    })
    registry.MustRegister(summary)
    summary.Observe(30)
    summary.Observe(60)
    summary.Observe(90)

    // Histogram metric
    histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
        Name:    "example_histogram",
        Help:    "An example histogram metric",
        Buckets: prometheus.LinearBuckets(0, 10, 5), // 5 buckets, starting at 0, width 10
    })
    registry.MustRegister(histogram)
    histogram.Observe(5)
    histogram.Observe(15)
    histogram.Observe(25)

    // Untyped metric
    untyped := prometheus.NewUntypedFunc(prometheus.UntypedOpts{
        Name: "example_untyped",
        Help: "An example untyped metric",
    }, func() float64 {
        return 12.34
    })
    registry.MustRegister(untyped)

    // Create a handler for the metrics endpoint
    handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

    // Set up the HTTP server
    http.Handle("/metrics", handler)
    http.ListenAndServe(":8080", nil)
}