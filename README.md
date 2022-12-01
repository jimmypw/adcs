# ADCS

A library and command line interface for scraping Active Directory Certificate Services Web enrollment Service web interface.

While attempting to automate public key infrastrucutre in an enterprise environment it quickly became apparent that the common approaches such as ACME and SCEP are inadequate when dealing with devices that do not support such prototols. An example of this is a Cisco network device I was configuring. All this device is able to do is generate CSR's. SCEP/NDES - specifically the renewal part requires access to the private key in order to renew the certificate, ACME, while cool is unsuitable for use within an enterprise network.

So we are left with either interfacing with ADCS directly through the MMC or the Web Enrolment serivice. When you have a thousand devices to deploy, doing stuff manually isn't very appealing. Here is where adcscli is born. adcscli is a scraper for microsofts web enrollment services. Given a CSR and a few other pieces of information the software will interact with Active Directory Certificate Services Web Enrollment Services and hopefully retrieve a signed certificate.

## Here Be Dragons

adcs is a scraper. Because Web Enrollment Services does not have an API it instead relies on regular expressions. The patterns being matched may change and this software may stop working at any time.

## Installation

Head to https://github.com/jimmypw/adcs/releases for binaries for your favourite operating system.

## Usage

There are two modes of operation, the first is submitting a csr. The second is checking the status of the csr.

### To submit a new csr:

`adcscli -new -csr csr.csr -out crt.crt -password 'supersecurepassword' -username auser -url http://192.168.252.140/certsrv/ -template webtemplate`

The CSR will be submitted to the web enrollment service and will produce one of two responses.

#### The certificate request requires admin approval

In this case a command will be returned to the user to check the status of the request.

for example `adcscli -pend -url http://192.168.252.140/certsrv/ -username 'username' -password 'password' -requestid 3395`

The command will emit the exit status of 1 to indicate a pending request.

Notice that this command does not return the username and password. These are required to submit the pending request.

#### The certificate request was successful

In this case a certificate will be returned to the user. The certficicate can be saved to the filesystem by using the -out command line switch. If `-out` is not supplied the certificate will be printed to stdout.

### To check the status of an existing request

The responses are the same as submitting a new CSR.