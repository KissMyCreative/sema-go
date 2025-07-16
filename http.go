package vexil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/jsnfwlr/sse/v2"
)

type Client struct {
	sse     *sse.Client
	address string
	logger  Logger
}

type Flag struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func NewClient(address string, baseClient *http.Client, logger Logger) (vexilClient *Client) {
	c := sse.NewClient(address + "/api/events")

	c.Connection = baseClient

	return &Client{
		address: address,
		sse:     c,
		logger:  logger,
	}
}

// FlagsToEnv uses the Client to make a HTTP request to get all the available
// flags from the Vexil server for the $environment and then updates the local
// environment variavles according to the flag data
func (c *Client) FlagsToEnv(environment string) (fault error) {
	return nil
}

// FlagsToFunc uses the Client to make a HTTP request to get all the available
// flags from the Vexil server for the $environment and then calls the $handler
// once for each flag, passing it the flag data.
func (c *Client) FlagsToFunc(environment string, handler func(flag Flag)) (fault error) {
	url, err := url.JoinPath(c.address, "api", "environment", environment, "flag")
	if err != nil {
		return fmt.Errorf("could not build url: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("could not build request: %w", err)
	}

	resp, err := c.sse.Connection.Do(req)
	if err != nil {
		return fmt.Errorf("could not make request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("error making request - %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not make request: %w", err)
	}

	var flags []Flag

	err = json.Unmarshal(b, &flags)
	if err != nil {
		return fmt.Errorf("could not handle response: %w", err)
	}

	for _, f := range flags {
		handler(f)
	}

	return nil
}

// EventsToEnv uses the Client to subscribe to updates to flags for $environment
// from the server, and updates the local environment variables according to the
// flag data.
// Note: this is non-blocking, if the calling application closes, the connection
// will be closed along with it.
func (c *Client) EventsToEnv(environment string) {
	err := c.FlagsToEnv(environment)
	if err != nil {
		fmt.Printf("error retrieving to flags: %v\n", err)
		return
	}

	err = c.sse.Subscribe(environment, func(msg *sse.Event) {
	})
	if err != nil {
		fmt.Printf("error subscribing to events: %v\n", err)
		return
	}
}

// EventsToFunc uses the Client to subscribe to updates to flags for $environment
// from the server, and calls the $handler function with the flag data.
// Note: this is non-blocking, if the calling application closes, the connection
// will be closed along with it.
func (c *Client) EventsToFunc(environment string, handler func(flag Flag)) {
	err := c.FlagsToFunc(environment, handler)
	if err != nil {
		fmt.Printf("error subscribing to events: %v\n", err)
		return
	}

	err = c.sse.Subscribe(environment, func(msg *sse.Event) {
		// Call the provided handler with the received message
		var f Flag

		_ = json.Unmarshal(msg.Data, &f)

		handler(f)
	})
	if err != nil {
		fmt.Printf("error subscribing to events: %v\n", err)
		return
	}
}
