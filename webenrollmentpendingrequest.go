package adcs

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	ntlmssp "github.com/Azure/go-ntlmssp"
)

// WebEnrollmentPendingRequest is used to attempt to retrieve a pending request
type WebEnrollmentPendingRequest struct {
	webenrollmentserver *WebEnrollmentServer
	requestid           int
}

// Submit implements the WebEnrollmentRequest interface
func (wepr *WebEnrollmentPendingRequest) Submit() (WebEnrollmentResponse, error) {
	response := WebEnrollmentResponse{
		webenrollmentrequest: wepr,
		requestid:            wepr.requestid,
		status:               PENDING,
	}

	httpresponse, err := wepr.postHTTPRequest()
	if err != nil {
		return WebEnrollmentResponse{}, err
	}

	var respbody bytes.Buffer
	respbody.ReadFrom(httpresponse.Body)

	response.status = wepr.parseSuccessStatus(respbody.Bytes())

	switch response.status {
	case SUCCESS:
		response.requestid = wepr.requestid
		// retrieve certificate
		response.certificatedata, err = wepr.GetServer().GetCertificate(response.requestid)
		if err != nil {
			return WebEnrollmentResponse{}, err
		}
	case PENDING:
		response.requestid = wepr.requestid
	case UNAUTHORIZED:
		return WebEnrollmentResponse{}, errors.New("Unauthorized: Access is denied")
	case FAIL:
		fallthrough
	default:
		// need to try and establish what went wrong here
		panic("The request failed and i do not know why")
	}

	return response, nil
}

// GetServer retrueves a pointer to the current WebEnrollment server to satisfy the interface
func (wepr *WebEnrollmentPendingRequest) GetServer() *WebEnrollmentServer {
	return wepr.webenrollmentserver
}

func (wepr WebEnrollmentPendingRequest) postHTTPRequest() (*http.Response, error) {

	client := &http.Client{
		Transport: ntlmssp.Negotiator{
			RoundTripper: &http.Transport{},
		},
	}

	postbody := wepr.pendingRequestBody()
	req, _ := http.NewRequest("POST", wepr.webenrollmentserver.newCertificateRequestURL(), postbody)
	req.SetBasicAuth(wepr.webenrollmentserver.Username, wepr.webenrollmentserver.Password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (wepr WebEnrollmentPendingRequest) pendingRequestBody() io.Reader {
	// 		Mode	chkpnd
	// 		ReqID	339
	// 		SaveCert	yes
	// 		TargetStoreFlags	0

	thisReqParams := pendingRequestParameters{
		Mode:             "chkpnd",
		TargetStoreFlags: 0,
		SaveCert:         "yes",
		ReqID:            wepr.requestid,
	}

	return strings.NewReader(thisReqParams.String())
}

func (wepr WebEnrollmentPendingRequest) parseSuccessStatus(resp []byte) int {
	var returndata int
	issued := regexp.MustCompile("The certificate you requested was issued to you")
	pending := regexp.MustCompile("still pending")
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

func (wepr WebEnrollmentPendingRequest) parseSuccessRequestNumber(response string) int {
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

// // CheckPendingRequest checks the status of a pending request
// func (wes *WebEnrollmentServer) CheckPendingRequest(cookiename, cookieval, requestid string) (WebEnrollmentResponse, error) {
// 	/*
// 		curl 'http://192.168.252.140'
// 		-H 'Host: 192.168.252.140'
// 		-H 'User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0'
// 		-H 'Accept: text/html,application/xhtml+xml,application/xml'
// 		-H 'Accept-Language: en-GB,en;q=0.5'
// 		--compressed
// 		-H 'Referer: http://192.168.252.140/certsrv/certckpn.asp'
// 		-H 'Content-Type: application/x-www-form-urlencoded'
// 		-H 'Cookie: Requests=%5B339%2C0%2Cyes%2CSaved%2DRequest+Certificate+%2827%2F02%2F2019++09%3A58%3A30%29%5D; ASPSESSIONIDACCABSRQ=MPAFGPCDOMACAOEKENNDCDKP'
// 		-H 'Connection: keep-alive'
// 		-H 'Upgrade-Insecure-Requests: 1'
// 		--data ''

// 		Mode	chkpnd
// 		ReqID	339
// 		SaveCert	yes
// 		TargetStoreFlags	0
// 	*/
// 	return response
// }
