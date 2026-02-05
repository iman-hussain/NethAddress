package utils

import "strings"

// NormalizeAddressInput normalises postcode and house number for API lookups.
// Postcodes are uppercased with spaces removed; house numbers are trimmed.
func NormalizeAddressInput(postcode, houseNumber string) (string, string) {
	return strings.ToUpper(strings.ReplaceAll(postcode, " ", "")),
		strings.TrimSpace(houseNumber)
}
