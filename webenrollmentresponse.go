package adcs

const (
	// SUCCESS status
	SUCCESS = 0
	// PENDING status
	PENDING = 1
	// FAIL status
	FAIL = 3
)

// WebEnrollmentResponse struct contains the parsed response from adcs
type WebEnrollmentResponse struct {
	webenrollmentrequest WebEnrollmentRequest
	certificatedata      []byte
	status               int
	requestid            int
}

// GetCertData returns a string of the signed certificate
func (wer WebEnrollmentResponse) GetCertData() []byte {
	return wer.certificatedata
}

// GetStatus returns a const reflecting the status of the signing request
func (wer WebEnrollmentResponse) GetStatus() int {
	return wer.status
}

// GetRequestID will return the request ID
func (wer WebEnrollmentResponse) GetRequestID() int {
	return wer.requestid
}

// GetRequestCookieName will return the name of the cookie associated with the request
func (wer WebEnrollmentResponse) GetRequestCookieName() string {
	return wer.webenrollmentrequest.GetServer().cookie.Name
}

// GetRequestCookieVal will return the value of the request cookie
func (wer WebEnrollmentResponse) GetRequestCookieVal() string {
	return wer.webenrollmentrequest.GetServer().cookie.Value
}

// GetRequestURL will reututn the url that the request was issued against.
func (wer WebEnrollmentResponse) GetRequestURL() string {
	return wer.webenrollmentrequest.GetServer().URL
}
