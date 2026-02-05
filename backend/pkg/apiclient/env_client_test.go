package apiclient

import (
	"context"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchAsbestosData(t *testing.T) {
	client := &ApiClient{}
	cfg := &config.Config{}
	ctx := context.Background()

	// Test with Amsterdam coordinates
	data, err := client.FetchAsbestosData(ctx, cfg, 52.3676, 4.9041)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if data == nil {
		t.Fatal("Expected AsbestosData, got nil")
	}
	if data.HasAsbestosReport {
		t.Error("Expected HasAsbestosReport to be false by default")
	}
	if data.Status != "unknown" {
		t.Errorf("Expected Status 'unknown', got '%s'", data.Status)
	}
}

func TestFetchAsbestosDataLegacy(t *testing.T) {
	client := &ApiClient{}
	cfg := &config.Config{}
	geom := struct{}{} // dummy geometry

	data, err := client.FetchAsbestosDataLegacy(cfg, geom)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if data == nil {
		t.Fatal("Expected AsbestosData, got nil")
	}
	if data.HasAsbestosReport {
		t.Error("Expected HasAsbestosReport to be false by default")
	}
}
