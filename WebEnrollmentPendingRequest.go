package adcs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

type pendingRequestParameters struct {
	mode             string
	targetStoreFlags int
	saveCert         string
	reqID            int
}

// String() will convert the request parameters in to a query string
func (reqParams *pendingRequestParameters) querystring() string {
	reqParamsReflection := reflect.Indirect(reflect.ValueOf(reqParams))

	var parameters []string
	for i := 0; i < reqParamsReflection.NumField(); i++ {
		thisparameter := fmt.Sprintf("%s=%s", reqParamsReflection.Type().Field(i).Name, reqParamsReflection.Field(i).Interface())
		parameters = append(parameters, thisparameter)
	}

	return strings.Join(parameters, "&")
}

// WebEnrollmentPendingRequest is used to attempt to retrieve a pending request
type WebEnrollmentPendingRequest struct {
	webenrollmentserver *WebEnrollmentServer
	requestid           int
	cookiename          string
	cookieval           string
}

// Submit implements the WebEnrollmentRequest interface
func (request *WebEnrollmentPendingRequest) Submit() (response WebEnrollmentResponse, err error) {
	response.webenrollmentrequest = request
	response.requestid = request.requestid
	response.status = PENDING

	response.httpresponse, err = request.postHTTPRequest()
	if err != nil {
		return WebEnrollmentResponse{}, err
	}

	var respbody bytes.Buffer
	respbody.ReadFrom(response.httpresponse.Body)

	response.status = scrapePendingRequestStatus(respbody.Bytes())

	switch response.status {
	case SUCCESS:
		response.requestid = request.requestid
		// retrieve certificate
		response.certificatedata, err = request.GetServer().GetCertificate(response.requestid)
		if err != nil {
			return WebEnrollmentResponse{}, err
		}
	case PENDING:
		response.cookiename = request.cookiename
		response.cookieval = request.cookieval
		response.requestid = request.requestid
	case UNAUTHORIZED:
		return WebEnrollmentResponse{}, errors.New("unauthorized: Access is denied")
	case DENIED:
		return WebEnrollmentResponse{}, errors.New("request was denied")
	case FAIL:
		return WebEnrollmentResponse{}, errors.New("unknown error has occurred")
	default:
		return WebEnrollmentResponse{}, fmt.Errorf("the request failed and I don't know why\nresponse.status =  %d\nResponse body:%s", response.status, respbody.String())
	}

	return response, nil
}

// GetServer retrueves a pointer to the current WebEnrollment server to satisfy the interface
func (request *WebEnrollmentPendingRequest) GetServer() *WebEnrollmentServer {
	return request.webenrollmentserver
}

func (request WebEnrollmentPendingRequest) postHTTPRequest() (response *http.Response, err error) {

	client := NewClient()

	postbody := request.pendingRequestBody()
	req, _ := http.NewRequest("POST", request.webenrollmentserver.newCertificateRequestURL(), postbody)
	req.AddCookie(&http.Cookie{
		Name:  request.cookiename,
		Value: request.cookieval,
	})
	req.SetBasicAuth(request.webenrollmentserver.Username, request.webenrollmentserver.Password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err = client.Do(req)
	return
}

// pendingRequestBody will build the POST request body
func (request WebEnrollmentPendingRequest) pendingRequestBody() io.Reader {
	// 		Mode	chkpnd
	// 		ReqID	339
	// 		SaveCert	yes
	// 		TargetStoreFlags	0

	thisReqParams := pendingRequestParameters{
		mode:             "chkpnd",
		targetStoreFlags: 0,
		saveCert:         "yes",
		reqID:            request.requestid,
	}

	return strings.NewReader(thisReqParams.querystring())
}

func scrapePendingRequestStatus(resp []byte) int {
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
