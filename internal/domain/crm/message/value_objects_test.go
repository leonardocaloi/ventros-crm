package message

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMediaURL_Valid(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "https URL",
			url:  "https://example.com/image.jpg",
		},
		{
			name: "http URL",
			url:  "http://example.com/video.mp4",
		},
		{
			name: "URL with path",
			url:  "https://cdn.example.com/media/files/document.pdf",
		},
		{
			name: "URL with query params",
			url:  "https://example.com/image.jpg?size=large&format=webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, err := NewMediaURL(tt.url)
			assert.NoError(t, err)
			assert.Equal(t, tt.url, mu.Value())
		})
	}
}

func TestNewMediaURL_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		{
			name:    "empty URL",
			url:     "",
			wantErr: ErrMediaURLEmpty,
		},
		{
			name:    "invalid URL",
			url:     "not a url",
			wantErr: ErrMediaURLInvalid,
		},
		{
			name:    "invalid scheme",
			url:     "ftp://example.com/file.txt",
			wantErr: ErrMediaURLInvalid,
		},
		{
			name:    "no scheme",
			url:     "example.com/image.jpg",
			wantErr: ErrMediaURLInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMediaURL(tt.url)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestNewSecureMediaURL(t *testing.T) {
	t.Run("accepts https", func(t *testing.T) {
		mu, err := NewSecureMediaURL("https://example.com/image.jpg")
		assert.NoError(t, err)
		assert.True(t, mu.IsSecure())
	})

	t.Run("rejects http", func(t *testing.T) {
		_, err := NewSecureMediaURL("http://example.com/image.jpg")
		assert.Error(t, err)
		assert.Equal(t, ErrMediaURLInsecure, err)
	})
}

func TestMediaURL_IsSecure(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "https is secure",
			url:  "https://example.com/image.jpg",
			want: true,
		},
		{
			name: "http is not secure",
			url:  "http://example.com/image.jpg",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, _ := NewMediaURL(tt.url)
			assert.Equal(t, tt.want, mu.IsSecure())
		})
	}
}

func TestMediaURL_Domain(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantDomain string
	}{
		{
			name:       "simple domain",
			url:        "https://example.com/image.jpg",
			wantDomain: "example.com",
		},
		{
			name:       "subdomain",
			url:        "https://cdn.example.com/image.jpg",
			wantDomain: "cdn.example.com",
		},
		{
			name:       "with port",
			url:        "https://example.com:8080/image.jpg",
			wantDomain: "example.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, _ := NewMediaURL(tt.url)
			assert.Equal(t, tt.wantDomain, mu.Domain())
		})
	}
}

func TestMediaURL_Path(t *testing.T) {
	mu, _ := NewMediaURL("https://example.com/media/files/image.jpg")
	assert.Equal(t, "/media/files/image.jpg", mu.Path())
}

func TestMediaURL_Scheme(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantScheme string
	}{
		{
			name:       "https scheme",
			url:        "https://example.com/image.jpg",
			wantScheme: "https",
		},
		{
			name:       "http scheme",
			url:        "http://example.com/image.jpg",
			wantScheme: "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, _ := NewMediaURL(tt.url)
			assert.Equal(t, tt.wantScheme, mu.Scheme())
		})
	}
}

func TestMediaURL_Extension(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantExt string
	}{
		{
			name:    "jpg extension",
			url:     "https://example.com/image.jpg",
			wantExt: "jpg",
		},
		{
			name:    "png extension",
			url:     "https://example.com/image.PNG",
			wantExt: "png",
		},
		{
			name:    "no extension",
			url:     "https://example.com/image",
			wantExt: "",
		},
		{
			name:    "extension with query params",
			url:     "https://example.com/image.jpg?size=large",
			wantExt: "jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, _ := NewMediaURL(tt.url)
			assert.Equal(t, tt.wantExt, mu.Extension())
		})
	}
}

func TestMediaURL_IsImage(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "jpg is image",
			url:  "https://example.com/photo.jpg",
			want: true,
		},
		{
			name: "png is image",
			url:  "https://example.com/photo.png",
			want: true,
		},
		{
			name: "mp4 is not image",
			url:  "https://example.com/video.mp4",
			want: false,
		},
		{
			name: "no extension",
			url:  "https://example.com/file",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, _ := NewMediaURL(tt.url)
			assert.Equal(t, tt.want, mu.IsImage())
		})
	}
}

func TestMediaURL_IsVideo(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "mp4 is video",
			url:  "https://example.com/video.mp4",
			want: true,
		},
		{
			name: "avi is video",
			url:  "https://example.com/video.avi",
			want: true,
		},
		{
			name: "jpg is not video",
			url:  "https://example.com/image.jpg",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, _ := NewMediaURL(tt.url)
			assert.Equal(t, tt.want, mu.IsVideo())
		})
	}
}

