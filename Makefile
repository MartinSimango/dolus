build: 
	@go build ./... 

build-optimized:
	@go build  -ldflags="-s -w" -o $(GOPATH)/bin/dolus-optimized cmd/dolus/main.go

install: build
	@go install ./...

run: install
	@dolus

run-optimized: build-optimized
	@dolus-optimized

debug: $(GOPATH)/bin/dlv
	@dlv debug cmd/dolus/main.go

size:
	@du -h $(GOPATH)/bin/dolus

size-optimized:
	@du -h $(GOPATH)/bin/dolus-optimized

### TOOLS ###
$(GOPATH)/bin/dlv:
	go install github.com/go-delve/delve/cmd/dlv      



