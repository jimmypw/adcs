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
	Mode             string
	TargetStoreFlags int
	SaveCert         string
	ReqID            int
}

func (reqParams *pendingRequestParameters) String() string {
	reqParamsReflection := reflect.Indirect(reflect.ValueOf(reqParams))

	var parameters []string
	for i := 0; i < reqParamsReflection.NumField(); i++ {
		parameters = append(parameters, fmt.Sprintf("%s=%s", reqParamsReflection.Type().Field(i).Name, reqParamsReflection.Field(i).Interface()))
	}

	return strings.Join(parameters, "&")
}

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
