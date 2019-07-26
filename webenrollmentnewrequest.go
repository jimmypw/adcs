package adcs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	ntlmssp "github.com/Azure/go-ntlmssp"
)

// WebEnrollmentNewRequest handles the current request. It is not expected to be called directly but through the Submit method.
type WebEnrollmentNewRequest struct {
	webenrollmentserver *WebEnrollmentServer
	csr                 []byte
	template            string
}

// Submit implements the WebEnrollmentRequest interface
func (wer *WebEnrollmentNewRequest) Submit() (WebEnrollmentResponse, error) {
	response := WebEnrollmentResponse{
		webenrollmentrequest: wer,
	}

	httpresponse, err := wer.postHTTPRequest()
	if err != nil {
		return WebEnrollmentResponse{}, err
	}

	var respbody bytes.Buffer
	respbody.ReadFrom(httpresponse.Body)

	response.status = wer.parseSuccessStatus(respbody.Bytes())

	switch response.status {
	case SUCCESS:
		// parse certificate number
		response.requestid = wer.parseSuccessRequestNumber(respbody.String())
		// retrieve certificate
		response.certificatedata, err = wer.GetServer().GetCertificate(response.requestid)
		if err != nil {
			return WebEnrollmentResponse{}, err
		}
	case PENDING:
		// parse certificate number
		response.requestid = wer.parsePendingRequestNumber(respbody.String())
	case UNAUTHORIZED:
		return WebEnrollmentResponse{}, errors.New("Access is denied due to invalid credentials")
	case FAIL:
		fallthrough
	default:
		// need to try and establish what went wrong here
		panic(fmt.Sprintf("The request failed and I don't know why\nresponse.status =  %d\nResponse body:\n", response.status, respbody.String()))
	}

	return response, nil
}

// GetServer retrueves a pointer to the current WebEnrollment server to satisfy the interface
func (wer *WebEnrollmentNewRequest) GetServer() *WebEnrollmentServer {
	return wer.webenrollmentserver
}

func (wer WebEnrollmentNewRequest) postHTTPRequest() (*http.Response, error) {

	client := &http.Client{
		Transport: ntlmssp.Negotiator{
			RoundTripper: &http.Transport{},
		},
	}

	postbody := wer.certificateRequestBody()
	req, _ := http.NewRequest("POST", wer.webenrollmentserver.newCertificateRequestURL(), postbody)
	req.SetBasicAuth(wer.webenrollmentserver.Username, wer.webenrollmentserver.Password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (wer WebEnrollmentNewRequest) parseSuccessStatus(resp []byte) int {
	var returndata int
	issued := regexp.MustCompile("Certificate Issued")
	pending := regexp.MustCompile("Your certificate request has been received.")
	unauthorized := regexp.MustCompile("Unauthorized: Access is denied due to invalid credentials.")

	if issued.Match(resp) {
		returndata = SUCCESS
	} else if pending.Match(resp) {
		returndata = PENDING
	} else if unauthorized.Match(resp) {
		returndata = UNAUTHORIZED
	} else {
		returndata = FAIL
	}

	return returndata
}

func (wer WebEnrollmentNewRequest) stringifyCertificateRequest() string {
	// CERT=$(cat foo.csr | tr -d '\n\r' | sed 's/+/%2B/g' | tr -s ' ' '+')
	returndata := string(wer.csr)
	re1 := regexp.MustCompile("\r")
	re2 := regexp.MustCompile("\n")
	re3 := regexp.MustCompile("\\+")
	re4 := regexp.MustCompile(" ")
	returndata = re1.ReplaceAllString(returndata, "")
	returndata = re2.ReplaceAllString(returndata, "")
	returndata = re3.ReplaceAllString(returndata, "%2B")
	returndata = re4.ReplaceAllString(returndata, "+")
	return returndata
}

func (wer WebEnrollmentNewRequest) certAttributes() string {
	// This function is wrong. It needs to take in to account user supplied certificate attributes.
	return fmt.Sprintf("CertificateTemplate:%s", wer.template)
}

func (wer WebEnrollmentNewRequest) certificateRequestBody() io.Reader {
	var postbody strings.Builder
	postbody.WriteString(fmt.Sprintf("Mode=newreq"))
	postbody.WriteByte('&')
	postbody.WriteString(fmt.Sprintf("CertRequest=%s", wer.stringifyCertificateRequest()))
	postbody.WriteByte('&')
	postbody.WriteString(fmt.Sprintf("CertAttrib=%s", wer.certAttributes()))
	return strings.NewReader(postbody.String())

}

func (wer WebEnrollmentNewRequest) parsePendingRequestNumber(response string) int {
	re := regexp.MustCompile(`Your Request Id is (\d+).`)

	match := re.FindStringSubmatch(response)
	if len(match) != 2 {
		// no match
		return -1
	}
	returndata, err := strconv.Atoi(match[1])
	if err != nil {
		return -1
	}
	return returndata
}

func (wer WebEnrollmentNewRequest) parseSuccessRequestNumber(response string) int {
	re := regexp.MustCompile(`certnew.cer\?ReqID=(\d+)&amp;Enc=b64`)

	match := re.FindStringSubmatch(response)
	if len(match) != 2 {
		// no match
		return -1
	}
	returndata, err := strconv.Atoi(match[1])
	if err != nil {
		return -1
	}
	return returndata
}
