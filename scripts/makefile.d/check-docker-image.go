package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// example
// $go run check-docker-image.go hub.docker.io/bluemir/wikinote:v1.0.0
func main() {
	if len(os.Args) < 2 {
		panic("not enough args")
	}

	registry, str := split2(os.Args[1], "/")
	image, tag := split2(str, ":")

	info := reqAuthHead(registry, image, tag)
	token := reqIssueToken(info)
	reqCheckImageTag(registry, image, tag, token)

}
func reqAuthHead(registry, image, tag string) authInfo {
	resp, err := http.Head(fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, image, tag))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	header := resp.Header.Get("Www-Authenticate")
	_, data := split2(header, " ")

	info := authInfo{}

	arr := strings.Split(data, ",")
	for _, str := range arr {
		k, v := split2(str, "=")
		switch k {
		case "realm":
			info.realm = strings.Trim(v, `"`)
		case "scope":
			info.scope = strings.Trim(v, `"`)
		case "service":
			info.service = strings.Trim(v, `"`)
		}
	}
	return info
}
func reqIssueToken(info authInfo) string {
	req, err := http.NewRequest(http.MethodGet, info.realm, nil)
	if err != nil {
		panic(err)
	}
	query := req.URL.Query()

	query.Add("scope", info.scope)
	query.Add("service", info.service)

	req.URL.RawQuery = query.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	res := struct {
		Token string `json:"token"`
	}{}

	if err := decoder.Decode(&res); err != nil {
		panic(err)
	}

	return res.Token
}
func reqCheckImageTag(registry, image, tag, token string) {
	resp, err := http.Head(fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, image, tag))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		panic("image not exist")
	}
	fmt.Printf("image exist")
}
func split2(str string, sep string) (string, string) {
	arr := strings.SplitN(str, sep, 2)
	if len(arr) < 2 {
		return arr[0], ""
	}

	return arr[0], arr[1]
}

type authInfo struct {
	realm   string
	scope   string
	service string
}
