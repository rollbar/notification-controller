package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	eventv1 "github.com/fluxcd/pkg/apis/event/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// Mock Rollbar server
func newMockRollbarServer(t *testing.T, expectedPayload RollbarPayload) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var actualPayload RollbarPayload
		err := json.NewDecoder(r.Body).Decode(&actualPayload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("Expected %v", expectedPayload)
		fmt.Printf("Actual %v", actualPayload)
		assert.Equal(t, expectedPayload, actualPayload)
		w.WriteHeader(http.StatusOK)
	}))
}

func TestNewRollbar(t *testing.T) {
	tests := []struct {
		token       string
		environment string
		expectError bool
	}{
		{"valid_token", "production", false},
		{"", "production", true},
	}

	for _, test := range tests {
		rollbar, err := NewRollbar(test.token, test.environment, "", nil, "")
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, test.token, rollbar.Token)
			assert.Equal(t, test.environment, rollbar.Environment)
		}
	}
}

func TestRollbarPost(t *testing.T) {
	event := eventv1.Event{
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "rollbar-pod",
			Namespace: "default",
		},
		Severity:            eventv1.EventSeverityError,
		Message:             "An error occurred",
		ReportingController: "rollbar-controller",
	}

	expectedPayload := RollbarPayload{
		AccessToken: "test_token",
		Data: RollbarData{
			Environment: "test_env",
			Body: RollbarBody{
				Message: RollbarMessage{
					Body: event.Message,
				},
			},
			Level:     "error",
			Timestamp: time.Now().Unix(),
			Platform:  "linux",
			Server: RollbarServer{
				Host: "pod/rollbar-pod.default",
			},
		},
	}

	server := newMockRollbarServer(t, expectedPayload)
	defer server.Close()

	rollbar, err := NewRollbar("test_token", "test_env", "", nil, server.URL)
	require.NoError(t, err)

	err = rollbar.Post(context.Background(), event)
	require.NoError(t, err)
}

func TestRollbarPostNonErrorEvent(t *testing.T) {
	event := eventv1.Event{
		Severity: eventv1.EventSeverityInfo,
	}

	rollbar, err := NewRollbar("test_token", "test_env", "", nil, "")
	require.NoError(t, err)

	err = rollbar.Post(context.Background(), event)
	require.NoError(t, err)
}
