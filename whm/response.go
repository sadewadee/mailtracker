package whm

import "strings"

// CPanelApiResponse - Response for cPanel API v3 via WHM proxy (legacy method)
type CPanelApiResponse struct {
	Func   string          `json:"func"`
	Result CPanelApiResult `json:"result"`
}

type CPanelApiResult struct {
	MetaData MetaData  `json:"metadata"`
	Warnings *[]string `json:"warnings"`
	Errors   *[]string `json:"errors"`
	Messages *[]string `json:"messages"`
	Data     int       `json:"data"`
	Status   int       `json:"status"`
}

func (a CPanelApiResult) ErrorMessage() string {
	if a.Errors == nil {
		return ""
	}

	return strings.Join(*a.Errors, ", ")
}

// UAPIResponse - Response for modern UAPI direct calls
type UAPIResponse struct {
	Func       string          `json:"func"`
	Module     string          `json:"module"`
	Status     int             `json:"status"`
	StatusMsg  string          `json:"statusmsg"`
	Errors     *[]string       `json:"errors"`
	Messages   *[]string       `json:"messages"`
	Warnings   *[]string       `json:"warnings"`
	Data       interface{}     `json:"data"`
	Metadata   UAPIMetadata    `json:"metadata"`
}

type UAPIMetadata struct {
	HTTPStatusCode int `json:"HTTPStatusCode"`
	Transformed    int `json:"transformed"`
}

func (r UAPIResponse) IsSuccess() bool {
	return r.Status == 1
}

func (r UAPIResponse) ErrorMessage() string {
	if r.Errors == nil || len(*r.Errors) == 0 {
		if r.StatusMsg != "" {
			return r.StatusMsg
		}
		return ""
	}
	return strings.Join(*r.Errors, ", ")
}

// ApiResponse - WHM API 1 response
type ApiResponse struct {
	Metadata MetaData `json:"metadata"`
	Data     Data     `json:"data"`
}

type MetaData struct {
	Version int    `json:"version"`
	Reason  string `json:"reason"`
	Result  int    `json:"result"`
	Command string `json:"command"`
}

type Data struct {
	Domains  []Domain  `json:"domains"`
	Accounts []Account `json:"acct"`
	UserData UserData  `json:"userdata"`

	Reason string `json:"reason"`
	Result string `json:"result"`
	Error  string `json:"error"`
}

// Configuration variables
var ApiUser = "root"
var ApiToken = ""
var ApiHost = "127.0.0.1"

// PreferModernUAPI - When true, try modern UAPI first before falling back to legacy
var PreferModernUAPI = true

var apiURI = "/json-api/"
var cpanelApiURL = apiURI + "cpanel"

var Log func(string, ...interface{})

// cPanelApiURL generates URL for cPanel API v3 via WHM proxy (legacy method)
func cPanelApiURL(module string, function string, user string) string {
	return apiURI + "cpanel?api.version=1" + "&cpanel_jsonapi_user=" + user +
		"&cpanel_jsonapi_apiversion=3" +
		"&cpanel_jsonapi_module=" + module +
		"&cpanel_jsonapi_func=" + function
}

// uapiURL generates URL for modern UAPI direct calls
func uapiURL(module string, function string) string {
	return "/execute/" + module + "/" + function
}
