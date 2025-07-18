.PHONY: all clean

all: windows linux macos 

windows: check-env
	GOOS=windows GOARCH=amd64 go build -o builds/invoice.exe
	cd builds && zip invoice-${PACKAGE_VERSION}-windows-x64.zip invoice.exe
	GOOS=windows GOARCH=386 go build -o builds/invoice.exe
	cd builds && zip invoice-${PACKAGE_VERSION}-windows-x86.zip invoice.exe
	GOOS=windows GOARCH=arm64 go build -o builds/invoice.exe
	cd builds && zip invoice-${PACKAGE_VERSION}-windows-arm64.zip invoice.exe
	cd builds && rm -f *.exe

linux: check-env
	GOOS=linux GOARCH=amd64 go build -o builds/invoice
	cd builds && tar czf invoice-${PACKAGE_VERSION}-linux-amd64.tar.gz invoice
	GOOS=linux GOARCH=386 go build -o builds/invoice
	cd builds && tar czf invoice-${PACKAGE_VERSION}-linux-i386.tar.gz invoice
	GOOS=linux GOARCH=arm64 go build -o builds/invoice
	cd builds && tar czf invoice-${PACKAGE_VERSION}-linux-arm64.tar.gz invoice
	cd builds && rm -f invoice

macos: check-env
	GOOS=darwin GOARCH=amd64 go build -o builds/invoice
	cd builds && tar czf invoice-${PACKAGE_VERSION}-macos-amd64.tar.gz invoice
	GOOS=darwin GOARCH=arm64 go build -o builds/invoice
	cd builds && tar czf invoice-${PACKAGE_VERSION}-macos-arm64.tar.gz invoice
	cd builds && rm -f invoice

clean:
	rm -rf builds/

check-env:
	@test -n "$(PACKAGE_VERSION)" || { echo "Error: PACKAGE_VERSION is not set"; exit 1; }