package salt

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"

	"net/http"
	"os"

	"github.com/byuoitav/common/log"
	"github.com/coreos/go-systemd/daemon"
)

type loginResponse struct {
	Eauth       string   `json:"eauth,omitempty"`
	Expire      float64  `json:"expire,omitempty"`
	Permissions []string `json:"perms,omitempty"`
	Start       float64  `json:"start,omitempty"`
	Token       string   `json:"token,omitempty"`
	User        string   `json:"user,omitempty"`
}

type saltConnection struct {
	Token    string
	Expires  float64
	Response *http.Response
	Reader   *bufio.Reader
}

func login() (saltConnection, error) {

	var connection saltConnection

	log.L.Debugf("Logging into salt master...")
	baseURL := os.Getenv("SALT_MASTER_ADDRESS")

	values := make(map[string]string)
	values["username"] = os.Getenv("SALT_EVENT_USERNAME")
	values["password"] = os.Getenv("SALT_EVENT_PASSWORD")
	values["eauth"] = "pam"

	log.L.Debugf("trying to connect to %v, username %v", baseURL, values["username"])

	body, err := json.Marshal(values)
	if err != nil {
		log.L.Debugf("Error marshalling salt login body: %s", err.Error())
		return connection, err
	}

	url := baseURL + "/login"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.L.Debugf("Error creating request: %s", err.Error())
		return connection, err
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	log.L.Debugf("Sending request [%v]...", request)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transport}

	log.L.Debugf("ERROR DANGER DANGER 2")
	connection.Response, err = client.Do(request)
	log.L.Debugf("ERROR DANGER DANGER 3")
	if err != nil {
		log.L.Debugf("ERROR DANGER DANGER")
		log.L.Debugf("Error sending request: %s", err.Error())
		return connection, err
	}

	responseBody := make(map[string][]loginResponse)

	body, err = ioutil.ReadAll(connection.Response.Body)
	if err != nil {
		log.L.Debugf("Error reading response: %s", err.Error())
		return connection, err
	}

	err = json.Unmarshal(body, &responseBody)
	loginResponse := responseBody["return"][0]

	connection.Token = loginResponse.Token
	connection.Expires = loginResponse.Expire
	log.L.Debugf("Success!")

	log.L.Debugf("Subscribing to salt events...")

	url = baseURL + "/events"
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		log.L.Debugf("Error making request: %s", err.Error())
		return connection, err
	}

	request.Header.Add("X-Auth-Token", connection.Token)
	connection.Response, err = client.Do(request)
	if err != nil {
		log.L.Debugf("Error sending request: %s", err.Error())
		return connection, err
	}

	connection.Reader = bufio.NewReader(connection.Response.Body)

	daemon.SdNotify(false, "READY=1")

	return connection, nil
}
