all: env
	go build -o demo *.go
env:
	go get "github.com/gin-gonic/gin"
	go get "github.com/pili-engineering/pili-sdk-go/pili"
	go get "github.com/pili-engineering/pili-sdk-go.v2/pili"

build-linux: env
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o linux-demo *.go
