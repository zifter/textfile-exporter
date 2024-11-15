package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type config struct {
	MetricsFilePath string        `mapstructure:"metrics_file_path"`
	MetricsEndpoint string        `mapstructure:"metrics_endpoint"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
}

// Структура для хранения метрик
type MetricsExporter struct {
	content []byte
	mu      sync.RWMutex
}

func (m *MetricsExporter) loadMetricsFromFile(metricsFilePath string) error {
	data, err := os.ReadFile(metricsFilePath)
	if err != nil {
		return fmt.Errorf("error during reading file: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.content = data

	return nil
}

func (m *MetricsExporter) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	m.mu.RLock()
	defer m.mu.RUnlock()

	w.Write(m.content)
}

var conf config

func init() {
	viper.SetDefault("serve_addr", ":8080")
	viper.SetDefault("metrics_file_path", "metrics.txt")
	viper.SetDefault("metrics_endpoint", "/metrics")
	viper.SetDefault("refresh_interval", 0*time.Second)

	viper.AutomaticEnv()
	if err := viper.Unmarshal(&conf); err != nil {
		panic(err)
	}
}

func main() {
	log.Printf("metrics file: %s, refresh interval: %d",
		conf.MetricsFilePath,
		conf.RefreshInterval)

	exporter := &MetricsExporter{}
	if err := exporter.loadMetricsFromFile(conf.MetricsFilePath); err != nil {
		log.Fatalf("error during loading metrics from file: %v", err)
	}

	if conf.RefreshInterval > 0 {
		go func() {
			ticker := time.NewTicker(conf.RefreshInterval)
			defer ticker.Stop()
			for range ticker.C {
				if err := exporter.loadMetricsFromFile(conf.MetricsFilePath); err != nil {
					log.Printf("failed to load metrics file: %v", err)
				}
			}
		}()
	}

	http.HandleFunc(conf.MetricsEndpoint, exporter.metricsHandler)

	log.Printf("metrics are exposed on %s", conf.MetricsEndpoint)

	err := http.ListenAndServe(":8080", nil)
	if err == http.ErrServerClosed {
		log.Printf("HTTP/HTTPS server closed")
		os.Exit(0)
	} else {
		log.Fatal("Unable to start HTTP/HTTPS listener")
	}
}
