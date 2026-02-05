package holidays

import (
	"sort"

	"github.com/rickar/cal/v2"
	"github.com/rickar/cal/v2/de"
	"github.com/rickar/cal/v2/dk"
	"github.com/rickar/cal/v2/es"
	"github.com/rickar/cal/v2/fi"
	"github.com/rickar/cal/v2/fr"
	"github.com/rickar/cal/v2/gb"
	"github.com/rickar/cal/v2/it"
	"github.com/rickar/cal/v2/no"
	"github.com/rickar/cal/v2/se"
	"github.com/rickar/cal/v2/us"
)

// Registry maps lowercase ISO country codes to their holiday definitions.
var Registry = map[string][]*cal.Holiday{
	"de": de.Holidays,
	"dk": dk.Holidays,
	"es": es.Holidays,
	"fi": fi.Holidays,
	"fr": fr.Holidays,
	"gb": gb.Holidays,
	"it": it.Holidays,
	"no": no.Holidays,
	"se": se.Holidays,
	"us": us.Holidays,
}

// SupportedCountries returns a sorted list of supported country codes.
func SupportedCountries() []string {
	codes := make([]string, 0, len(Registry))
	for code := range Registry {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}
