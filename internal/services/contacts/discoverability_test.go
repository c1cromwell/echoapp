package contacts

import "testing"

func TestIsPhoneDiscoverable(t *testing.T) {
	trueVal := true
	falseVal := false

	cases := []struct {
		name  string
		tier  int
		optIn *bool
		want  bool
	}{
		{"tier3 default", 3, nil, true},
		{"tier4 default", 4, nil, true},
		{"tier2 default off", 2, nil, false},
		{"tier1 default off", 1, nil, false},
		{"tier1 explicit on", 1, &trueVal, true},
		{"tier3 explicit off", 3, &falseVal, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsPhoneDiscoverable(tc.tier, tc.optIn); got != tc.want {
				t.Fatalf("IsPhoneDiscoverable(%d, %v) = %v, want %v", tc.tier, tc.optIn, got, tc.want)
			}
		})
	}
}
