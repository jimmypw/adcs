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
