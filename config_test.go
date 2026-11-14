package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const configTemplate = `{"signing_key":"sk","sid":"s id","ecode":"ec","partner":"pi","app_key":"ak","device_id":"phone1","package_name":"com.example.app","api_host":%q,"ch_key":"ck","bridge_port":1984,"cameras":[{"camera_id":"cam1","camera_name":"Nursery"},{"camera_id":"cam2"}]}`

const urlTemplate = "tuya://%s?device_id=cam1&sid=s+id&ecode=ec&partner_identity=pi&app_key=ak&signing_key=sk&ch_key=ck&package_name=com.example.app&client_device_id=phone1"

func TestSessionURL(t *testing.T) {
	configWithHost := func(host string) string {
		return fmt.Sprintf(configTemplate, host)
	}
	configWithTalkback := func(talkback string) string {
		return strings.Replace(configWithHost("a1.tuyaus.com"), `"ch_key":"ck",`, `"ch_key":"ck","talkback":`+talkback+`,`, 1)
	}
	configWithoutField := func(field string) string {
		return strings.Replace(configWithHost("a1.tuyaus.com"), `"`+field+`":`, `"x_`+field+`":`, 1)
	}

	tests := []struct {
		name      string
		config    string
		camera    string
		flagChKey string
		wantURL   string
		wantError string
	}{
		{"bare host", configWithHost("a1.tuyaus.com"), "cam1", "", fmt.Sprintf(urlTemplate, "a1.tuyaus.com"), ""},
		{"https host with path", configWithHost("https://a1.tuyaus.com/api.json"), "cam1", "", fmt.Sprintf(urlTemplate, "a1.tuyaus.com"), ""},
		{"host with port", configWithHost("a1.tuyaus.com:8443"), "cam1", "", fmt.Sprintf(urlTemplate, "a1.tuyaus.com:8443"), ""},
		{"query is dropped", configWithHost("a1.tuyaus.com?sid=evil"), "cam1", "", fmt.Sprintf(urlTemplate, "a1.tuyaus.com"), ""},
		{"fragment is dropped", configWithHost("https://a1.tuyaus.com#x"), "cam1", "", fmt.Sprintf(urlTemplate, "a1.tuyaus.com"), ""},
		{"ampersand in host", configWithHost("a1.tuyaus.com&sid=evil"), "cam1", "", "", "invalid field api_host"},
		{"scheme only", configWithHost("https://"), "cam1", "", "", "invalid field api_host"},
		{"talkback true", configWithTalkback("true"), "cam1", "", fmt.Sprintf(urlTemplate, "a1.tuyaus.com") + "&backchannel=1", ""},
		{"talkback false", configWithTalkback("false"), "cam1", "", fmt.Sprintf(urlTemplate, "a1.tuyaus.com"), ""},
		{"talkback missing", configWithHost("a1.tuyaus.com"), "cam1", "", fmt.Sprintf(urlTemplate, "a1.tuyaus.com"), ""},
		{"talkback not a boolean", configWithTalkback(`"yes"`), "cam1", "", "", "invalid field talkback"},
		{"ch_key from flag only", configWithoutField("ch_key"), "cam1", "fk", strings.Replace(fmt.Sprintf(urlTemplate, "a1.tuyaus.com"), "ch_key=ck", "ch_key=fk", 1), ""},
		{"ch_key JSON wins over flag", configWithHost("a1.tuyaus.com"), "cam1", "fk", fmt.Sprintf(urlTemplate, "a1.tuyaus.com"), ""},
		{"empty host uses default", configWithHost(""), "cam1", "", fmt.Sprintf(urlTemplate, "a1.tuyaeu.com"), ""},
		{"unknown camera", configWithHost("a1.tuyaus.com"), "cam9", "", "", "unknown camera cam9"},
		{"missing signing_key", configWithoutField("signing_key"), "cam1", "", "", "missing field signing_key"},
		{"missing sid", configWithoutField("sid"), "cam1", "", "", "missing field sid"},
		{"missing ecode", configWithoutField("ecode"), "cam1", "", "", "missing field ecode"},
		{"missing partner", configWithoutField("partner"), "cam1", "", "", "missing field partner"},
		{"missing app_key", configWithoutField("app_key"), "cam1", "", "", "missing field app_key"},
		{"missing device_id", configWithoutField("device_id"), "cam1", "", "", "missing field device_id"},
		{"missing package_name", configWithoutField("package_name"), "cam1", "", "", "missing field package_name"},
		{"missing ch_key", configWithoutField("ch_key"), "cam1", "", "", "missing field ch_key"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir() + "/session.json"
			if err := os.WriteFile(path, []byte(test.config), 0o600); err != nil {
				t.Fatal(err)
			}
			gotURL, gotError := "", ""
			session, err := Load(path, test.flagChKey)
			if err == nil {
				gotURL, err = session.URL(test.camera)
			}
			if err != nil {
				gotError = err.Error()
			}
			if gotURL != test.wantURL || gotError != test.wantError {
				t.Errorf("got (%q, %q), want (%q, %q)", gotURL, gotError, test.wantURL, test.wantError)
			}
		})
	}
}
