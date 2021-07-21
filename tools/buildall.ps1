$archs=("amd64", "i386")
$oss=("linux", "windows")

mkdir build

foreach ($a in $archs) {
    foreach ($o in $oss) {
        $outname="out/adcscli-$o-$a"
        if ($o -eq "windows") {
            $outname+=".exe"
        }
        go build -o $outname .\cli\adcscli
    }
}