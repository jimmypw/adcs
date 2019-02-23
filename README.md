# ADCS

A library and command line interface for Active Directory Certificate Services web enrollment service.

While attempting to automate public key infrastrucutre in an enterprise environment it quickly became apparent that the common approaches such as ACME and SCEP are inadequate when dealing with devices that do not support such prototols. An example of this is a Cisco network device I was configuring. All this device is able to do is generate CSR's. SCEP/NDES - specifically the renewal part requires access to the private key in order to renew the certificate, ACME would would require reconfiguration of DNS every time the certificate was to be renewed.

So we are left with either interfacing with ADCS directly through the MMC or the Web Enrolment serivice. When you have a thousand devices to deploy, doing stuff manually doesn't scores very high apathy scale. Here is where adcscli is born. The command line interface is the linux equivilent to what the certutil.exe command is on windows. However this command will submit CSR's to the Web Enrollment service and scrape the response. If the command was successful the signed certificate will be presented back to the user. If the request requires manual intervention then a command will be generated to retrieve the response.

## Installation

go install github.com/jimmypw/adcs/cli/...

I may if i remember put a binary on the github releases page. Check there also.

## Usage

There are two modes of operation, the first is submitting a csr. The second is checking the status of the csr.

### To submit a new csr:

adcscli -csr csr.csr -out crt.crt -password 'supersecurepassword' -username auser -url http://192.168.252.140/certsrv/ -template webtemplate

The CSR will be submitted to the web enrollment service and will produce one of two responses.

- The certificate request requires admin approval

In this case a command will be returned to the user to check the status of the request.

for example `adcscli -pend -url http://192.168.252.140/certsrv/ -cookiename ASPSESSIONIDAR -cookieval NBANOMME -requestid 3395`

The command will emit the exit status of 1 to indicate a pending request.

Notice that this command does not return the username and password. These are required to submit the pending request.

- The certificate request was successful

In this case a certificate will be returned to the user. The certficicate can be saved to the filesystem by using the -out command line switch.

### To check the status of an existing request

The responses are the same as submitting a new CSR.