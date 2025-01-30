.PHONY: all clean

all: windows linux darwin cleanup-binaries

windows: check-env
	GOOS=windows GOARCH=amd64 go build -o builds/invoice.exe
	cd builds && zip invoice-cli-${PACKAGE_VERSION}-windows-x64.zip invoice.exe
	GOOS=windows GOARCH=386 go build -o builds/invoice.exe
	cd builds && zip invoice-cli-${PACKAGE_VERSION}-windows-x86.zip invoice.exe

linux: check-env
	GOOS=linux GOARCH=amd64 go build -o builds/invoice
	cd builds && tar czf invoice-cli-${PACKAGE_VERSION}-linux-amd64.tar.gz invoice
	GOOS=linux GOARCH=386 go build -o builds/invoice
	cd builds && tar czf invoice-cli-${PACKAGE_VERSION}-linux-i386.tar.gz invoice
	GOOS=linux GOARCH=arm64 go build -o builds/invoice
	cd builds && tar czf invoice-cli-${PACKAGE_VERSION}-linux-arm64.tar.gz invoice

darwin: check-env
	GOOS=darwin GOARCH=amd64 go build -o builds/invoice
	cd builds && tar czf invoice-cli-${PACKAGE_VERSION}-darwin-amd64.tar.gz invoice
	GOOS=darwin GOARCH=arm64 go build -o builds/invoice
	cd builds && tar czf invoice-cli-${PACKAGE_VERSION}-darwin-arm64.tar.gz invoice

clean:
	rm -rf builds/

cleanup-binaries:
	cd builds && rm -f *.exe
	cd builds && rm -f invoice

check-env:
	@test -n "$(PACKAGE_VERSION)" || { echo "Error: PACKAGE_VERSION is not set"; exit 1; }