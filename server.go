package adcs

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	ntlmssp "github.com/Azure/go-ntlmssp"
)

// WebEnrollmentServer structure defines the ADCS web enrollment server
type WebEnrollmentServer struct {
	URL      string
	Username string
	Password string
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
func (wes *WebEnrollmentServer) CheckPendingRequest(requestid int) (WebEnrollmentResponse, error) {
	wer := WebEnrollmentPendingRequest{
		webenrollmentserver: wes,
		requestid:           requestid,
	}

	response, err := wer.Submit()
	if err != nil {
		return WebEnrollmentResponse{}, err
	}
	return response, nil
}

func (wes WebEnrollmentServer) newCertificateRequestURL() string {
	return fmt.Sprintf("%s/certfnsh.asp", wes.URL)
}
func (wes WebEnrollmentServer) newCertificateResponseURL() string {
	return fmt.Sprintf("%s/certnew.cer", wes.URL)
}

// NewClient builds an http client object for ntlm authentication
func NewClient() *http.Client {
	return &http.Client{
		Transport: ntlmssp.Negotiator{
			RoundTripper: &http.Transport{
				TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
			},
		},
	}
}

// getCertificate will retrieve the specified certificate from the server
func (wes *WebEnrollmentServer) getCertificate(requestid string) ([]byte, error) {
	client := NewClient()

	url := fmt.Sprintf("%s?ReqID=%s&Enc=b64", wes.newCertificateResponseURL(), requestid)

	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(wes.Username, wes.Password)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("unable to request certificate")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	return buf.Bytes(), nil
}

// GetCertificate will retrieve the specified ID certificate from the server
func (wes *WebEnrollmentServer) GetCertificate(requestid int) ([]byte, error) {
	return wes.getCertificate(strconv.Itoa(requestid))
}

// GetCACertificate will retrieve the CA certificate from the server
func (wes *WebEnrollmentServer) GetCACertificate() ([]byte, error) {
	return wes.getCertificate("CACert")
}
