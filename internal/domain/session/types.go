package session

import "errors"

// Status representa o status de uma sessão.
type Status string

const (
	StatusActive         Status = "active"
	StatusEnded          Status = "ended"
	StatusExpired        Status = "expired"
	StatusManuallyClosed Status = "manually_closed"
)

func (s Status) String() string {
	return string(s)
}

// IsValid verifica se o status é válido.
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusEnded, StatusExpired, StatusManuallyClosed:
		return true
	default:
		return false
	}
}

// EndReason representa o motivo de encerramento.
type EndReason string

const (
	ReasonInactivityTimeout EndReason = "inactivity_timeout"
	ReasonManualClose       EndReason = "manual_close"
	ReasonContactRequest    EndReason = "contact_request"
	ReasonAgentClose        EndReason = "agent_close"
	ReasonSystemClose       EndReason = "system_close"
)

func (r EndReason) String() string {
	return string(r)
}

// Sentiment representa o sentimento da conversa.
type Sentiment string

const (
	SentimentPositive Sentiment = "positive"
	SentimentNeutral  Sentiment = "neutral"
	SentimentNegative Sentiment = "negative"
	SentimentMixed    Sentiment = "mixed"
)

func (s Sentiment) String() string {
	return string(s)
}

// ParseStatus converte string para Status.
func ParseStatus(s string) (Status, error) {
	status := Status(s)
	if !status.IsValid() {
		return "", errors.New("invalid status")
	}
	return status, nil
}

// ParseEndReason converte string para EndReason.
func ParseEndReason(s string) (EndReason, error) {
	reason := EndReason(s)
	switch reason {
	case ReasonInactivityTimeout, ReasonManualClose, ReasonContactRequest, ReasonAgentClose, ReasonSystemClose:
		return reason, nil
	default:
		return "", errors.New("invalid end reason")
	}
}

// ParseSentiment converte string para Sentiment.
func ParseSentiment(s string) (Sentiment, error) {
	sentiment := Sentiment(s)
	switch sentiment {
	case SentimentPositive, SentimentNeutral, SentimentNegative, SentimentMixed:
		return sentiment, nil
	default:
		return "", errors.New("invalid sentiment")
	}
}
