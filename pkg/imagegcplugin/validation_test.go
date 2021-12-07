package imagegcplugin

import (
	"testing"
	"time"
)

func TestPolicyValidation(t *testing.T) {
	testcases := []struct {
		p        GcPolicy
		hasError bool
	}{
		// LowThresholdPercent must be in (1, 100]
		{
			p:        GcPolicy{LowThresholdPercent: 0},
			hasError: true,
		},
		{
			p:        GcPolicy{LowThresholdPercent: 101},
			hasError: true,
		},
		// HighThresholdPercent must be in (1, 100]
		{
			p:        GcPolicy{LowThresholdPercent: 1, HighThresholdPercent: 0},
			hasError: true,
		},
		{
			p:        GcPolicy{LowThresholdPercent: 1, HighThresholdPercent: 101},
			hasError: true,
		},
		// MinAge must be > 0
		{
			p:        GcPolicy{LowThresholdPercent: 1, HighThresholdPercent: 2},
			hasError: true,
		},
		// Whitelist must have element
		{
			p:        GcPolicy{LowThresholdPercent: 1, HighThresholdPercent: 2, MinAge: 1 * time.Second},
			hasError: true,
		},
		{
			p:        GcPolicy{LowThresholdPercent: 1, HighThresholdPercent: 2, MinAge: 1 * time.Second, Whitelist: []string{"pause"}},
			hasError: false,
		},
		// LowThresholdPercent must be lower HighThresholdPercent
		{
			p:        GcPolicy{LowThresholdPercent: 5, HighThresholdPercent: 2, MinAge: 1 * time.Second, Whitelist: []string{"pause"}},
			hasError: true,
		},
	}

	for _, tc := range testcases {
		if err := validateGCPolicy(tc.p); (err != nil) != tc.hasError {
			t.Errorf("expect hasError=%v, but got error=%v", tc.hasError, err)
		}
	}
}
