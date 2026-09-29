package semtheme

import (
	"slices"
	"testing"
)

func TestColorInfo(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		wantColors []string
		wantKind   string
		wantExtra  []string
	}{
		{"none declared", "[styles]\nTitle = \"{{[red:black]}}\"\n", nil, "", nil},
		{"monochrome", "[metadata]\ncolors = [\"$primary\", \"$secondary\"]\n[palette]\nprimary = \"magenta\"\nsecondary = \"black\"\n[styles]\nDialog = \"{{[$primary:$secondary]}}\"\nDone = \"{{[bright-magenta:white]}}\"\n",
			[]string{"magenta", "black"}, ColorsMonochrome, nil},
		{"semi-monochrome", "[metadata]\ncolors = [\"yellow\", \"black\"]\n[styles]\nDialog = \"{{[yellow:black]}}\"\nOk = \"{{[green:black]}}\"\nBad = \"{{[black:bright-red]}}\"\n",
			[]string{"yellow", "black"}, ColorsSemiMonochrome, []string{"green", "red"}},
		{"multi-color", "[metadata]\ncolors = [\"blue\", \"cyan\", \"white\"]\n[styles]\nDialog = \"{{[black:white]}}\"\n",
			[]string{"blue", "cyan", "white"}, ColorsMultiColor, nil},
		{"tinted", "[metadata]\ncolors = [\"$primary\"]\n[palette]\nprimary = \"$fg\"\nfg = \"base05\"\n",
			[]string{"base05"}, ColorsTinted, nil},
		{"undefined palette name kept", "[metadata]\ncolors = [\"$nope\", \"red\"]\n",
			[]string{"$nope", "red"}, ColorsMonochrome, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf, err := Parse([]byte(tt.data))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			info := tf.ColorInfo()
			if !slices.Equal(info.Colors, tt.wantColors) || info.Kind != tt.wantKind || !slices.Equal(info.Extra, tt.wantExtra) {
				t.Errorf("ColorInfo() = %+v; want colors %v kind %q extra %v", info, tt.wantColors, tt.wantKind, tt.wantExtra)
			}
		})
	}
}
