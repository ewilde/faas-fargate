GO_FILES?=$$(find . -name '*.go' |grep -v vendor)
TAG?=latest
SQUASH?=false

default: lint build test testacc

test: goimportscheck
	go test -v . .

testacc: goimportscheck
	go test -count=1 -v . -run="TestAcc" -timeout 20m

local-build: goimportscheck vet
	@go install
	@echo "Build succeeded"

build:
	docker build -t ewilde/faas-fargate:$(TAG) . --squash=${SQUASH}

release:
	go get github.com/goreleaser/goreleaser; \
	goreleaser; \

clean:
	rm -rf pkg/

goimports:
	goimports -w $(GO_FILES)

goimportscheck:
	@sh -c "'$(CURDIR)/scripts/goimportscheck.sh'"

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

lint:
	@echo "golint ."
	@golint -set_exit_status $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Lint found errors in the source code. Please check the reported errors"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

.PHONY: build test testacc vet goimports goimportscheck errcheck test-compile lint
