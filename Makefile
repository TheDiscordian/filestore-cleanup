VERSION = $(shell git tag --contains)

default:
	go fmt
	go build -ldflags "-X main.version=$(VERSION)"

rel:
	go fmt
	mkdir rel/

	CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=$(VERSION)" -o filestore-cleanup
	upx filestore-cleanup
	tar -caf filestore-cleanup-linux64.tar.xz filestore-cleanup LICENSE README.md
	mv filestore-cleanup-linux64.tar.xz rel/

	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags "-X main.version=$(VERSION)" -o filestore-cleanup
	upx filestore-cleanup
	tar -caf filestore-cleanup-linuxARM.tar.xz filestore-cleanup LICENSE README.md
	mv filestore-cleanup-linuxARM.tar.xz rel/

	CGO_ENABLED=0 GOOS=darwin go build -ldflags "-X main.version=$(VERSION)" -o filestore-cleanup
	upx filestore-cleanup
	tar -caf filestore-cleanup-darwin64.tar.gz filestore-cleanup LICENSE README.md
	mv filestore-cleanup-darwin64.tar.gz rel/

	CGO_ENABLED=0 GOOS=windows go build -ldflags "-X main.version=$(VERSION)" -o filestore-cleanup.exe
	upx filestore-cleanup.exe
	zip filestore-cleanup-win64.zip filestore-cleanup.exe LICENSE README.md
	mv filestore-cleanup-win64.zip rel/
