package adcs

import "fmt"

// VersionMaj is the major version
var VersionMaj = 1

// VersionMin is the minor version
var VersionMin = 1

// VersionPat is the patch version
var VersionPat = 1

// ShowVersion Generate the version string
func ShowVersion() {
	fmt.Printf("v%d.%d.%d\n", VersionMaj, VersionMin, VersionPat)
}

// ShowSignature Generate the application signature
func ShowSignature() {
	fmt.Printf("adcscli version %d.%d.%d https://github.com/jimmypw/adcs\n", VersionMaj, VersionMin, VersionPat)
}
