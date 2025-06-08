package config

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

var client *http.Client

type Versions struct {
	Name       string `json:"name"`
	TarballUrl string `json:"tarball_url"`
}

func init() {
	client = &http.Client{Timeout: time.Second * 10}
}

func checkVersion(v string) (bool, error) {
	var versions []Versions

	res, err := client.Get("https://api.github.com/repos/kubernetes/kubernetes/tags")
	if err != nil {
		return false, err
	}

	body := res.Body
	defer body.Close()

	bytes, err := io.ReadAll(body)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(bytes, &versions)
	if err != nil {
		return false, err
	}

	for _, version := range versions {
		if strings.ToLower(v) == strings.ToLower(version.Name) {
			return true, nil
		}
	}

	return false, nil
}
