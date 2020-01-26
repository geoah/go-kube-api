COVEROUT     := ./coverage.out

# Go env vars
export GO111MODULE=on

# Verbose output
ifdef VERBOSE
V = -v
endif

# Install deps
.PHONY: deps
deps:
	$(info Installing go dependencies)
	@go mod download

# Tidy go modules
.PHONY: tidy
tidy:
	$(info Tidying go modules)
	@go mod tidy

# Run go test
.PHONY: test
test:
	$(info Running tests)
	@go test $(V) \
		-tags="$(TAGS)" \
		-count=1 \
		--race \
		-covermode=atomic \
		-coverprofile=$(COVEROUT) \
		./...

# Lint go code
.PHONY: lint
lint:
	@golint ./...

# Clean up everything
.PHONY: clean
clean:
	rm -f *.cov
	rm -rf ./bin

# Build binary
.PHONY: build
build: clean deps
	$(info Building binary to bin/go-kube-api)
	@CGO_ENABLED=0 go build -o bin/go-kube-api \
		-installsuffix cgo \
		-ldflags '-s -w' \
		./cmd

# Build docker image
.PHONY: docker
docker:
	docker build -t go-kube-api:dev .
