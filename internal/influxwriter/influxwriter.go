package influxwriter

import (
	"context"
	"encoding/json"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// InfluxDBWriter encapsulates the InfluxDB client and configuration
type InfluxDBWriter struct {
	client   influxdb2.Client
	org      string
	bucket   string
	writeAPI api.WriteAPIBlocking
}

// NewInfluxDBWriter initializes and returns a new InfluxDBWriter instance
func NewInfluxDBWriter(url, token, org, bucket string) *InfluxDBWriter {
	client := influxdb2.NewClient(url, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)
	return &InfluxDBWriter{
		client:   client,
		org:      org,
		bucket:   bucket,
		writeAPI: writeAPI,
	}
}

// WriteJSON writes a JSON object to InfluxDB
func (w *InfluxDBWriter) WriteJSON(stats InfluxMABStats) {
	// Convert InfluxMABStats to JSON string
	jsonData, err := json.Marshal(stats)
	if err != nil {
		log.Fatalf("%s Error marshalling JSON: %v\n", INFLUXDB, err)
	}

	// Create a new data point
	point := influxdb2.NewPointWithMeasurement("mab_agent_stats").
		AddTag("new_data", "new_data").
		AddField("json_data", string(jsonData)).
		SetTime(time.Now().UTC())

	// Write the point to InfluxDB
	err = w.writeAPI.WritePoint(context.Background(), point)
	if err != nil {
		log.Fatalf("%s Error writing point to InfluxDB: %v\n", INFLUXDB, err)
	}

	log.Println(INFLUXDB, "Statistics successfully written to InfluxDB")
}

// Close closes the InfluxDB client connection
func (w *InfluxDBWriter) Close() {
	w.client.Close()
}
