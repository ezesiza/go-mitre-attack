package main

import (
	"context"
	"os"
	"strings"
	"testing"
)

// main.go
func Run(ctx context.Context, args []string, env map[string]string) error {
	// ...core logic, using args/env instead of os.Args/os.Getenv directly
	return nil
}

func TestMainRun(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, os.Args, getEnvMap())
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
}

func getEnvMap() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}
