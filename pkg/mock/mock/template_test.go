package mock

import "testing"

func TestIDName(t *testing.T) {
	var testCases = []struct {
		name     string
		identity string
		nth      int
		expected string
	}{
		{
			"0nth place returns Victor for 38211020353",
			"38211020353",
			0,
			"Victor",
		},
		{
			"1nth place returns Foxtrot for 38211020353",
			"38211020353",
			1,
			"Foxtrot",
		},
		{
			"1nth place returns Foxtrot for 38211020349",
			"38211020349",
			1,
			"Uniform",
		},
		{
			"out of range place returns Sierra for 38211020353",
			"38211020353",
			1000,
			"Sierra",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := fromIdentity(tc.identity)

			idName := tmpl.IDNameNth(tc.nth)

			if idName != tc.expected {
				t.Errorf("incorrect idName! %v", idName)
			}
		})
	}

}

func TestIDNvl2(t *testing.T) {
	var testCases = []struct {
		name     string
		identity string
		nth      int
		expected string
	}{
		{
			"0nth place returns odd for 38211020353",
			"38211020353",
			0,
			"odd",
		},
		{
			"0nth place returns odd for 48211020353",
			"48211020353",
			0,
			"even",
		},
		{
			"last place returns odd for 38211020353",
			"38211020353",
			11,
			"odd",
		},
		{
			"last place returns odd for 38211020354",
			"38211020354",
			11,
			"even",
		},
		{
			"99 out of range place returns even for 38211020354",
			"38211020354",
			99,
			"even",
		},
		{
			"-99 out of range place returns odd for 38211020354",
			"38211020354",
			-99,
			"odd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := fromIdentity(tc.identity)

			result := tmpl.IDNvl2(tc.nth, "odd", "even")

			if result != tc.expected {
				t.Errorf("incorrect value! %v != %v", tc.expected, result)
			}
		})
	}

}
