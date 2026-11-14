package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"unicode"
)

// Same default as havent, so a session saved without api_host keeps working.
const defaultAPIHost = "a1.tuyaeu.com"

// Session holds the havent bridge config fields needed to build a go2rtc URL.
// Unrelated fields such as bridge_port are ignored on decode.
type Session struct {
	SigningKey  string   `json:"signing_key"`
	SID         string   `json:"sid"`
	Ecode       string   `json:"ecode"`
	Partner     string   `json:"partner"`
	AppKey      string   `json:"app_key"`
	DeviceID    string   `json:"device_id"`
	PackageName string   `json:"package_name"`
	APIHost     string   `json:"api_host"`
	ChKey       string   `json:"ch_key"`
	Talkback    bool     `json:"talkback"`
	Cameras     []Camera `json:"cameras"`
}

type Camera struct {
	ID string `json:"camera_id"`
}

func Load(path string) (Session, error) {
	var session Session
	data, err := os.ReadFile(path)
	if err != nil {
		return session, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &session); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) {
			return session, fmt.Errorf("invalid field %s", typeError.Field)
		}
		return session, fmt.Errorf("parse config json at offset %d", jsonErrorOffset(err))
	}
	return session, nil
}

// jsonErrorOffset avoids echoing the decoder message, which can quote file content.
func jsonErrorOffset(err error) int64 {
	switch typed := err.(type) {
	case *json.SyntaxError:
		return typed.Offset
	case *json.UnmarshalTypeError:
		return typed.Offset
	}
	return 0
}

// normalizeAPIHost keeps only host and port, so a stored scheme, path, query or
// fragment can never add parameters to the printed URL.
func normalizeAPIHost(rawHost string) (string, error) {
	trimmed := strings.TrimSpace(rawHost)
	if trimmed == "" {
		return defaultAPIHost, nil
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" || strings.ContainsFunc(parsed.Host, isInvalidHostRune) {
		return "", errors.New("invalid field api_host")
	}
	return parsed.Host, nil
}

func isInvalidHostRune(character rune) bool {
	return !unicode.IsLetter(character) && !unicode.IsDigit(character) && !strings.ContainsRune(".-:[]", character)
}

// URL builds the tuya:// URL for one camera. Errors name fields, never values.
func (session Session) URL(cameraID string) (string, error) {
	required := []struct{ name, value string }{
		{"signing_key", session.SigningKey},
		{"sid", session.SID},
		{"ecode", session.Ecode},
		{"partner", session.Partner},
		{"app_key", session.AppKey},
		{"device_id", session.DeviceID},
		{"package_name", session.PackageName},
		{"ch_key", session.ChKey},
	}
	for _, field := range required {
		if field.value == "" {
			return "", fmt.Errorf("missing field %s", field.name)
		}
	}
	found := false
	for _, camera := range session.Cameras {
		if camera.ID == cameraID {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("unknown camera %s", cameraID)
	}

	// Built by hand to keep the parameter order go2rtc documents.
	params := []struct{ name, value string }{
		{"device_id", cameraID},
		{"sid", session.SID},
		{"ecode", session.Ecode},
		{"partner_identity", session.Partner},
		{"app_key", session.AppKey},
		{"signing_key", session.SigningKey},
		{"ch_key", session.ChKey},
		{"package_name", session.PackageName},
		{"client_device_id", session.DeviceID},
	}
	apiHost, err := normalizeAPIHost(session.APIHost)
	if err != nil {
		return "", err
	}
	pairs := make([]string, 0, len(params))
	for _, param := range params {
		pairs = append(pairs, param.name+"="+url.QueryEscape(param.value))
	}
	if session.Talkback {
		pairs = append(pairs, "backchannel=1")
	}
	return "tuya://" + apiHost + "?" + strings.Join(pairs, "&"), nil
}
