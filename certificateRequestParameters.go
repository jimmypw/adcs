package adcs

import (
	"fmt"
	"reflect"
	"strings"
)

type certificateRequestParameters struct {
	Mode             string
	CertRequest      string
	CertAttrib       string
	FriendlyType     string
	ThumbPrint       string
	TargetStoreFlags int
	SaveCert         string
}

func (certReqParams *certificateRequestParameters) String() string {
	certReqParamsReflection := reflect.Indirect(reflect.ValueOf(certReqParams))

	var parameters []string
	for i := 0; i < certReqParamsReflection.NumField(); i++ {
		parameters = append(parameters, fmt.Sprintf("%s=%s", certReqParamsReflection.Type().Field(i).Name, certReqParamsReflection.Field(i).Interface()))
	}

	return strings.Join(parameters, "&")
}