func TestMediaURL_IsAudio(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "mp3 is audio",
			url:  "https://example.com/song.mp3",
			want: true,
		},
		{
			name: "wav is audio",
			url:  "https://example.com/sound.wav",
			want: true,
		},
		{
			name: "jpg is not audio",
			url:  "https://example.com/image.jpg",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, _ := NewMediaURL(tt.url)
			assert.Equal(t, tt.want, mu.IsAudio())
		})
	}
}

func TestMediaURL_Equals(t *testing.T) {
	mu1, _ := NewMediaURL("https://example.com/image.jpg")
	mu2, _ := NewMediaURL("https://example.com/image.jpg")
	mu3, _ := NewMediaURL("https://example.com/different.jpg")

	assert.True(t, mu1.Equals(mu2))
	assert.False(t, mu1.Equals(mu3))
}

func TestMediaURL_String(t *testing.T) {
	url := "https://example.com/image.jpg"
	mu, _ := NewMediaURL(url)
	assert.Equal(t, url, mu.String())
}

func TestNewMessageText_Valid(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "simple text",
			text: "Hello World",
			want: "Hello World",
		},
		{
			name: "text with emojis",
			text: "Hello 👋 World 🌍",
			want: "Hello 👋 World 🌍",
		},
		{
			name: "text with special characters",
			text: "Olá! Como você está? 😊",
			want: "Olá! Como você está? 😊",
		},
		{
			name: "maximum length text",
			text: strings.Repeat("a", MaxTextLength),
			want: strings.Repeat("a", MaxTextLength),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt, err := NewMessageText(tt.text)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, mt.Value())
		})
	}
}

func TestNewMessageText_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr error
	}{
		{
			name:    "empty text",
			text:    "",
			wantErr: ErrTextEmpty,
		},
		{
			name:    "text too long",
			text:    strings.Repeat("a", MaxTextLength+1),
			wantErr: ErrTextTooLong,
		},
		{
			name:    "invalid UTF-8",
			text:    string([]byte{0xff, 0xfe, 0xfd}),
			wantErr: ErrTextInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMessageText(tt.text)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestMessageText_Length(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantLength int
	}{
		{
			name:       "simple text",
			text:       "Hello",
			wantLength: 5,
		},
		{
			name:       "text with emoji (counts as 1 char)",
			text:       "👋",
			wantLength: 1,
		},
		{
			name:       "text with multiple emojis",
			text:       "Hello 👋 World 🌍",
			wantLength: 15,
		},
		{
			name:       "portuguese text",
			text:       "Olá",
			wantLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt, err := NewMessageText(tt.text)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantLength, mt.Length())
		})
	}
}

func TestMessageText_Truncate(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxLength int
		want      string
	}{
		{
			name:      "no truncation needed",
			text:      "Hello",
			maxLength: 10,
			want:      "Hello",
		},
		{
			name:      "truncate with ellipsis",
			text:      "Hello World, this is a long message",
			maxLength: 15,
			want:      "Hello World,...",
		},
		{
			name:      "truncate very short",
			text:      "Hello World",
			maxLength: 3,
			want:      "Hel",
		},
		{
			name:      "truncate to zero",
			text:      "Hello",
			maxLength: 0,
			want:      "",
		},
		{
			name:      "truncate with emojis",
			text:      "Hello 👋 World 🌍 How are you?",
			maxLength: 12,
			want:      "Hello 👋 W...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt, err := NewMessageText(tt.text)
			assert.NoError(t, err)

			truncated := mt.Truncate(tt.maxLength)
			assert.Equal(t, tt.want, truncated.Value())
		})
	}
}

func TestMessageText_Contains(t *testing.T) {
	mt, _ := NewMessageText("Hello World, how are you?")

	tests := []struct {
		name   string
		substr string
		want   bool
	}{
		{
			name:   "contains word",
			substr: "World",
			want:   true,
		},
		{
			name:   "contains phrase",
			substr: "how are",
			want:   true,
		},
		{
			name:   "does not contain",
			substr: "xyz",
			want:   false,
		},
		{
			name:   "empty substring",
			substr: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mt.Contains(tt.substr)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestMessageText_Equals(t *testing.T) {
	mt1, _ := NewMessageText("Hello World")
	mt2, _ := NewMessageText("Hello World")
	mt3, _ := NewMessageText("Different")

	assert.True(t, mt1.Equals(mt2))
	assert.False(t, mt1.Equals(mt3))
}

func TestMessageText_IsEmpty(t *testing.T) {
	mt1, _ := NewMessageText("Hello")
	assert.False(t, mt1.IsEmpty())

	// Empty MessageText (created via zero value)
	mt2 := MessageText{}
	assert.True(t, mt2.IsEmpty())
}

func TestMessageText_String(t *testing.T) {
	text := "Hello World"
	mt, _ := NewMessageText(text)
	assert.Equal(t, text, mt.String())
}
