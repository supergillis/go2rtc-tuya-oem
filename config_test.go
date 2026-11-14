package main

import (
	"os"
	"strings"
	"testing"
)

const fullConfig = `{"signing_key":"sk","sid":"s id","ecode":"ec","partner":"pi","app_key":"ak","device_id":"phone1","package_name":"com.example.app","api_host":%q,"ch_key":"ck","talkback":true,"bridge_port":1984,"cameras":[{"camera_id":"cam1","camera_name":"Nursery"},{"camera_id":"cam2"}]}`

const fullURL = "tuya://a1.tuyaus.com?device_id=cam1&sid=s+id&ecode=ec&partner_identity=pi&app_key=ak&signing_key=sk&ch_key=ck&package_name=com.example.app&client_device_id=phone1"

func TestSessionURL(t *testing.T) {
	withoutField := func(field string) string {
		return strings.Replace(strings.Replace(fullConfig, `"`+field+`":`, `"x_`+field+`":`, 1), "%q", `"a1.tuyaus.com"`, 1)
	}
	withHost := func(host string) string {
		return strings.Replace(fullConfig, "%q", `"`+host+`"`, 1)
	}

	tests := []struct {
		name      string
		config    string
		camera    string
		wantURL   string
		wantError string
	}{
		{"bare host", withHost("a1.tuyaus.com"), "cam1", fullURL, ""},
		{"https host with path", withHost("https://a1.tuyaus.com/api.json"), "cam1", fullURL, ""},
		{"empty host uses default", withHost(""), "cam1", strings.Replace(fullURL, "a1.tuyaus.com", "a1.tuyaeu.com", 1), ""},
		{"unknown camera", withHost("a1.tuyaus.com"), "cam9", "", "unknown camera cam9"},
		{"missing signing_key", withoutField("signing_key"), "cam1", "", "missing field signing_key"},
		{"missing sid", withoutField("sid"), "cam1", "", "missing field sid"},
		{"missing ecode", withoutField("ecode"), "cam1", "", "missing field ecode"},
		{"missing partner", withoutField("partner"), "cam1", "", "missing field partner"},
		{"missing app_key", withoutField("app_key"), "cam1", "", "missing field app_key"},
		{"missing device_id", withoutField("device_id"), "cam1", "", "missing field device_id"},
		{"missing package_name", withoutField("package_name"), "cam1", "", "missing field package_name"},
		{"missing ch_key", withoutField("ch_key"), "cam1", "", "missing field ch_key"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir() + "/session.json"
			if err := writeFile(path, test.config); err != nil {
				t.Fatal(err)
			}
			session, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			gotURL, err := session.URL(test.camera)
			gotError := ""
			if err != nil {
				gotError = err.Error()
			}
			if gotURL != test.wantURL || gotError != test.wantError {
				t.Errorf("got (%q, %q), want (%q, %q)", gotURL, gotError, test.wantURL, test.wantError)
			}
		})
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o600)
}
