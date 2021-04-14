GOTOOLS = \
	gotest.tools/gotestsum

test: tools
	gotestsum --format=short-verbose $(TEST) $(TESTARGS)

generate:
	cd testdata && make generate

modules:
	go mod download && go mod verify

test-circle:
	mkdir -p test-results/terraform-json
	gotestsum --format=short-verbose --junitfile test-results/terraform-json/results.xml

tools:
	go install $(GOTOOLS)

.PHONY: test generate modules test-circle tools
