package cloudprint

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-wheels/tk/randutil"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

const (
	apiURL = "https://open-api.10ss.net/"

	authorizeURL        = apiURL + "oauth/oauth"
	addPrinterURL       = apiURL + "printer/addprinter"
	deletePrinterURL    = apiURL + "printer/deleteprinter"
	printURL            = apiURL + "print/index"
	getPrinterStatusURL = apiURL + "printer/getprintstatus"
)

type APIResponse struct {
	Error            string          `json:"error"`
	ErrorDescription string          `json:"error_description"`
	Body             json.RawMessage `json:"body"`
}

type APIClient struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	tokenStore   TokenStore
}

func NewAPIClient(clientID, clientSecret string, tokenStore TokenStore) *APIClient {
	return &APIClient{
		httpClient:   &http.Client{},
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenStore:   tokenStore,
	}
}

func (c *APIClient) GetPrinterStatus(machineCode string) (apiResp *APIResponse, err error) {
	token, err := c.tokenStore.Get(c.clientID)
	if err != nil {
		return
	}
	data := make(url.Values)
	data.Set("access_token", token)
	data.Set("machine_code", machineCode)

	apiResp, err = c.PostForm(getPrinterStatusURL, data)
	return
}

func (c *APIClient) Print(machineCode, content string) (apiResp *APIResponse, err error) {
	token, err := c.tokenStore.Get(c.clientID)
	if err != nil {
		return
	}
	data := make(url.Values)
	data.Set("access_token", token)
	data.Set("machine_code", machineCode)
	data.Set("content", content)
	data.Set("origin_id", randutil.RandAlnumStr(32))

	apiResp, err = c.PostForm(printURL, data)
	return
}

func (c *APIClient) DeletePrinter(machineCode string) (apiResp *APIResponse, err error) {
	token, err := c.tokenStore.Get(c.clientID)
	if err != nil {
		return
	}
	data := make(url.Values)
	data.Set("access_token", token)
	data.Set("machine_code", machineCode)

	apiResp, err = c.PostForm(deletePrinterURL, data)
	return
}

func (c *APIClient) AddPrinter(machineCode, msign string) (apiResp *APIResponse, err error) {
	token, err := c.tokenStore.Get(c.clientID)
	if err != nil {
		return
	}
	data := make(url.Values)
	data.Set("machine_code", machineCode)
	data.Set("msign", msign)
	data.Set("access_token", token)

	apiResp, err = c.PostForm(addPrinterURL, data)
	return
}

func (c *APIClient) Authorize() (err error) {
	token, err := c.tokenStore.Get(c.clientID)
	if err != nil {
		return
	}
	if token != "" {
		return
	}

	data := make(url.Values)
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "all")

	apiResp, err := c.PostForm(authorizeURL, data)
	if err != nil {
		return
	}
	token = gjson.GetBytes(apiResp.Body, "access_token").String()
	err = c.tokenStore.Set(c.clientID, token)
	return
}

func (c *APIClient) PostForm(url string, data url.Values) (apiResp *APIResponse, err error) {
	timestamp := TimestampStr()

	data.Set("client_id", c.clientID)
	data.Set("sign", c.Sign(timestamp))
	data.Set("id", RequestID())
	data.Set("timestamp", timestamp)

	resp, err := c.httpClient.PostForm(url, data)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	apiResp = &APIResponse{}

	err = json.NewDecoder(resp.Body).Decode(apiResp)
	if err != nil {
		return
	}
	if apiResp.Error != "0" || apiResp.ErrorDescription != "success" {
		err = fmt.Errorf("cloudprint: %s (code: %s)", apiResp.ErrorDescription, apiResp.Error)
	}
	return
}

func (c *APIClient) Sign(timestamp string) string {
	s := c.clientID + timestamp + c.clientSecret
	w := md5.New()
	io.WriteString(w, s)
	return hex.EncodeToString(w.Sum(nil))
}

func RequestID() string {
	return uuid.NewString()
}

func TimestampStr() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
