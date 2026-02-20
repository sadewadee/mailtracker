package whm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// SuspendEmail suspends outgoing email for the given email address.
// It tries modern UAPI first, then falls back to legacy WHM proxy method.
func SuspendEmail(email string) error {
	Log("Suspending %s", email)

	domain := email[strings.Index(email, "@")+1:]
	info, err := UserDataInfo(domain)
	if err != nil {
		Log("UserDataInfo error: %v", err)
		return err
	}

	// Try modern UAPI first if enabled
	if PreferModernUAPI {
		err := suspendEmailModernUAPI(email, info.User)
		if err == nil {
			return nil
		}
		Log("Modern UAPI failed, falling back to legacy: %v", err)
	}

	// Fallback to legacy WHM proxy method
	return suspendEmailLegacy(email, info.User)
}

// suspendEmailModernUAPI uses modern UAPI via cPanel port 2083
func suspendEmailModernUAPI(email, cpanelUser string) error {
	Log("Trying modern UAPI for suspend_outgoing")

	endpoint := uapiURL("Email", "suspend_outgoing") + "?email=" + url.QueryEscape(email)
	Log("UAPI endpoint: %s", endpoint)

	conn, err := CPanelDialer()
	if err != nil {
		return fmt.Errorf("CPanelDialer error: %v", err)
	}
	defer conn.Close()

	clientConn := httputil.NewClientConn(conn, nil)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("request creation error: %v", err)
	}

	// Modern UAPI uses "cpanel" auth header format with cPanel username
	req.Header.Set("Authorization", fmt.Sprintf("cpanel %s:%s", cpanelUser, ApiToken))

	resp, err := clientConn.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body error: %v", err)
	}

	Log("UAPI response: %s", string(body))

	var record UAPIResponse
	if err := json.Unmarshal(body, &record); err != nil {
		return fmt.Errorf("json unmarshal error: %v", err)
	}

	if record.IsSuccess() {
		Log("Modern UAPI suspend_outgoing successful")
		return nil
	}

	return fmt.Errorf("UAPI error: %s", record.ErrorMessage())
}

// suspendEmailLegacy uses legacy WHM proxy to cPanel API v3
func suspendEmailLegacy(email, cpanelUser string) error {
	Log("Using legacy WHM proxy for suspend_outgoing")

	urlString := cPanelApiURL("Email", "suspend_outgoing", cpanelUser) + "&email=" + url.QueryEscape(email)
	Log("calling: %s", urlString)
	time.Sleep(1000 * time.Millisecond)

	conn, err := WHMDialer()
	if err != nil {
		return err
	}
	defer conn.Close()
	clientConn := httputil.NewClientConn(conn, nil)

	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("whm %s:%s", ApiUser, ApiToken))

	resp, err := clientConn.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	Log("Reading body")
	body, err := ioutil.ReadAll(resp.Body)
	Log("called: %s", body)
	if err != nil {
		return err
	}
	Log("result %#v", string(body))

	var record CPanelApiResponse
	if err := json.Unmarshal(body, &record); err != nil {
		return err
	}

	if record.Result.Status == 1 {
		return nil
	}

	Log("metadata: %#v", record.Result.MetaData)
	return fmt.Errorf(record.Result.ErrorMessage())
}

// UnSuspendEmail unsuspends outgoing email for the given email address.
// It tries modern UAPI first, then falls back to legacy WHM proxy method.
func UnSuspendEmail(email string) error {
	Log("UnSuspending: %s", email)

	domain := email[strings.Index(email, "@")+1:]
	info, err := UserDataInfo(domain)
	if err != nil {
		return err
	}

	// Try modern UAPI first if enabled
	if PreferModernUAPI {
		err := unsuspendEmailModernUAPI(email, info.User)
		if err == nil {
			return nil
		}
		Log("Modern UAPI failed, falling back to legacy: %v", err)
	}

	// Fallback to legacy WHM proxy method
	return unsuspendEmailLegacy(email, info.User)
}

