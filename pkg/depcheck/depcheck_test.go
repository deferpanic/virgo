package depcheck

import (
	"testing"
)

func TestIsValidDarwin(t *testing.T) {
	tt := []struct {
		Version  string
		Validity bool
	}{

		{"10.11.3", false},
		{"9.1.3", false},
		{"7.1.3", false},
		{"6.1.2", false},
		{"", false},
		{"9", false},
		{"9.12", false},

		{"11.11.4", true},
		{"10.11.5", true},
		{"10.11.6", true},
		{"10.12", true},
		{"11", true},
		{"10.12.2", true},
		{"10.12.3", true},
		{"10.12.6", true},
		{"10.13.1", true},
		{"10.13.3", true},
		{"10.14.0", true},
	}

	for _, tc := range tt {
		v := IsValidDarwin(tc.Version)
		if v != tc.Validity {
			t.Errorf("Exepcted IsValidDarwin(\"%s\") to be %v, but was %v", tc.Version, tc.Validity, v)
		}
	}
}
