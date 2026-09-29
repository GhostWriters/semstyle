package semtheme

import "testing"

func TestVariant(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantVariant  string
		wantDetected bool
	}{
		{"declared", "[metadata]\nvariant = \"Light\"\n", VariantLight, false},
		{"declared wins over slots", "[metadata]\nvariant = \"dark\"\n[palette]\nbg = \"base00\"\n", VariantDark, false},
		{"palette slot", "[palette]\nbg = \"base00\"\n", VariantTinted, true},
		{"style slot", "[styles]\nTitle = \"{{[BASE0D:-:B]}}\"\n", VariantTinted, true},
		{"slot behind palette chain", "[palette]\nprimary = \"$foreground\"\nforeground = \"base05\"\n[styles]\nTitle = \"{{[$primary:-]}}\"\n", VariantTinted, true},
		{"unknown value, no slots", "[metadata]\nvariant = \"dim\"\n[styles]\nTitle = \"{{[blue:black]}}\"\n", "", false},
		{"slot-like names aren't slots", "[palette]\nbase = \"blue\"\nbase18 = \"red\"\n", "", false},
		{"none", "[styles]\nTitle = \"{{[blue:black]}}\"\n", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf, err := Parse([]byte(tt.data))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			v, detected := tf.Variant()
			if v != tt.wantVariant || detected != tt.wantDetected {
				t.Errorf("Variant() = %q, %v; want %q, %v", v, detected, tt.wantVariant, tt.wantDetected)
			}
		})
	}
}
