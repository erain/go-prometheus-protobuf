package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
)

func main() {
	url := "http://localhost:8080/metrics"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Accept", "application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching metrics: %v\n", err)
		return
	}
	defer resp.Body.Close()

	metrics, err := parseBinaryMetrics(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing metrics: %v\n", err)
		return
	}

	for _, mf := range metrics {
		fmt.Printf("Metric: %s\n", mf.GetName())
		fmt.Printf("Help: %s\n", mf.GetHelp())
		fmt.Printf("Type: %s\n", mf.GetType())
		for _, m := range mf.GetMetric() {
			fmt.Printf("  Labels: %v\n", labelPairs(m.GetLabel()))
			printMetricValue(m)
		}
		fmt.Println()
	}
}

func parseBinaryMetrics(r io.Reader) ([]*dto.MetricFamily, error) {
	var metrics []*dto.MetricFamily
	reader := bufio.NewReader(r)

	for {
		length, err := binary.ReadUvarint(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading metric length: %v", err)
		}

		buffer := make([]byte, length)
		_, err = io.ReadFull(reader, buffer)
		if err != nil {
			return nil, fmt.Errorf("error reading metric data: %v", err)
		}

		var mf dto.MetricFamily
		if err := proto.Unmarshal(buffer, &mf); err != nil {
			return nil, fmt.Errorf("error unmarshaling metric family: %v", err)
		}
		metrics = append(metrics, &mf)
	}

	return metrics, nil
}

func labelPairs(labels []*dto.LabelPair) string {
	result := "{"
	for i, label := range labels {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s=%q", label.GetName(), label.GetValue())
	}
	result += "}"
	return result
}

func printMetricValue(m *dto.Metric) {
	switch {
	case m.Gauge != nil:
		fmt.Printf("  Gauge Value: %f\n", m.Gauge.GetValue())
	case m.Counter != nil:
		fmt.Printf("  Counter Value: %f\n", m.Counter.GetValue())
	case m.Summary != nil:
		fmt.Printf("  Summary:\n")
		fmt.Printf("    Sample Count: %d\n", m.Summary.GetSampleCount())
		fmt.Printf("    Sample Sum: %f\n", m.Summary.GetSampleSum())
		for _, q := range m.Summary.GetQuantile() {
			fmt.Printf("    Quantile %.2f: %f\n", q.GetQuantile(), q.GetValue())
		}
	case m.Histogram != nil:
		fmt.Printf("  Histogram:\n")
		fmt.Printf("    Sample Count: %d\n", m.Histogram.GetSampleCount())
		fmt.Printf("    Sample Sum: %f\n", m.Histogram.GetSampleSum())
		for _, b := range m.Histogram.GetBucket() {
			fmt.Printf("    Bucket [%f]: %d\n", b.GetUpperBound(), b.GetCumulativeCount())
		}
	case m.Untyped != nil:
		fmt.Printf("  Untyped Value: %f\n", m.Untyped.GetValue())
	default:
		fmt.Println("  Unknown metric type")
	}
}
