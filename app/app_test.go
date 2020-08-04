package app

import (
	"context"
	"crypto/tls"
	"github.com/madflojo/healthchecks-example/config"
	"github.com/madflojo/testcerts"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestBadConfigs(t *testing.T) {
	cfgs := make(map[string]config.Config)
	cfgs["invalid listener address"] = config.Config{
		EnableTLS:      false,
		ListenAddr:     "pandasdonotbelonghere",
		DisableLogging: true,
	}
	cfgs["invalid TLS Config"] = config.Config{
		EnableTLS:      true,
		CertFile:       "/tmp/doesntexist",
		KeyFile:        "/tmp/doesntexist",
		ListenAddr:     "0.0.0.0:8443",
		DisableLogging: true,
	}

	// Loop through bad configs, creating sub-tests as we go
	for k, v := range cfgs {
		t.Run("Testing "+k, func(t *testing.T) {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(5)*time.Second))
			defer cancel()
			go func() {
				<-ctx.Done()
				err := ctx.Err()
				if err == context.DeadlineExceeded {
					Stop()
				}
			}()
			err := Run(v)
			if err == nil || err == ErrShutdown {
				t.Errorf("Expected error when starting server, got nil")
			}
		})
	}
}

func TestRunningServer(t *testing.T) {
	go func() {
		err := Run(config.Config{
			Debug:          true,
			EnableTLS:      false,
			ListenAddr:     "localhost:9000",
			DisableLogging: true,
		})
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer Stop()

	// Wait for app to start
	time.Sleep(10 * time.Second)

	t.Run("Check Health HTTP Handler", func(t *testing.T) {
		r, err := http.Get("http://localhost:9000/health")
		if err != nil {
			t.Errorf("Unexpected error when requesting health status - %s", err)
		}
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking health - %d", r.StatusCode)
		}
	})
}

func TestRunningTLSServer(t *testing.T) {
	// Create Test Certs
	err := testcerts.GenerateCertsToFile("/tmp/cert", "/tmp/key")
	if err != nil {
		t.Errorf("Failed to create certs - %s", err)
		t.FailNow()
	}
	defer os.Remove("/tmp/cert")
	defer os.Remove("/tmp/key")

	// Disable Host Checking globally
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Start Server in goroutine
	go func() {
		err := Run(config.Config{
			Debug:          true,
			EnableTLS:      true,
			ListenAddr:     "localhost:9000",
			CertFile:       "/tmp/cert",
			KeyFile:        "/tmp/key",
			DisableLogging: true,
		})
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer Stop()

	// Wait for app to start
	time.Sleep(10 * time.Second)

	t.Run("Check Health HTTP Handler", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/health")
		if err != nil {
			t.Errorf("Unexpected error when requesting health status - %s", err)
			t.FailNow()
		}
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking health - %d", r.StatusCode)
		}
	})

	t.Run("Check Ready HTTP Handler", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/ready")
		if err != nil {
			t.Errorf("Unexpected error when requesting ready status - %s", err)
			t.FailNow()
		}
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking readiness - %d", r.StatusCode)
		}
	})

}
