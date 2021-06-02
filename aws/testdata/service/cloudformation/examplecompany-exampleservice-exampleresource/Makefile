.PHONY: build test clean

build:
	make -f makebuild  # this runs build steps required by the cfn cli

test:
	cfn generate
	env GOOS=linux go build -ldflags="-s -w" -o bin/handler cmd/main.go

clean:
	rm -rf bin
