.PHONY: build test clean

build: mod test
	CGO_ENABLED=0 sam build

test: mod
	sam validate --lint
	go test ./...

mod:
	go mod tidy

clean:
	rm -rf .aws-sam
