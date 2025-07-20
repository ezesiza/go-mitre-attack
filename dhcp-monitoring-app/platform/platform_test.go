package platform

import (
	"context"
	"dhcp-monitoring-app/config"
	"time"

	"testing"
)

func TestNewDefaultServiceContainer(t *testing.T) {
	cfg := config.NewDefaultConfig()
	container, err := NewDefaultServiceContainer(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if container == nil {
		t.Error("expected non-nil container")
	}
}

func TestPlatform_StartupShutdown(t *testing.T) {
	cfg := config.NewDefaultConfig()
	container, err := NewDefaultServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create service container: %v", err)
	}
	p, err := NewPlatformBuilder().WithServiceContainer(container).Build()
	if err != nil {
		t.Fatalf("Failed to build platform: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := p.Start(ctx); err != nil && err != context.Canceled {
			t.Errorf("Platform failed to start: %v", err)
		}
	}()
	// Give the platform a moment to start
	time.Sleep(100 * time.Millisecond)
	cancel()
	done := make(chan struct{})
	go func() {
		p.Stop()
		close(done)
	}()
	select {
	case <-done:
		// stopped successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Platform did not stop in time")
	}
}
