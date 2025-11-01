package apiclient

import (
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchAsbestosData(t *testing.T) {
	client := &ApiClient{}
	cfg := &config.Config{}
	geom := struct{}{} // dummy geometry

	data, err := client.FetchAsbestosData(cfg, geom)
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
