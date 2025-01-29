.PHONY: all clean

all: windows linux darwin

windows:
	GOOS=windows GOARCH=amd64 go build -o builds/windows/amd64/invoice.exe
	GOOS=windows GOARCH=386 go build -o builds/windows/386/invoice.exe
	GOOS=windows GOARCH=arm64 go build -o builds/windows/arm64/invoice.exe
	GOOS=windows GOARCH=arm go build -o builds/windows/arm/invoice.exe

linux:
	GOOS=linux GOARCH=amd64 go build -o builds/linux/amd64/invoice
	GOOS=linux GOARCH=386 go build -o builds/linux/386/invoice
	GOOS=linux GOARCH=arm64 go build -o builds/linux/arm64/invoice

darwin:
	GOOS=darwin GOARCH=amd64 go build -o builds/darwin/amd64/invoice
	GOOS=darwin GOARCH=arm64 go build -o builds/darwin/arm64/invoice

clean:
	rm -rf builds/