pkg:pkglinuxamd64 pkglinux386 pkgwindowsamd64 pkgwindows386 pkgdarwinarm64
# build
build:
	CGO_ENABLED=0 go build -ldflags "-X 'main.goVersion=$(shell go version)' -X 'main.gitHash=$(shell git show -s --format=%H)' -X 'main.buildTime=$(shell git show -s --format=%cd)'"

# darwin arm64 build and package
pkgdarwinarm64:builddarwinarm64
	tar -czf ./bin/darwin-arm64/clonefile_darwin-arm64.tar.gz ./bin/darwin-arm64/clonefile
builddarwinarm64:
	-rm ./bin/darwin-arm64/clonefile
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-X 'main.goVersion=$(shell go version)' -X 'main.gitHash=$(shell git show -s --format=%H)' -X 'main.buildTime=$(shell git show -s --format=%cd)'" -o ./bin/darwin-arm64/clonefile
upxdarwinarm64:
	-upx -9 ./bin/darwin-arm64/clonefile

# linux amd64 build and package
pkglinuxamd64:buildlinuxamd64 upxlinuxamd64
	tar -czf ./bin/linux-amd64/clonefile_linux-amd64.tar.gz ./bin/linux-amd64/clonefile
buildlinuxamd64:
	-rm ./bin/linux-amd64/clonefile
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X 'main.goVersion=$(shell go version)' -X 'main.gitHash=$(shell git show -s --format=%H)' -X 'main.buildTime=$(shell git show -s --format=%cd)'" -o ./bin/linux-amd64/clonefile
upxlinuxamd64:
	-upx -9 ./bin/linux-amd64/clonefile

# linux 386 build and package
pkglinux386:buildlinux386 upxlinux386
	tar -czf ./bin/linux-386/clonefile_linux-386.tar.gz ./bin/linux-386/clonefile
buildlinux386:
	-rm ./bin/linux-386/clonefile
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-X 'main.goVersion=$(shell go version)' -X 'main.gitHash=$(shell git show -s --format=%H)' -X 'main.buildTime=$(shell git show -s --format=%cd)'" -o ./bin/linux-386/clonefile
upxlinux386:
	-upx -9 ./bin/linux-386/clonefile

# windwos amd64 build and package
pkgwindowsamd64:buildwindowsamd64 upxwindowsamd64
	tar -czf ./bin/windows-amd64/clonefile_windows-amd64.tar.gz ./bin/windows-amd64/clonefile.exe
buildwindowsamd64:
	-rm ./bin/windows-amd64/clonefile.exe
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X 'main.goVersion=$(shell go version)' -X 'main.gitHash=$(shell git show -s --format=%H)' -X 'main.buildTime=$(shell git show -s --format=%cd)'" -o ./bin/windows-amd64/clonefile.exe
upxwindowsamd64:
	-upx -9 ./bin/windows-amd64/clonefile.exe

# windwos 386 build and package
pkgwindows386:buildwindows386 upxwindows386
	tar -czf ./bin/windows-386/clonefile_windows-386.tar.gz ./bin/windows-386/clonefile.exe
buildwindows386:
	-rm ./bin/windows-386/clonefile.exe
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-X 'main.goVersion=$(shell go version)' -X 'main.gitHash=$(shell git show -s --format=%H)' -X 'main.buildTime=$(shell git show -s --format=%cd)'" -o ./bin/windows-386/clonefile.exe
upxwindows386:
	-upx -9 ./bin/windows-386/clonefile.exe