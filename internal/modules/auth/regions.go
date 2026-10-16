package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// LoadRegencyCodes reads the Indonesian kabupaten/kota master list (a JSON
// array of {"id": "34.04", "name": "..."} items, as published by
// api-wilayah-indonesia v2) and returns the set of valid codes.
func LoadRegencyCodes(path string) (map[string]struct{}, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read regency master data: %w", err)
	}
	var items []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse regency master data: %w", err)
	}
	codes := make(map[string]struct{}, len(items))
	for _, item := range items {
		if !domesticCityCodePattern.MatchString(item.ID) {
			return nil, fmt.Errorf("regency master data contains an invalid code %q", item.ID)
		}
		codes[item.ID] = struct{}{}
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("regency master data is empty")
	}
	return codes, nil
}

// SetRegencyCodes installs the master list used to validate domestic city
// codes. Without it only the code format is checked.
func (s *Service) SetRegencyCodes(codes map[string]struct{}) {
	s.regencyCodes = codes
}

// regencyCodeKnown checks a domestic city code against the master list: the
// file from REGENCY_MASTER_PATH when configured, otherwise the cached region
// provider data. If neither is available it falls back to the format check
// already done by validateRegistrationProfile rather than blocking sign-ups.
func (s *Service) regencyCodeKnown(ctx context.Context, profile registrationProfile) bool {
	if profile.countryCode != "ID" {
		return true
	}
	if s.regencyCodes != nil {
		_, ok := s.regencyCodes[profile.cityCode]
		return ok
	}
	items, err := s.regencies(ctx, profile.cityCode[:2])
	if err != nil {
		return true
	}
	for _, item := range items {
		if item.ID == profile.cityCode {
			return true
		}
	}
	return false
}
