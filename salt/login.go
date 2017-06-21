package salt

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-systemd/daemon"
)

type Event struct {
	Tag  string                 `json:"tag"`
	Data map[string]interface{} `json:"data"`
}

type LoginResponse struct {
	Eauth       string   `json:"eauth,omitempty"`
	Expire      float64  `json:"expire,omitempty"`
	Permissions []string `json:"perms,omitempty"`
	Start       float64  `json:"start,omitempty"`
	Token       string   `json:"token,omitempty"`
	User        string   `json:"user,omitempty"`
}

type SaltConnection struct {
	Token    string
	Expires  float64
	Response *http.Response
	Reader   *bufio.Reader
}

var connection SaltConnection

func login() {

	log.Printf("Logging into salt master...")
	baseUrl := os.Getenv("SALT_MASTER_ADDRESS")

	values := make(map[string]string)
	values["username"] = os.Getenv("SALT_EVENT_USERNAME")
	values["password"] = os.Getenv("SALT_EVENT_PASSWORD")
	values["eauth"] = "pam"

	body, _ := json.Marshal(values)

	url := baseUrl + "/login"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error making request: %s", err.Error())
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	log.Printf("Sending request...")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transport}

	connection.Response, err = client.Do(request)
	if err != nil {
		log.Printf("Error sending request: %s", err.Error())
	}

	responseBody := make(map[string][]LoginResponse)

	body, err = ioutil.ReadAll(connection.Response.Body)
	if err != nil {
		log.Printf("Error reading response: %s", err.Error())
	}

	err = json.Unmarshal(body, &responseBody)
	loginResponse := responseBody["return"][0]

	connection.Token = loginResponse.Token
	connection.Expires = loginResponse.Expire
	log.Printf("Success!")

	log.Printf("Subscribing to salt events...")

	url = baseUrl + "/events"
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error making request: %s", err.Error())
	}

	request.Header.Add("X-Auth-Token", connection.Token)
	connection.Response, err = client.Do(request)
	if err != nil {
		log.Printf("Error sending request: %s", err.Error())
	}

	connection.Reader = bufio.NewReader(connection.Response.Body)

	daemon.SdNotify(false, "READY=1")
}
