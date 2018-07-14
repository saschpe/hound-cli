# Simple Makefile for hound-cli

all:
	go get github.com/soundhound/houndify-sdk-go
	go build hound.go

clean:
	rm ./hound
