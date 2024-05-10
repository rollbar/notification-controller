package notifier

import (
	"context"
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
func NewRollbar(token string, environment string, url string) (*Rollbar, error) {
	fmt.Printf("Starting Rollbar notifier!!!! :::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::\n")
	if token == "" {
		return nil, fmt.Errorf("Rollbar token cannot be empty")
	} else {
		// TODO Disable this
		fmt.Printf("RollbarNotifier ::: Received token %s\n", token)
	}

	if url == "" {
		url = defaultRollbarURL
	}

	fmt.Printf("RollbarNotifier ::: Posting to url %s\n", url)

	return &Rollbar{
		Token:       token,
		Environment: environment,
		RollbarUrl:  url,
	}, nil
}

// Post Rollbar message
func (r *Rollbar) Post(ctx context.Context, event eventv1.Event) error {
	// Only post error events
	// if event.Severity != eventv1.EventSeverityError {
	// 	return nil
	// }

	fmt.Printf("RollbarNotifier ::: Event:  %v\n", event)
	fmt.Printf("RollbarNotifier ::: Rollbar Object %v\n", r)
	fmt.Printf("RollbarNotifier ::: Severity %v\n", event.Severity)

	host := fmt.Sprintf("%s/%s.%s", strings.ToLower(event.InvolvedObject.Kind), event.InvolvedObject.Name, event.InvolvedObject.Namespace)
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

	err := postMessage(ctx, r.RollbarUrl, "", nil, payload, func(request *retryablehttp.Request) {
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("accept", "application/json")
		request.Header.Add("X-Rollbar-Access-Token", r.Token)
	})
	if err != nil {
		return fmt.Errorf("postMessage failed: %w\n", err)
	}
	return nil
}
