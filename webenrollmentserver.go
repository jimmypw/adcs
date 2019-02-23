package adcs

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	ntlmssp "github.com/Azure/go-ntlmssp"
)

// WebEnrollmentServer structure defines the ADCS web enrollment server
type WebEnrollmentServer struct {
	URL      string
	Username string
	Password string
	cookie   *http.Cookie
}

// SubmitNewRequest will submit the WebEnrollment reqest and populate the object with the response
func (wes *WebEnrollmentServer) SubmitNewRequest(incsr []byte, template string) (WebEnrollmentResponse, error) {
	wer := WebEnrollmentNewRequest{
		webenrollmentserver: wes,
		csr:                 incsr,
		template:            template,
	}
	// response, err := wer.do()
	response, err := wer.Submit()
	if err != nil {
		return WebEnrollmentResponse{}, err
	}

	return response, nil
}

// CheckPendingRequest will check to see if the request has been completed or not.
func (wes *WebEnrollmentServer) CheckPendingRequest(cookiename, cookieval string, requestid int) (WebEnrollmentResponse, error) {
	wer := WebEnrollmentPendingRequest{
		webenrollmentserver: wes,
		requestid:           requestid,
		cookiename:          cookiename,
		cookieval:           cookieval,
	}

	response, err := wer.Submit()
	if err != nil {
		return WebEnrollmentResponse{}, err
	}
	return response, nil
}

// SetCookie will restore a cookie to a previous state provided by the user
func (wes *WebEnrollmentServer) SetCookie(name string, value string) {
	wes.cookie = new(http.Cookie)
	wes.cookie.Name = name
	wes.cookie.Value = value
}

func (wes WebEnrollmentServer) newCertificateRequestURL() string {
	return fmt.Sprintf("%s/certfnsh.asp", wes.URL)
}
func (wes WebEnrollmentServer) newCertificateResponseURL() string {
	return fmt.Sprintf("%s/certnew.cer", wes.URL)
}

func (wes *WebEnrollmentServer) updateCookie() error {
	client := &http.Client{
		Transport: ntlmssp.Negotiator{
			RoundTripper: &http.Transport{},
		},
	}

	req, _ := http.NewRequest("GET", wes.newCertificateRequestURL(), nil)
	req.SetBasicAuth(wes.Username, wes.Password)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	theCookie := findCookieLike("ASPSESSIONID", resp.Cookies())
	if theCookie == nil {
		return errors.New("Unable to update cookie, invalid username and password?")
	}
	wes.cookie = theCookie
	return nil
}

// GetCertificate will retrieve the specified certificate from the server
func (wes *WebEnrollmentServer) GetCertificate(requestid int) ([]byte, error) {
	client := &http.Client{
		Transport: ntlmssp.Negotiator{
			RoundTripper: &http.Transport{},
		},
	}

	url := fmt.Sprintf("%s?ReqID=%d&Enc=b64", wes.newCertificateResponseURL(), requestid)

	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(wes.Username, wes.Password)
	req.AddCookie(wes.cookie)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("Unable to request certificate")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	return buf.Bytes(), nil
}

func findCookieLike(match string, cookies []*http.Cookie) *http.Cookie {
	re := regexp.MustCompile(match)
	for i := 0; i < len(cookies); i++ {
		if re.MatchString(cookies[i].Name) {
			return cookies[i]
		}
	}
	return nil
}
