package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadFromEnvironment(t *testing.T) {
	os.Setenv("Environment", "test")
	cfg := NewDefaultConfig()
	err := cfg.LoadFromEnvironment()
	if err != nil {
		t.Fatalf("expected no error, go %v", err)
	}
	if cfg.App.Environment != "test" {
		t.Errorf("expected environment 'test', got %s", cfg.App.Environment)
	}
}

func TestValidate(t *testing.T) {
	cfg := NewDefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
	cfg.Kafka.Brokers = []string{}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty brokers")
	}
}

func TestAppConfig_GetAppSettings(t *testing.T) {
	type fields struct {
		Kafka KafkaConfig
		App   AppSettings
	}
	tests := []struct {
		name   string
		fields fields
		want   AppSettings
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AppConfig{
				Kafka: tt.fields.Kafka,
				App:   tt.fields.App,
			}
			if got := c.GetAppSettings(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAppSettings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppConfig_GetKafkaConfig(t *testing.T) {
	type fields struct {
		Kafka KafkaConfig
		App   AppSettings
	}
	tests := []struct {
		name   string
		fields fields
		want   KafkaConfig
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AppConfig{
				Kafka: tt.fields.Kafka,
				App:   tt.fields.App,
			}
			if got := c.GetKafkaConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKafkaConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppConfig_IsDevelopment(t *testing.T) {
	type fields struct {
		Kafka KafkaConfig
		App   AppSettings
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AppConfig{
				Kafka: tt.fields.Kafka,
				App:   tt.fields.App,
			}
			if got := c.IsDevelopment(); got != tt.want {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppConfig_IsProduction(t *testing.T) {
	type fields struct {
		Kafka KafkaConfig
		App   AppSettings
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AppConfig{
				Kafka: tt.fields.Kafka,
				App:   tt.fields.App,
			}
			if got := c.IsProduction(); got != tt.want {
				t.Errorf("IsProduction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppConfig_LoadFromEnvironment(t *testing.T) {
	type fields struct {
		Kafka KafkaConfig
		App   AppSettings
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AppConfig{
				Kafka: tt.fields.Kafka,
				App:   tt.fields.App,
			}
			if err := c.LoadFromEnvironment(); (err != nil) != tt.wantErr {
				t.Errorf("LoadFromEnvironment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAppConfig_Validate(t *testing.T) {
	type fields struct {
		Kafka KafkaConfig
		App   AppSettings
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AppConfig{
				Kafka: tt.fields.Kafka,
				App:   tt.fields.App,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewDefaultConfig(t *testing.T) {
	tests := []struct {
		name string
		want *AppConfig
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
