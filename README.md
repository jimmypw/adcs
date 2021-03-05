# ADCS

A library and command line interface for Active Directory Certificate Services web enrollment service.

Cloud native services such as letsencrypt don't work very well in the enterprise. Many of us use Microsoft Active Directory Certificate Services. However the automation for this product is non existant. There is the Web Enrollment Services that does a good job at allowing engineers to be able request certificates through a web page hoever there is no API. So I created this utility to scrape web enrolment services web application to facilitate automation that requires certificates in the enterprise.

## Installation

Check the (releases)[https://github.com/jimmypw/adcs/releases] page. I've built binaries for many operating systems / archictures. 

## Usage

There are two modes of operation, the first is submitting a csr. The second is checking the status of the csr.

### To submit a new csr:

`adcscli -csr csr.csr -out crt.crt -password 'supersecurepassword' -username 'username' -url http://192.168.252.140/certsrv/ -template 'web server'`

The CSR will be submitted to the web enrollment service and will produce one of two responses.

#### The certificate request was successful

In this case a certificate will be returned to the user. The certficicate can be saved to the filesystem by using the `-out` command line switch.

#### The certificate request requires admin approval

In this case a command will be returned to the user to check the status of the request.

`adcscli -pend -url http://192.168.252.140/certsrv/ -username 'username' -password 'password' -requestid 3395`

The command will emit the exit status of 1 to indicate a pending request.