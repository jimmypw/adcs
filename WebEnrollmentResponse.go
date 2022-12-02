package adcs

import (
	"net/http"
	"regexp"
)

const (
	// SUCCESS status
	SUCCESS = 0
	// PENDING status
	PENDING = 1
	// UNAUTHORIZED status
	UNAUTHORIZED = 2
	// FAIL status
	FAIL = 3
	// DENIED status
	DENIED = 4
)

// WebEnrollmentResponse struct contains the parsed response from adcs
type WebEnrollmentResponse struct {
	webenrollmentrequest WebEnrollmentRequest
	certificatedata      []byte
	status               int
	requestid            int
	httpresponse         *http.Response
	cookiename           string
	cookieval            string
}

// GetCertData returns a byte array of the signed certificate
func (response WebEnrollmentResponse) GetCertData() []byte {
	return response.certificatedata
}

// GetStatus returns a const reflecting the status of the signing request
func (response WebEnrollmentResponse) GetStatus() int {
	return response.status
}

// GetRequestID will return the request ID
func (response WebEnrollmentResponse) GetRequestID() int {
	return response.requestid
}

// GetRequestURL will reututn the url that the request was issued against.
func (response WebEnrollmentResponse) GetRequestURL() string {
	return response.webenrollmentrequest.GetServer().URL
}

func (response *WebEnrollmentResponse) parseSessionCookie() {
	cookies := response.httpresponse.Cookies()

	cookiematcher := regexp.MustCompile("ASPSESSIONID")

	for i := 0; i < len(cookies); i++ {
		if cookiematcher.MatchString(cookies[i].Name) {
			response.cookiename = cookies[i].Name
			response.cookieval = cookies[i].Value
			break
		}
	}
}
