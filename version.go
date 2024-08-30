package adcs

import "fmt"

// VersionMaj is the major version
const VersionMaj = 1

// VersionMin is the minor version
const VersionMin = 2

// VersionPat is the patch version
const VersionPat = 1

// VersionSuffix is the suffix displayed after the version string.
const VersionSuffix = ""

// ShowVersion Generate the version string
func BuildVersionString() string {
	return fmt.Sprintf("v%d.%d.%d%s", VersionMaj, VersionMin, VersionPat, VersionSuffix)
}

// ShowSignature Generate the application signature
func ShowSignature() {
	fmt.Printf("adcscli %s https://github.com/jimmypw/adcs\n", BuildVersionString())
}
