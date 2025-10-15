package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHexColor_Valid(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  string
	}{
		{
			name:  "valid hex with #",
			color: "#FF5733",
			want:  "#FF5733",
		},
		{
			name:  "valid hex without #",
			color: "FF5733",
			want:  "#FF5733",
		},
		{
			name:  "lowercase hex",
			color: "#ff5733",
			want:  "#FF5733",
		},
		{
			name:  "lowercase hex without #",
			color: "ff5733",
			want:  "#FF5733",
		},
		{
			name:  "mixed case",
			color: "#Ff5733",
			want:  "#FF5733",
		},
		{
			name:  "with whitespace",
			color: "  #FF5733  ",
			want:  "#FF5733",
		},
		{
			name:  "black",
			color: "#000000",
			want:  "#000000",
		},
		{
			name:  "white",
			color: "#FFFFFF",
			want:  "#FFFFFF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc, err := NewHexColor(tt.color)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, hc.Value())
		})
	}
}

func TestNewHexColor_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		color   string
		wantErr error
	}{
		{
			name:    "empty string",
			color:   "",
			wantErr: ErrHexColorEmpty,
		},
		{
			name:    "too short",
			color:   "#FF57",
			wantErr: ErrHexColorInvalid,
		},
		{
			name:    "too long",
			color:   "#FF57333",
			wantErr: ErrHexColorInvalid,
		},
		{
			name:    "invalid characters",
			color:   "#GGGGGG",
			wantErr: ErrHexColorInvalid,
		},
		{
			name:    "invalid format",
			color:   "FF5733!",
			wantErr: ErrHexColorInvalid,
		},
		{
			name:    "spaces in middle",
			color:   "#FF 5733",
			wantErr: ErrHexColorInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHexColor(tt.color)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHexColor_ToRGB(t *testing.T) {
	tests := []struct {
		name  string
		color string
		wantR int
		wantG int
		wantB int
	}{
		{
			name:  "red",
			color: "#FF0000",
			wantR: 255,
			wantG: 0,
			wantB: 0,
		},
		{
			name:  "green",
			color: "#00FF00",
			wantR: 0,
			wantG: 255,
			wantB: 0,
		},
		{
			name:  "blue",
			color: "#0000FF",
			wantR: 0,
			wantG: 0,
			wantB: 255,
		},
		{
			name:  "white",
			color: "#FFFFFF",
			wantR: 255,
			wantG: 255,
			wantB: 255,
		},
		{
			name:  "black",
			color: "#000000",
			wantR: 0,
			wantG: 0,
			wantB: 0,
		},
		{
			name:  "custom color",
			color: "#FF5733",
			wantR: 255,
			wantG: 87,
			wantB: 51,
		},
		{
			name:  "gray",
			color: "#808080",
			wantR: 128,
			wantG: 128,
			wantB: 128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc, _ := NewHexColor(tt.color)
			r, g, b, err := hc.ToRGB()
			assert.NoError(t, err)
			assert.Equal(t, tt.wantR, r)
			assert.Equal(t, tt.wantG, g)
			assert.Equal(t, tt.wantB, b)
		})
	}
}

func TestHexColor_Brightness(t *testing.T) {
	tests := []struct {
		name           string
		color          string
		wantBrightness int
	}{
		{
			name:           "black is darkest",
			color:          "#000000",
			wantBrightness: 0,
		},
		{
			name:           "white is brightest",
			color:          "#FFFFFF",
			wantBrightness: 255,
		},
		{
			name:           "red has medium brightness",
			color:          "#FF0000",
			wantBrightness: 76, // 0.299 * 255
		},
		{
			name:           "green has high brightness",
			color:          "#00FF00",
			wantBrightness: 149, // 0.587 * 255
		},
		{
			name:           "blue has low brightness",
			color:          "#0000FF",
			wantBrightness: 29, // 0.114 * 255
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc, _ := NewHexColor(tt.color)
			assert.Equal(t, tt.wantBrightness, hc.Brightness())
		})
	}
}

func TestHexColor_IsDark(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		{
			name:  "black is dark",
			color: "#000000",
			want:  true,
		},
		{
			name:  "dark gray is dark",
			color: "#404040",
			want:  true,
		},
		{
			name:  "red is dark",
			color: "#FF0000",
			want:  true,
		},
		{
			name:  "blue is dark",
			color: "#0000FF",
			want:  true,
		},
		{
			name:  "white is not dark",
			color: "#FFFFFF",
			want:  false,
		},
		{
			name:  "light gray is not dark",
			color: "#C0C0C0",
			want:  false,
		},
		{
			name:  "yellow is not dark",
			color: "#FFFF00",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc, _ := NewHexColor(tt.color)
			assert.Equal(t, tt.want, hc.IsDark())
		})
	}
}

func TestHexColor_IsLight(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		{
			name:  "white is light",
			color: "#FFFFFF",
			want:  true,
		},
		{
			name:  "light gray is light",
			color: "#C0C0C0",
			want:  true,
		},
		{
			name:  "yellow is light",
			color: "#FFFF00",
			want:  true,
		},
		{
			name:  "black is not light",
			color: "#000000",
			want:  false,
		},
		{
			name:  "dark gray is not light",
			color: "#404040",
			want:  false,
		},
		{
			name:  "red is not light",
			color: "#FF0000",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc, _ := NewHexColor(tt.color)
			assert.Equal(t, tt.want, hc.IsLight())
		})
	}
}

