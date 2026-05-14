APP = agent
GOFLAGS = -ldflags="-s -w"

.PHONY: all win64 linux clean run

all: win64 linux

win64:
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o $(APP).exe .

linux:
	go build $(GOFLAGS) -o $(APP) .

run:
	./$(APP)

clean:
	rm -f $(APP) $(APP).exe
