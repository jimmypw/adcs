package adcs

import "fmt"

// VersionMaj is the major version
var VersionMaj = 1

// VersionMin is the minor version
var VersionMin = 0

// VersionPat is the patch version
var VersionPat = 0

// ShowVersion Generate the version string
func ShowVersion() {
	fmt.Printf("adcscli version %d.%d.%d\n", VersionMaj, VersionMin, VersionPat)
}
