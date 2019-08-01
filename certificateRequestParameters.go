package adcs

import (
	"fmt"
	"reflect"
	"strings"
)

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

func (certReqParams *certificateRequestParameters) String() string {
	certReqParamsReflection := reflect.Indirect(reflect.ValueOf(certReqParams))

	var parameters []string
	for i := 0; i < certReqParamsReflection.NumField(); i++ {
		parameters = append(parameters, fmt.Sprintf("%s=%s", certReqParamsReflection.Type().Field(i).Name, certReqParamsReflection.Field(i).Interface()))
	}

	return strings.Join(parameters, "&")
}

func (pendReqParams *pendingRequestParameters) String() string {
	pendReqParamsReflection := reflect.Indirect(reflect.ValueOf(pendReqParams))

	var parameters []string
	for i := 0; i < pendReqParamsReflection.NumField(); i++ {
		parameters = append(parameters, fmt.Sprintf("%s=%s", pendReqParamsReflection.Type().Field(i).Name, pendReqParamsReflection.Field(i).Interface()))
	}

	return strings.Join(parameters, "&")
}
