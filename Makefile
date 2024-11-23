GOBUILD = env CGO_ENABLED=0 go build -trimpath -ldflags '-w -s' -o cacao

default: cacao

all: linux-amd64 linux-arm64 linux-armv7 freebsd-amd64

frontend:
	cd frontend && npm run build || (npm install && npm run build)

cacao: frontend
	$(GOBUILD)

linux-amd64: frontend
	GOOS=linux GOARCH=amd64 $(GOBUILD)-$@

linux-arm64: frontend
	GOOS=linux GOARCH=arm64 $(GOBUILD)-$@

linux-armv7: frontend
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD)-$@

freebsd-amd64: frontend
	GOOS=freebsd GOARCH=amd64 $(GOBUILD)-$@

.PHONY: default all frontend cacao linux-amd64 linux-arm64 linux-armv7 freebsd-amd64
