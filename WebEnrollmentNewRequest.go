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

// WebEnrollmentNewRequest handles the current request. It is not expected to be called directly but through the Submit method.
type WebEnrollmentNewRequest struct {
	webenrollmentserver *WebEnrollmentServer
	csr                 []byte
	template            string
}

type certificateRequestParameters struct {
	mode             string
	targetStoreFlags int
	saveCert         string
	certRequest      string
	certAttrib       string
	friendlyType     string
	thumbPrint       string
}

func (reqParams *certificateRequestParameters) querystring() string {
	reqParamsReflection := reflect.Indirect(reflect.ValueOf(reqParams))

	var parameters []string
	for i := 0; i < reqParamsReflection.NumField(); i++ {
		thisparameter := fmt.Sprintf("%s=%s", reqParamsReflection.Type().Field(i).Name, reqParamsReflection.Field(i).Interface())
		parameters = append(parameters, thisparameter)
	}

	return strings.Join(parameters, "&")
}

// Submit is called from the WebEnrollmentServ implements the WebEnrollmentRequest interface
func (request *WebEnrollmentNewRequest) Submit() (response WebEnrollmentResponse, err error) {
	response.webenrollmentrequest = request
	response.httpresponse, err = request.postHTTPRequest()
	if err != nil {
		return
	}

	var responsebody bytes.Buffer
	responsebody.ReadFrom(response.httpresponse.Body)
	response.status = scrapeNewRequestStatus(responsebody.Bytes())

	switch response.status {
	case SUCCESS:
		response.requestid = scrapeRequestNumber(responsebody.String())
		response.certificatedata, err = request.GetServer().GetCertificate(response.requestid)
		if err != nil {
			return
		}
	case PENDING:
		response.parseSessionCookie()
		response.requestid = scrapePendingRequestNumber(responsebody.String())
	case UNAUTHORIZED:
		err = errors.New("access is denied due to invalid credentials")
		return
	case DENIED:
		err = errors.New("request was denied")
		return
	case FAIL:
		err = errors.New("unknown error has occurred")
		return
	default:
		err = fmt.Errorf("the request failed and I don't know why\nresponse.status =  %d\nResponse body:%s", response.status, responsebody.String())
		return
	}

	return
}

// GetServer retrieves a pointer to the current WebEnrollment server to satisfy the interface
func (request *WebEnrollmentNewRequest) GetServer() *WebEnrollmentServer {
	return request.webenrollmentserver
}

func (request WebEnrollmentNewRequest) postHTTPRequest() (*http.Response, error) {

	client := NewClient()

	postbody := request.certificateRequestBody()
	req, _ := http.NewRequest("POST", request.webenrollmentserver.newCertificateRequestURL(), postbody)
	req.SetBasicAuth(request.webenrollmentserver.Username, request.webenrollmentserver.Password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (wer WebEnrollmentNewRequest) stringifyCertificateRequest() string {
	// CERT=$(cat foo.csr | tr -d '\n\r' | sed 's/+/%2B/g' | tr -s ' ' '+')
	returndata := string(wer.csr)
	re1 := regexp.MustCompile(`\r`)
	re2 := regexp.MustCompile(`\n`)
	re3 := regexp.MustCompile(`\+`)
	re4 := regexp.MustCompile(` `)
	returndata = re1.ReplaceAllString(returndata, "")
	returndata = re2.ReplaceAllString(returndata, "")
	returndata = re3.ReplaceAllString(returndata, "%2B")
	returndata = re4.ReplaceAllString(returndata, "+")
	return returndata
}

func (request WebEnrollmentNewRequest) certAttributes() string {
	// This function is wrong. It needs to take in to account user supplied certificate attributes.
	return fmt.Sprintf("CertificateTemplate:%s", request.template)
}

func (request WebEnrollmentNewRequest) certificateRequestBody() io.Reader {
	timestamp := time.Now().Format(time.RFC1123)

	thisReqParams := certificateRequestParameters{
		mode:             "newreq",
		certRequest:      request.stringifyCertificateRequest(),
		certAttrib:       request.certAttributes(),
		friendlyType:     fmt.Sprintf("Saved-Request Certificate (%s)", timestamp),
		thumbPrint:       "",
		targetStoreFlags: 0,
		saveCert:         "yes",
	}

	return strings.NewReader(thisReqParams.querystring())
}

func scrapePendingRequestNumber(response string) int {
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

func scrapeRequestNumber(response string) int {
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

func scrapeNewRequestStatus(resp []byte) (returndata int) {
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

	return
}
