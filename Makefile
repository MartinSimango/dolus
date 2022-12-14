build: 
	@go build ./...

install: build
	@go install ./...

run: install
	@dolus

debug: $(GOPATH)/bin/dlv
	@dlv debug cmd/dolus/main.go

### TOOLS ###
$(GOPATH)/bin/dlv:
	go install github.com/go-delve/delve/cmd/dlv      



