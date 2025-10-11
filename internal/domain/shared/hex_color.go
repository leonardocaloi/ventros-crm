package shared

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrHexColorEmpty   = errors.New("hex color cannot be empty")
	ErrHexColorInvalid = errors.New("invalid hex color format (use #RRGGBB)")
)

var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type HexColor struct {
	value string
}

func NewHexColor(color string) (HexColor, error) {
	if color == "" {
		return HexColor{}, ErrHexColorEmpty
	}

	normalized := strings.ToUpper(strings.TrimSpace(color))

	if !strings.HasPrefix(normalized, "#") {
		normalized = "#" + normalized
	}

	if !hexColorRegex.MatchString(normalized) {
		return HexColor{}, ErrHexColorInvalid
	}

	return HexColor{value: normalized}, nil
}

func (hc HexColor) Value() string {
	return hc.value
}

func (hc HexColor) ToRGB() (r, g, b int, err error) {
	if len(hc.value) != 7 {
		return 0, 0, 0, ErrHexColorInvalid
	}

	rHex := hc.value[1:3]
	gHex := hc.value[3:5]
	bHex := hc.value[5:7]

	rVal, err := strconv.ParseInt(rHex, 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	gVal, err := strconv.ParseInt(gHex, 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	bVal, err := strconv.ParseInt(bHex, 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return int(rVal), int(gVal), int(bVal), nil
}

func (hc HexColor) Brightness() int {
	r, g, b, err := hc.ToRGB()
	if err != nil {
		return 0
	}

	brightness := (0.299 * float64(r)) + (0.587 * float64(g)) + (0.114 * float64(b))
	return int(brightness)
}

func (hc HexColor) IsDark() bool {
	return hc.Brightness() < 128
}

func (hc HexColor) IsLight() bool {
	return hc.Brightness() >= 128
}

func (hc HexColor) ContrastColor() HexColor {
	if hc.IsDark() {
		return HexColor{value: "#FFFFFF"}
	}
	return HexColor{value: "#000000"}
}

func (hc HexColor) String() string {
	return hc.value
}

func (hc HexColor) Equals(other HexColor) bool {
	return hc.value == other.value
}

func ColorRed() HexColor {
	return HexColor{value: "#FF0000"}
}

func ColorGreen() HexColor {
	return HexColor{value: "#00FF00"}
}

func ColorBlue() HexColor {
	return HexColor{value: "#0000FF"}
}

func ColorYellow() HexColor {
	return HexColor{value: "#FFFF00"}
}

func ColorWhite() HexColor {
	return HexColor{value: "#FFFFFF"}
}

func ColorBlack() HexColor {
	return HexColor{value: "#000000"}
}

func RGB(r, g, b int) (HexColor, error) {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return HexColor{}, errors.New("RGB values must be between 0 and 255")
	}

	hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)
	return HexColor{value: hex}, nil
}
