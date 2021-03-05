package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jimmypw/adcs"
)

type options struct {
	csr       *string
	url       *string
	username  *string
	password  *string
	template  *string
	out       *string
	requestid *int
	pend      *bool
	version   *bool
}

func parseSwitches() options {
	var opt options
	opt.csr = flag.String("csr", "", "The path to the certificate signing request")
	opt.url = flag.String("url", "", "The url to the web enrollment server http://webenroll/certsrv/")
	opt.username = flag.String("username", "", "The username to authenticate with")
	opt.password = flag.String("password", "", "The password to authenticate with")
	opt.template = flag.String("template", "", "The short name of the template you wish to use")
	opt.out = flag.String("out", "", "Where to save the certificate.")
	opt.requestid = flag.Int("requestid", 0, "The value of the cookie.")
	opt.pend = flag.Bool("pend", false, "Attempt to retrieve a pending request")
	opt.version = flag.Bool("v", false, "Show Version Information.")
	flag.Parse()
	return opt
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found

}

func processSuccessfulRequest(opt options, wer adcs.WebEnrollmentResponse) {
	if isFlagSet("out") {
		ioutil.WriteFile(*opt.out, wer.GetCertData(), 0644)
	} else {
		fmt.Printf("%s", wer.GetCertData())
	}
}

func processPendingRequest(opt options, wer adcs.WebEnrollmentResponse) {
	fmt.Printf("%s -pend -url %s -requestid %d\n", os.Args[0], wer.GetRequestURL(), wer.GetRequestID())
}

func main() {
	var response adcs.WebEnrollmentResponse
	opt := parseSwitches()
	wes := adcs.WebEnrollmentServer{
		URL:      *opt.url,
		Username: *opt.username,
		Password: *opt.password,
	}

	if *opt.version {
		adcs.ShowVersion()
		os.Exit(1)
	}

	if *opt.pend {
		// attempt to retrieve a pending request
		var err error

		response, err = wes.CheckPendingRequest(*opt.requestid)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	} else {
		// Is new request

		csr, err := ioutil.ReadFile(*opt.csr)

		if err != nil {
			fmt.Printf("Error: Unable to open certificate request %s\n", *opt.csr)
			os.Exit(2)
		}

		response, err = wes.SubmitNewRequest(csr, *opt.template)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

	}

	switch response.GetStatus() {
	case adcs.SUCCESS:
		processSuccessfulRequest(opt, response)
	case adcs.PENDING:
		processPendingRequest(opt, response)
	default:
	}

	os.Exit(response.GetStatus())
}