// unsuspendEmailModernUAPI uses modern UAPI via cPanel port 2083
func unsuspendEmailModernUAPI(email, cpanelUser string) error {
	Log("Trying modern UAPI for unsuspend_outgoing")

	endpoint := uapiURL("Email", "unsuspend_outgoing") + "?email=" + url.QueryEscape(email)
	Log("UAPI endpoint: %s", endpoint)

	conn, err := CPanelDialer()
	if err != nil {
		return fmt.Errorf("CPanelDialer error: %v", err)
	}
	defer conn.Close()

	clientConn := httputil.NewClientConn(conn, nil)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("request creation error: %v", err)
	}

	// Modern UAPI uses "cpanel" auth header format with cPanel username
	req.Header.Set("Authorization", fmt.Sprintf("cpanel %s:%s", cpanelUser, ApiToken))

	resp, err := clientConn.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body error: %v", err)
	}

	Log("UAPI response: %s", string(body))

	var record UAPIResponse
	if err := json.Unmarshal(body, &record); err != nil {
		return fmt.Errorf("json unmarshal error: %v", err)
	}

	if record.IsSuccess() {
		Log("Modern UAPI unsuspend_outgoing successful")
		return nil
	}

	return fmt.Errorf("UAPI error: %s", record.ErrorMessage())
}

// unsuspendEmailLegacy uses legacy WHM proxy to cPanel API v3
func unsuspendEmailLegacy(email, cpanelUser string) error {
	Log("Using legacy WHM proxy for unsuspend_outgoing")

	urlString := cPanelApiURL("Email", "unsuspend_outgoing", cpanelUser) + "&email=" + url.QueryEscape(email)

	conn, err := WHMDialer()
	if err != nil {
		return err
	}
	defer conn.Close()
	clientConn := httputil.NewClientConn(conn, nil)

	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("whm %s:%s", ApiUser, ApiToken))

	resp, err := clientConn.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	Log("result %#v", string(body))

	var record CPanelApiResponse
	if err := json.Unmarshal(body, &record); err != nil {
		return err
	}

	if record.Result.Status == 1 {
		return nil
	}

	Log("metadata: %#v", record.Result.MetaData)
	return fmt.Errorf(record.Result.ErrorMessage())
}

//================
// WHM API 1 Account-level functions (suspend entire account, not just email)

func SuspendAccountByEmail(email string) error {
	Log("Suspending account: %s", email)

	domain := email[strings.Index(email, "@")+1:]
	info, err := UserDataInfo(domain)
	if err != nil {
		return err
	}

	urlString := apiURI + "suspend_outgoing_email?api.version=1&user=" + url.QueryEscape(info.User)
	conn, err := WHMDialer()
	if err != nil {
		return err
	}
	defer conn.Close()
	clientConn := httputil.NewClientConn(conn, nil)

	Log("calling: %s", urlString)
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("whm %s:%s", ApiUser, ApiToken))

	resp, err := clientConn.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	for dec.More() {
		var record ApiResponse
		if err := dec.Decode(&record); err != nil {
			return err
		}

		Log("json %#v", record)
		if record.Metadata.Result == 1 {
			return nil
		} else {
			Log("metadata: %#v", record.Metadata)
			return fmt.Errorf(record.Metadata.Reason)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return fmt.Errorf("json error? %#v", body)
}

func UnsuspendAccountByEmail(email string) error {
	Log("Unsuspending account: %s", email)

	domain := email[strings.Index(email, "@")+1:]
	info, err := UserDataInfo(domain)
	if err != nil {
		return err
	}

	urlString := apiURI + "unsuspend_outgoing_email?api.version=1&user=" + url.QueryEscape(info.User)
	conn, err := WHMDialer()
	if err != nil {
		return err
	}
	defer conn.Close()
	clientConn := httputil.NewClientConn(conn, nil)

	Log("calling: %s", urlString)
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("whm %s:%s", ApiUser, ApiToken))

	resp, err := clientConn.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	for dec.More() {
		var record ApiResponse
		if err := dec.Decode(&record); err != nil {
			return err
		}

		Log("json %#v", record)
		if record.Metadata.Result == 1 {
			return nil
		} else {
			Log("metadata: %#v", record.Metadata)
			return fmt.Errorf(record.Metadata.Reason)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return fmt.Errorf("json error? %#v", body)
}
