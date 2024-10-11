package mockserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

func Api(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("\"description\":\"AWX REST API\""))
}

func Rpc2Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Println("receive post request!!!")
		fmt.Println("head:", r.Header)
		respBody, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		resp := map[string]any{}
		json.Unmarshal(respBody, &resp)

		resp33045 := map[string]any{}
		json.Unmarshal([]byte(`{"id": 1, "method": "global.login", "params": { "userName": "admin", "ipAddr": "127.0.0.1", "loginType": "Loopback", "clientType": "Local", "authorityType": "Default", "passwordType": "Default", "password": "Default"}, "session": 0}`), &resp33045)
		resp33044 := map[string]any{}
		json.Unmarshal([]byte(`{"id": 1, "method": "global.login", "params": {"authorityType": "Default", "clientType": "NetKeyboard", "loginType": "Direct", "password": "Not Used", "passwordType": "Default", "userName": "admin"}, "session": 0}`), &resp33044)

		if reflect.DeepEqual(resp, resp33045) {
			w.WriteHeader(200)
			w.Write([]byte("{\"result\":false,\"id\":1,\"params\":\"\",\"session\":\"\",\"mac\":\"\"}"))
		} else if reflect.DeepEqual(resp, resp33044) {
			w.WriteHeader(200)
			w.Write([]byte("{\"result\":true,\"id\":1,\"params\":\"\",\"session\":\"abcdefghijklmnopqistuvwxyz\",\"mac\":\"\"}"))
		} else {
			w.WriteHeader(200)
			w.Write([]byte("none"))
		}
	}
}

func HeadlessLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Click to Request URL</title>
<script>
function requestURL() {
const url = "http://192.168.1.177:9080/headless"; // 指定的网址
window.location.href = url; // 导航到指定的网址
}
</script>
</head>
<body>
<input type="button" id="requestUrlButton" onclick="requestURL()" value="Request URL"/>
<label for="requestUrlButton">点击这里请，求网址</label>
</body>
</html>`))
}

func Headless(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("You have logged in as admin"))
}

func CVE_2020_7057(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	fmt.Println("username:", username)
	if username == "test" {
		w.WriteHeader(200)
		w.Write([]byte("<isIrreversible>true</isIrreversible>"))
		return
	}
	w.WriteHeader(500)
}