func TestHexColor_ContrastColor(t *testing.T) {
	tests := []struct {
		name         string
		color        string
		wantContrast string
	}{
		{
			name:         "black background needs white text",
			color:        "#000000",
			wantContrast: "#FFFFFF",
		},
		{
			name:         "dark blue needs white text",
			color:        "#0000FF",
			wantContrast: "#FFFFFF",
		},
		{
			name:         "dark red needs white text",
			color:        "#FF0000",
			wantContrast: "#FFFFFF",
		},
		{
			name:         "white background needs black text",
			color:        "#FFFFFF",
			wantContrast: "#000000",
		},
		{
			name:         "yellow background needs black text",
			color:        "#FFFF00",
			wantContrast: "#000000",
		},
		{
			name:         "light gray needs black text",
			color:        "#C0C0C0",
			wantContrast: "#000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc, _ := NewHexColor(tt.color)
			contrast := hc.ContrastColor()
			assert.Equal(t, tt.wantContrast, contrast.Value())
		})
	}
}

func TestHexColor_String(t *testing.T) {
	color := "#FF5733"
	hc, _ := NewHexColor(color)
	assert.Equal(t, color, hc.String())
}

func TestHexColor_Equals(t *testing.T) {
	hc1, _ := NewHexColor("#FF5733")
	hc2, _ := NewHexColor("#FF5733")
	hc3, _ := NewHexColor("#0000FF")
	hc4, _ := NewHexColor("ff5733") // lowercase, should normalize to #FF5733

	assert.True(t, hc1.Equals(hc2))
	assert.True(t, hc1.Equals(hc4)) // case-insensitive
	assert.False(t, hc1.Equals(hc3))
}

func TestPredefinedColors(t *testing.T) {
	tests := []struct {
		name    string
		colorFn func() HexColor
		wantHex string
		wantR   int
		wantG   int
		wantB   int
	}{
		{
			name:    "ColorRed",
			colorFn: ColorRed,
			wantHex: "#FF0000",
			wantR:   255,
			wantG:   0,
			wantB:   0,
		},
		{
			name:    "ColorGreen",
			colorFn: ColorGreen,
			wantHex: "#00FF00",
			wantR:   0,
			wantG:   255,
			wantB:   0,
		},
		{
			name:    "ColorBlue",
			colorFn: ColorBlue,
			wantHex: "#0000FF",
			wantR:   0,
			wantG:   0,
			wantB:   255,
		},
		{
			name:    "ColorYellow",
			colorFn: ColorYellow,
			wantHex: "#FFFF00",
			wantR:   255,
			wantG:   255,
			wantB:   0,
		},
		{
			name:    "ColorWhite",
			colorFn: ColorWhite,
			wantHex: "#FFFFFF",
			wantR:   255,
			wantG:   255,
			wantB:   255,
		},
		{
			name:    "ColorBlack",
			colorFn: ColorBlack,
			wantHex: "#000000",
			wantR:   0,
			wantG:   0,
			wantB:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := tt.colorFn()
			assert.Equal(t, tt.wantHex, color.Value())

			r, g, b, err := color.ToRGB()
			assert.NoError(t, err)
			assert.Equal(t, tt.wantR, r)
			assert.Equal(t, tt.wantG, g)
			assert.Equal(t, tt.wantB, b)
		})
	}
}

func TestRGB(t *testing.T) {
	tests := []struct {
		name    string
		r       int
		g       int
		b       int
		wantHex string
		wantErr error
	}{
		{
			name:    "valid RGB",
			r:       255,
			g:       87,
			b:       51,
			wantHex: "#FF5733",
			wantErr: nil,
		},
		{
			name:    "black",
			r:       0,
			g:       0,
			b:       0,
			wantHex: "#000000",
			wantErr: nil,
		},
		{
			name:    "white",
			r:       255,
			g:       255,
			b:       255,
			wantHex: "#FFFFFF",
			wantErr: nil,
		},
		{
			name:    "r out of range (negative)",
			r:       -1,
			g:       0,
			b:       0,
			wantHex: "",
			wantErr: ErrHexColorInvalid,
		},
		{
			name:    "r out of range (too high)",
			r:       256,
			g:       0,
			b:       0,
			wantHex: "",
			wantErr: ErrHexColorInvalid,
		},
		{
			name:    "g out of range",
			r:       0,
			g:       300,
			b:       0,
			wantHex: "",
			wantErr: ErrHexColorInvalid,
		},
		{
			name:    "b out of range",
			r:       0,
			g:       0,
			b:       -50,
			wantHex: "",
			wantErr: ErrHexColorInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc, err := RGB(tt.r, tt.g, tt.b)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantHex, hc.Value())
			}
		})
	}
}

func TestRGB_RoundTrip(t *testing.T) {
	// Test that converting RGB -> Hex -> RGB gives same values
	tests := []struct {
		name string
		r    int
		g    int
		b    int
	}{
		{
			name: "red",
			r:    255,
			g:    0,
			b:    0,
		},
		{
			name: "custom color",
			r:    123,
			g:    45,
			b:    67,
		},
		{
			name: "gray",
			r:    128,
			g:    128,
			b:    128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// RGB -> Hex
			hc, err := RGB(tt.r, tt.g, tt.b)
			assert.NoError(t, err)

			// Hex -> RGB
			r, g, b, err := hc.ToRGB()
			assert.NoError(t, err)

			// Should be same
			assert.Equal(t, tt.r, r)
			assert.Equal(t, tt.g, g)
			assert.Equal(t, tt.b, b)
		})
	}
}
