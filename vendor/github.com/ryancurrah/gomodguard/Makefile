current_dir = $(shell pwd)
version = $(shell printf '%s' $$(cat VERSION))

.PHONEY: lint
lint:
	golangci-lint run -v --enable-all --disable funlen,gochecknoglobals,lll ./...

.PHONEY: build
build:
	go build -o gomodguard cmd/gomodguard/main.go

.PHONEY: dockerbuild
dockerbuild:
	docker build --build-arg GOMODGUARD_VERSION=${version} --tag ryancurrah/gomodguard:${version} .
 
.PHONEY: run
run: build
	./gomodguard

.PHONEY: dockerrun
dockerrun: dockerbuild
	docker run -v "${current_dir}/.gomodguard.yaml:/.gomodguard.yaml" ryancurrah/gomodguard:latest

.PHONEY: release
release:
	git tag ${version}
	git push --tags
	goreleaser --skip-validate --rm-dist

.PHONEY: clean
clean:
	rm -rf dist/
	rm -f gomodguard

.PHONEY: install-tools-mac
install-tools-mac:
	brew install goreleaser/tap/goreleaser
