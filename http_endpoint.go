package gitbot

/*
inspired by https://github.com/kubernetes/test-infra/blob/master/prow/cmd/hook/server.go
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosminilie/gitbot/gitlabhook"
	"github.com/go-kit/kit/log"
	gitlab "github.com/xanzy/go-gitlab"
)

// Server implements http.Handler. It validates incoming GitLab webhooks and
// then dispatches them to the appropriate plugins.
type Server struct {
	Service Service
	Logger  log.Logger
}

// ServeHTTP validates an incoming webhook and invokes the service handler for them.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Header checks: It must be a POST with an event type and a signature.
	if r.Method != http.MethodPost {
		http.Error(w, "405 Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	eventType := r.Header.Get("X-Gitlab-Event")
	if eventType == "" {
		http.Error(w, "400 Bad Request: Missing X-Gitlab-Event Header", http.StatusBadRequest)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "500 Internal Server Error: Failed to read request body", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Event received. Have a nice day.")

	err = s.demuxEvent(eventType, payload)
	if err != nil {
		s.Logger.Log(
			"Caller", "ServeHTTP",
			"Action", "demuxEvent",
			"eventType", eventType,
			"Error", err,
		)
	}

}

func (s *Server) demuxEvent(eventType string, payload []byte) error {
	switch eventType {
	case "Merge Request Hook":
		var req gitlab.MergeEvent
		if err := json.Unmarshal(payload, &req); err != nil {
			return fmt.Errorf("failed to Unmarshal Merge Event with :%s raw body:%s", err, string(payload))

		}
		go s.Service.GitHook(s.Logger, req)
	case "Note Hook":
		var req gitlabhook.MergeRequestCommentEvent
		if err := json.Unmarshal(payload, &req); err != nil {
			return fmt.Errorf("Failed to Unmarshal MergeComment Event with :%s raw body:%s", err, string(payload))
		}

		go s.Service.GitHook(s.Logger, req)
	default:
		s.Logger.Log(
			"Caller", "demuxEvent",
			"EventType", eventType,
			"Result", "Unknow Event",
			//"Payload", string(payload),
		)

	}

	return nil
}
