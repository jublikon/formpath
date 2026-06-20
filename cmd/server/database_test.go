package main

import (
	"strings"
	"testing"
)

func TestValidateRawObjectStoreConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     appConfig
		wantErr bool
	}{
		{
			name: "complete configuration",
			cfg: appConfig{
				S3Endpoint:        "localhost:9000",
				S3AccessKeyID:     "formpath",
				S3SecretAccessKey: "formpath-secret",
			},
		},
		{
			name:    "missing configuration",
			cfg:     appConfig{},
			wantErr: true,
		},
		{
			name: "partial configuration",
			cfg: appConfig{
				S3Endpoint:    "localhost:9000",
				S3AccessKeyID: "formpath",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRawObjectStoreConfig(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected raw object store configuration error")
				}
				if !strings.Contains(err.Error(), "S3_ENDPOINT") {
					t.Fatalf("Expected actionable configuration error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Expected valid raw object store configuration, got %v", err)
			}
		})
	}
}
