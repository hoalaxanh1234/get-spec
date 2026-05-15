APP = agent
GOFLAGS = -ldflags="-s -w"

.PHONY: all win64 linux clean run install-tools

all: win64 linux

install-tools:
	go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

win64:
	go generate
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o $(APP).exe .

linux:
	go build $(GOFLAGS) -o $(APP) .

run:
	./$(APP)

clean:
	rm -f $(APP) $(APP).exe resource.syso
