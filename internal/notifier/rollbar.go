package notifier

import (
	"context"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	eventv1 "github.com/fluxcd/pkg/apis/event/v1beta1"
	"github.com/hashicorp/go-retryablehttp"
)

// Rollbar holds the access token and environment information
type Rollbar struct {
	Token       string
	Environment string
	ProxyURL    string
	CertPool    *x509.CertPool
	RollbarUrl  string
}

// RollbarPayload represents the payload to send to Rollbar
type RollbarPayload struct {
	AccessToken string      `json:"access_token"`
	Data        RollbarData `json:"data"`
}

type RollbarData struct {
	Environment string        `json:"environment"`
	Body        RollbarBody   `json:"body"`
	Level       string        `json:"level"`
	Timestamp   int64         `json:"timestamp"`
	Platform    string        `json:"platform"`
	Server      RollbarServer `json:"server"`
}

type RollbarBody struct {
	Message RollbarMessage `json:"message"`
}

type RollbarMessage struct {
	Body string `json:"body"`
}

type RollbarServer struct {
	Host string `json:"host"`
}

const defaultRollbarURL = "https://api.rollbar.com/api/1/item/"

// NewRollbar validates the Rollbar token and returns a Rollbar object
func NewRollbar(token string, environment string, proxyURL string, certPool *x509.CertPool, url string) (*Rollbar, error) {
	if token == "" {
		return nil, fmt.Errorf("Rollbar token cannot be empty")
	}

	if url == "" {
		url = defaultRollbarURL
	}

	return &Rollbar{
		Token:       token,
		Environment: environment,
		ProxyURL:    proxyURL,
		CertPool:    certPool,
		RollbarUrl:  url,
	}, nil
}

// Post Rollbar message
func (r *Rollbar) Post(ctx context.Context, event eventv1.Event) error {
	// Only post error events
	fmt.Printf("Context: %v", ctx)
	fmt.Printf("Event: %v", event)
	if event.Severity != eventv1.EventSeverityError {
		return nil
	}

	host := fmt.Sprintf("%s/%s.%s", strings.ToLower(event.InvolvedObject.Kind), event.InvolvedObject.Name, event.InvolvedObject.Namespace)
	fmt.Printf("HOST: %v", host)
	payload := RollbarPayload{
		AccessToken: r.Token,
		Data: RollbarData{
			Environment: r.Environment,
			Body: RollbarBody{
				Message: RollbarMessage{
					Body: event.Message,
				},
			},
			Level:     "error",
			Timestamp: time.Now().Unix(),
			Platform:  "linux",
			Server: RollbarServer{
				Host: host,
			},
		},
	}

	err := postMessage(ctx, r.RollbarUrl, r.ProxyURL, r.CertPool, payload, func(request *retryablehttp.Request) {
		request.Header.Add("Content-Type", "application/json")
	})
	if err != nil {
		return fmt.Errorf("postMessage failed: %w", err)
	}
	return nil
}
