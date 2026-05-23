package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const defaultBase = "http://localhost:8080"

func main() {
	base := os.Getenv("HTTP_BASE")
	if base == "" {
		base = defaultBase
	}

	client := &http.Client{Timeout: 15 * time.Second}

	var health map[string]any
	if err := getJSON(client, base+"/health", "", nil, &health); err != nil {
		fail("health", err)
	}
	fmt.Println("health ok")

	var login apitypesAuth
	if err := postJSON(client, base+"/api/v1/auth/guest-login", "", map[string]string{
		"deviceId":      "http-smoke-device",
		"clientVersion": "1.0.0",
	}, &login); err != nil {
		fail("guest-login", err)
	}
	if login.SessionToken == "" {
		fail("guest-login", fmt.Errorf("missing session token"))
	}
	fmt.Printf("guest login player=%d\n", login.PlayerID)

	auth := "Bearer " + login.SessionToken

	var claim map[string]any
	if err := postJSON(client, base+"/api/v1/idle/claim", auth, map[string]int64{
		"claimedThroughMs": time.Now().UnixMilli(),
	}, &claim); err != nil {
		fail("idle claim", err)
	}
	fmt.Println("idle claim ok")

	var push map[string]any
	if err := postJSON(client, base+"/api/v1/stages/push", auth, map[string]int32{
		"targetStageIndex": login.StageState.StageIndex,
	}, &push); err != nil {
		fail("stage push", err)
	}
	fmt.Println("stage push ok")

	var chest map[string]any
	if err := postJSON(client, base+"/api/v1/equipment/chests/open", auth, map[string]int32{
		"count": 1,
	}, &chest); err != nil {
		fail("chest open", err)
	}
	fmt.Println("chest open ok")

	fmt.Println("http smoke passed")
}

type apitypesAuth struct {
	SessionToken string `json:"sessionToken"`
	PlayerID     int64  `json:"playerId"`
	StageState   struct {
		StageIndex int32 `json:"stageIndex"`
	} `json:"stageState"`
}

func getJSON(client *http.Client, url, auth string, body any, out any) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(http.MethodGet, url, r)
	if err != nil {
		return err
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decode(resp, out)
}

func postJSON(client *http.Client, url, auth string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decode(resp, out)
}

func decode(resp *http.Response, out any) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(data))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

func fail(step string, err error) {
	fmt.Fprintf(os.Stderr, "%s failed: %v\n", step, err)
	os.Exit(1)
}
