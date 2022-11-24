package adcs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// WebEnrollmentRequest interface is implemented by the
// various kinds of requests that can be performed
type WebEnrollmentRequest interface {
	Submit() (WebEnrollmentResponse, error)
	GetServer() *WebEnrollmentServer
}

// WebEnrollmentNewRequest handles the current request. It is not expected to be called directly but through the Submit method.
type WebEnrollmentNewRequest struct {
	webenrollmentserver *WebEnrollmentServer
	csr                 []byte
	template            string
}

// WebEnrollmentPendingRequest is used to attempt to retrieve a pending request
type WebEnrollmentPendingRequest struct {
	webenrollmentserver *WebEnrollmentServer
	requestid           int
}

type certificateRequestParameters struct {
	Mode             string
	TargetStoreFlags int
	SaveCert         string
	CertRequest      string
	CertAttrib       string
	FriendlyType     string
	ThumbPrint       string
}

type pendingRequestParameters struct {
	Mode             string
	TargetStoreFlags int
	SaveCert         string
	ReqID            int
}

func (reqParams *certificateRequestParameters) String() string {
	reqParamsReflection := reflect.Indirect(reflect.ValueOf(reqParams))

	var parameters []string
	for i := 0; i < reqParamsReflection.NumField(); i++ {
		parameters = append(parameters, fmt.Sprintf("%s=%s", reqParamsReflection.Type().Field(i).Name, reqParamsReflection.Field(i).Interface()))
	}

	return strings.Join(parameters, "&")
}

func (reqParams *pendingRequestParameters) String() string {
	reqParamsReflection := reflect.Indirect(reflect.ValueOf(reqParams))

	var parameters []string
	for i := 0; i < reqParamsReflection.NumField(); i++ {
		parameters = append(parameters, fmt.Sprintf("%s=%s", reqParamsReflection.Type().Field(i).Name, reqParamsReflection.Field(i).Interface()))
	}

	return strings.Join(parameters, "&")
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
	case DENIED:
		return WebEnrollmentResponse{}, errors.New("request was denied")
	case FAIL:
		fallthrough
	default:
		// need to try and establish what went wrong here
		panic(fmt.Sprintf("The request failed and I don't know why\nresponse.status =  %d\nResponse body:%s\n", response.status, respbody.String()))
	}

	return response, nil
}

// GetServer retrueves a pointer to the current WebEnrollment server to satisfy the interface
func (wer *WebEnrollmentNewRequest) GetServer() *WebEnrollmentServer {
	return wer.webenrollmentserver
}

func (wer WebEnrollmentNewRequest) postHTTPRequest() (*http.Response, error) {

	client := NewClient()

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
	denied := regexp.MustCompile("Your certificate request was denied.")

	if issued.Match(resp) {
		returndata = SUCCESS
	} else if pending.Match(resp) {
		returndata = PENDING
	} else if unauthorized.Match(resp) {
		returndata = UNAUTHORIZED
	} else if denied.Match(resp) {
		returndata = DENIED
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
	timestamp := time.Now().Format(time.RFC1123)

	thisReqParams := certificateRequestParameters{
		Mode:             "newreq",
		CertRequest:      wer.stringifyCertificateRequest(),
		CertAttrib:       wer.certAttributes(),
		FriendlyType:     fmt.Sprintf("Saved-Request Certificate (%s)", timestamp),
		ThumbPrint:       "",
		TargetStoreFlags: 0,
		SaveCert:         "yes",
	}

	return strings.NewReader(thisReqParams.String())
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
		return WebEnrollmentResponse{}, errors.New("Fail: Unknown error has occurred")
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

	client := NewClient()

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
