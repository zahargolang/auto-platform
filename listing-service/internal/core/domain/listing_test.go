package core_domain

import (
	"errors"
	"strings"
	"testing"
	core_errors "listing-service/internal/core/errors"

	"github.com/google/uuid"
)

func validListingArgs() (
	userID uuid.UUID,
	title, description string,
	price int64,
	make_, model string,
	year, mileage int,
	color string,
	bodyType BodyType,
	fuelType FuelType,
	transmission TransmissionType,
	engineVolume float64,
	city, region string,
) {
	return uuid.New(),
		"Toyota Camry 2018",
		"Clean title, one owner",
		15000000,
		"Toyota",
		"Camry",
		2018,
		45000,
		"black",
		BodyTypeSedan,
		FuelTypeGasoline,
		TransmissionAutomatic,
		2.5,
		"Almaty",
		"Almaty Region"
}

func TestListing_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(l *Listing)
		wantErr bool
	}{
		{
			name:    "valid listing",
			mutate:  func(l *Listing) {},
			wantErr: false,
		},
		{
			name:    "title too short",
			mutate:  func(l *Listing) { l.Title = "ab" },
			wantErr: true,
		},
		{
			name:    "title too long",
			mutate:  func(l *Listing) { l.Title = strings.Repeat("a", 201) },
			wantErr: true,
		},
		{
			name:    "price zero",
			mutate:  func(l *Listing) { l.Price = 0 },
			wantErr: true,
		},
		{
			name:    "price negative",
			mutate:  func(l *Listing) { l.Price = -1 },
			wantErr: true,
		},
		{
			name:    "year too old",
			mutate:  func(l *Listing) { l.Year = 1899 },
			wantErr: true,
		},
		{
			name:    "year in the far future",
			mutate:  func(l *Listing) { l.Year = 3000 },
			wantErr: true,
		},
		{
			name:    "mileage negative",
			mutate:  func(l *Listing) { l.Mileage = -1 },
			wantErr: true,
		},
		{
			name:    "engine volume negative",
			mutate:  func(l *Listing) { l.EngineVolume = -0.1 },
			wantErr: true,
		},
		{
			name:    "city empty",
			mutate:  func(l *Listing) { l.City = "" },
			wantErr: true,
		},
		{
			name:    "city too long",
			mutate:  func(l *Listing) { l.City = strings.Repeat("a", 101) },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing := NewListing(validListingArgs())
			tt.mutate(&listing)

			err := listing.Validate()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Validate() expected error, got nil")
				}
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("Validate() error = %v, want wrapped %v", err, core_errors.ErrInvalidArgument)
				}
				return
			}

			if err != nil {
				t.Fatalf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestNewListing_DefaultsStatusToActive(t *testing.T) {
	listing := NewListing(validListingArgs())

	if listing.Status != ListingStatusActive {
		t.Fatalf("expected default status %q, got %q", ListingStatusActive, listing.Status)
	}
}
