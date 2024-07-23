package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
)

func main() {
	// Text format
	fmt.Println("Text format:")
	textMetrics := getMetrics("text/plain")
	fmt.Println(string(textMetrics))

	// Binary format
	fmt.Println("\nBinary format:")
	binaryMetrics := getMetrics("application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited")
	parseBinaryMetrics(binaryMetrics)
}

func getMetrics(acceptHeader string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8080/metrics", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", acceptHeader)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return body
}

func parseBinaryMetrics(data []byte) {
	reader := bytes.NewReader(data)
	for {
		mf := &dto.MetricFamily{}
		size, err := readDelimited(reader, mf)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if size == 0 {
			break
		}

		fmt.Printf("Name: %s\n", mf.GetName())
		fmt.Printf("Help: %s\n", mf.GetHelp())
		fmt.Printf("Type: %s\n", mf.GetType())

		for _, m := range mf.GetMetric() {
			fmt.Printf("  Labels: %v\n", m.GetLabel())
			if m.Gauge != nil {
				fmt.Printf("  Gauge: %f\n", m.Gauge.GetValue())
			}
			if m.Counter != nil {
				fmt.Printf("  Counter: %f\n", m.Counter.GetValue())
			}
			// Add other metric types as needed
		}
		fmt.Println()
	}
}

// The readDelimited function is designed to read Protocol Buffer messages that are encoded in a delimited format.
// This format is used by Prometheus when sending metrics in binary form.
//
// Parameters:
//   - r io.Reader: The source from which to read the delimited message.
//   - m proto.Message: A Protocol Buffer message to unmarshal the data into.
func readDelimited(r io.Reader, m proto.Message) (int, error) {
	// a 1-byte buffer to read data one byte at a time.
	buf := make([]byte, 1)
	size := uint64(0)
	// This loop implements varint decoding. Varints are used to encode integers using a variable number of bytes.
	// Each byte uses 7 bits for the number and 1 bit to indicate if more bytes follow.
	// - It reads one byte at a time.
	// - The lower 7 bits of each byte are used to construct the size.
	// - If the most significant bit (0x80) is set, it means another byte follows.
	// - This continues until a byte with the most significant bit unset is found.
	for shift := uint(0); ; shift += 7 {
		if _, err := io.ReadFull(r, buf); err != nil {
			return 0, err
		}
		b := uint64(buf[0])
		size |= (b & 0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
	}
	// After reading the size, if it's greater than 0, proceed to read the actual message:
	if size > 0 {
		buf = make([]byte, size)
		if _, err := io.ReadFull(r, buf); err != nil {
			return 0, err
		}
		return int(size), proto.Unmarshal(buf, m)
	}
	return 0, nil
}
