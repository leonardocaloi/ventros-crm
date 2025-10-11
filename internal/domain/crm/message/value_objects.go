package message

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"
)

var (
	ErrMediaURLEmpty    = errors.New("media URL cannot be empty")
	ErrMediaURLInvalid  = errors.New("media URL is not valid")
	ErrMediaURLInsecure = errors.New("media URL must use HTTPS")
)

type MediaURL struct {
	value string
	url   *url.URL
}

func NewMediaURL(urlStr string) (MediaURL, error) {
	if urlStr == "" {
		return MediaURL{}, ErrMediaURLEmpty
	}

	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return MediaURL{}, ErrMediaURLInvalid
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return MediaURL{}, ErrMediaURLInvalid
	}

	return MediaURL{
		value: urlStr,
		url:   parsedURL,
	}, nil
}

func NewSecureMediaURL(urlStr string) (MediaURL, error) {
	mu, err := NewMediaURL(urlStr)
	if err != nil {
		return MediaURL{}, err
	}

	if !mu.IsSecure() {
		return MediaURL{}, ErrMediaURLInsecure
	}

	return mu, nil
}

func (mu MediaURL) Value() string {
	return mu.value
}

func (mu MediaURL) IsSecure() bool {
	return mu.url != nil && strings.ToLower(mu.url.Scheme) == "https"
}

func (mu MediaURL) Domain() string {
	if mu.url == nil {
		return ""
	}
	return mu.url.Host
}

func (mu MediaURL) Path() string {
	if mu.url == nil {
		return ""
	}
	return mu.url.Path
}

func (mu MediaURL) Scheme() string {
	if mu.url == nil {
		return ""
	}
	return strings.ToLower(mu.url.Scheme)
}

func (mu MediaURL) Extension() string {
	if mu.url == nil {
		return ""
	}

	path := mu.url.Path
	lastDot := strings.LastIndex(path, ".")
	if lastDot == -1 || lastDot == len(path)-1 {
		return ""
	}

	return strings.ToLower(path[lastDot+1:])
}

func (mu MediaURL) IsImage() bool {
	ext := mu.Extension()
	imageExts := []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

func (mu MediaURL) IsVideo() bool {
	ext := mu.Extension()
	videoExts := []string{"mp4", "avi", "mov", "mkv", "webm", "flv", "wmv"}
	for _, vidExt := range videoExts {
		if ext == vidExt {
			return true
		}
	}
	return false
}

func (mu MediaURL) IsAudio() bool {
	ext := mu.Extension()
	audioExts := []string{"mp3", "wav", "ogg", "m4a", "flac", "aac", "wma"}
	for _, audExt := range audioExts {
		if ext == audExt {
			return true
		}
	}
	return false
}

func (mu MediaURL) String() string {
	return mu.value
}

func (mu MediaURL) Equals(other MediaURL) bool {
	return mu.value == other.value
}

const (
	MaxTextLength = 4096
)

var (
	ErrTextEmpty   = errors.New("message text cannot be empty")
	ErrTextTooLong = errors.New("message text exceeds maximum length of 4096 characters")
	ErrTextInvalid = errors.New("message text contains invalid characters")
)

type MessageText struct {
	value string
}

func NewMessageText(text string) (MessageText, error) {
	if text == "" {
		return MessageText{}, ErrTextEmpty
	}

	length := utf8.RuneCountInString(text)
	if length > MaxTextLength {
		return MessageText{}, ErrTextTooLong
	}

	if !utf8.ValidString(text) {
		return MessageText{}, ErrTextInvalid
	}

	return MessageText{value: text}, nil
}

func (mt MessageText) Value() string {
	return mt.value
}

func (mt MessageText) Length() int {
	return utf8.RuneCountInString(mt.value)
}

func (mt MessageText) IsEmpty() bool {
	return mt.value == ""
}

func (mt MessageText) Truncate(maxLength int) MessageText {
	if maxLength <= 0 {
		return MessageText{value: ""}
	}

	runes := []rune(mt.value)
	if len(runes) <= maxLength {
		return mt
	}

	if maxLength <= 3 {
		return MessageText{value: string(runes[:maxLength])}
	}

	truncated := string(runes[:maxLength-3]) + "..."
	return MessageText{value: truncated}
}

func (mt MessageText) Contains(substr string) bool {
	return utf8.RuneCountInString(substr) > 0 &&
		len(mt.value) > 0 &&
		containsString(mt.value, substr)
}

func (mt MessageText) String() string {
	return mt.value
}

func (mt MessageText) Equals(other MessageText) bool {
	return mt.value == other.value
}

func containsString(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
