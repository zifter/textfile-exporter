package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

type config struct {
	ServeAddr       string        `mapstructure:"serve_addr"`
	MetricsFilePath string        `mapstructure:"metrics_file_path"`
	MetricsEndpoint string        `mapstructure:"metrics_endpoint"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	LogOutput       string        `mapstructure:"log_output"`
}

func loadConfig() (config, error) {
	viper.SetDefault("serve_addr", ":8080")
	viper.SetDefault("metrics_file_path", "metrics.txt")
	viper.SetDefault("metrics_endpoint", "/metrics")
	viper.SetDefault("log_output", "stdout")
	viper.SetDefault("refresh_interval", 0*time.Second)

	viper.AutomaticEnv()

	var conf config
	if err := viper.Unmarshal(&conf); err != nil {
		return config{}, fmt.Errorf("failed to parse config: %w", err)
	}
	return conf, nil
}

type MetricsExporter struct {
	content []byte
	mu      sync.RWMutex
}

func (m *MetricsExporter) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.content = data
	return nil
}

func (m *MetricsExporter) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, _ = w.Write(m.content)
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func main() {
	conf, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	if conf.LogOutput == "stdout" {
		log.SetOutput(os.Stdout)
	}

	log.Printf("metrics file: %s, refresh interval: %v", conf.MetricsFilePath, conf.RefreshInterval)

	exporter := &MetricsExporter{}
	if err := exporter.loadFromFile(conf.MetricsFilePath); err != nil {
		log.Fatalf("failed to load metrics: %v", err)
	}

	if conf.RefreshInterval > 0 {
		go func() {
			ticker := time.NewTicker(conf.RefreshInterval)
			defer ticker.Stop()
			for range ticker.C {
				if err := exporter.loadFromFile(conf.MetricsFilePath); err != nil {
					log.Printf("failed to reload metrics: %v", err)
				}
			}
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc(conf.MetricsEndpoint, exporter.handler)
	mux.HandleFunc("/", okHandler)

	srv := &http.Server{
		Addr:    conf.ServeAddr,
		Handler: mux,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("metrics exposed on %s%s", conf.ServeAddr, conf.MetricsEndpoint)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
