# Simple Makefile for hound-cli

all:
	go get github.com/casimir/xdg-go
	go get github.com/soundhound/houndify-sdk-go
	go get github.com/spf13/viper
	go build hound.go

clean:
	rm ./hound
