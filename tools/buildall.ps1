$archs=("amd64", "386")
$oss=("linux", "windows")

foreach ($a in $archs) {
    foreach ($o in $oss) {
        $outname="out/adcscli-$o-$a"
        $env:GOOS=$o
        $env:GOARCH=$a
        if ($o -eq "windows") {
            $outname+=".exe"
        }
        go build -o $outname .\cli\adcscli
    }
}