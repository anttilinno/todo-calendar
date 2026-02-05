package holidays

import (
	"fmt"
	"time"

	"github.com/rickar/cal/v2"
)

// Provider looks up holidays for a specific country.
type Provider struct {
	cal     *cal.Calendar
	country string
}

// NewProvider creates a holiday provider for the given country code.
// Returns an error if the country code is not in the Registry.
func NewProvider(countryCode string) (*Provider, error) {
	hols, ok := Registry[countryCode]
	if !ok {
		return nil, fmt.Errorf("unsupported country code: %q (supported: %v)", countryCode, SupportedCountries())
	}

	c := &cal.Calendar{}
	c.AddHoliday(hols...)

	return &Provider{cal: c, country: countryCode}, nil
}

// HolidaysInMonth returns a map of day numbers that are holidays
// in the given year and month.
func (p *Provider) HolidaysInMonth(year int, month time.Month) map[int]bool {
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
	result := make(map[int]bool)

	for day := 1; day <= daysInMonth; day++ {
		// Use noon to avoid timezone edge cases (research pitfall #4).
		date := time.Date(year, month, day, 12, 0, 0, 0, time.Local)
		actual, observed, _ := p.cal.IsHoliday(date)
		if actual || observed {
			result[day] = true
		}
	}

	return result
}

// Country returns the country code for this provider.
func (p *Provider) Country() string {
	return p.country
}
