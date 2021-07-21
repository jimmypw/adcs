package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jimmypw/adcs"
)

type options struct {
	pend   *bool
	newreq *bool

	csr   *string
	stdin *bool

	out    *string
	stdout *bool

	url       *string
	username  *string
	password  *string
	template  *string
	requestid *int

	version *bool
}

func parseSwitches() options {
	var opt options

	// Commands
	opt.pend = flag.Bool("pend", false, "Attempt to retrieve a pending request")
	opt.newreq = flag.Bool("new", false, "Submit a new request")

	// Options
	opt.csr = flag.String("csr", "", "The path to the certificate signing request")
	opt.stdin = flag.Bool("stdin", false, "Provides a CSR through STDIN")

	opt.out = flag.String("out", "", "Where to save the certificate, if not specified, defaults to STDOUT")

	opt.url = flag.String("url", "", "The url to the web enrollment server http://webenroll/certsrv/")
	opt.username = flag.String("username", "", "The username to authenticate with")
	opt.password = flag.String("password", "", "The password to authenticate with")
	opt.template = flag.String("template", "", "The short name of the template you wish to use")
	opt.requestid = flag.Int("requestid", 0, "The value of the cookie.")
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

func missingOption(name string) {
	os.Stderr.WriteString(fmt.Sprintf("Error: Option -%s is required.\n", name))
	os.Exit(1)
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

func processResponse(opt options, response adcs.WebEnrollmentResponse) {
	switch response.GetStatus() {
	case adcs.SUCCESS:
		processSuccessfulRequest(opt, response)
	case adcs.PENDING:
		processPendingRequest(opt, response)
	default:
		os.Stderr.WriteString("Request Failed.\n")
	}
}

func main() {
	opt := parseSwitches()

	if *opt.version {
		adcs.ShowSignature()
		os.Exit(1)
	}

	if *opt.pend && *opt.newreq {
		os.Stderr.WriteString("You must only use one of -new and -pend.\n")
		os.Exit(1)
	}

	if *opt.pend {
		// attempt to retrieve a pending request
		// requires the following to be set
		//   url
		//   username
		//   password
		//   requestid
		if !isFlagSet("url") {
			missingOption("url")
		}
		if !isFlagSet("username") {
			missingOption("username")
		}
		if !isFlagSet("password") {
			missingOption("password")
		}
		if !isFlagSet("requestid") {
			missingOption("requestid")
		}

		wes := adcs.WebEnrollmentServer{
			URL:      *opt.url,
			Username: *opt.username,
			Password: *opt.password,
		}

		response, err := wes.CheckPendingRequest(*opt.requestid)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		processResponse(opt, response)
	} else if *opt.newreq {
		// attempt to submit a new request
		// requires the following to be set
		//   url
		//   username
		//   password
		//   one of: csr / stdin
		if !isFlagSet("url") {
			missingOption("url")
		}
		if !isFlagSet("username") {
			missingOption("username")
		}
		if !isFlagSet("password") {
			missingOption("password")
		}
		if !isFlagSet("template") {
			missingOption("template")
		}

		if !(isFlagSet("csr") || isFlagSet("stdin")) {
			os.Stderr.WriteString("You must specify either -csr or -stdin.\n")
			os.Exit(1)
		}

		if isFlagSet("csr") && isFlagSet("stdin") {
			os.Stderr.WriteString("You must specify only one of either -csr or -stdin.\n")
			os.Exit(1)
		}

		wes := adcs.WebEnrollmentServer{
			URL:      *opt.url,
			Username: *opt.username,
			Password: *opt.password,
		}

		var csr []byte
		if isFlagSet("csr") {
			var err error
			csr, err = ioutil.ReadFile(*opt.csr)

			if err != nil {
				fmt.Printf("Error: Unable to open certificate request: %s\n", *opt.csr)
				os.Exit(1)
			}
		} else if isFlagSet("stdin") {
			csrlen, err := os.Stdin.Read(csr)
			if err != nil {
				fmt.Printf("Error: read csr from stdin: %s\n", *opt.csr)
				os.Exit(1)
			}
			if csrlen == 0 {
				os.Stderr.WriteString("Error: no data read from stdin\n")
				os.Exit(1)
			}
		} else {
			os.Stderr.WriteString("No valid CSR source specified.\n")
			os.Exit(1)
		}

		response, err := wes.SubmitNewRequest(csr, *opt.template)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		processResponse(opt, response)

	} else {
		os.Stderr.WriteString("You must specify one of -new or -pend\n")
		os.Exit(1)
	}

}
