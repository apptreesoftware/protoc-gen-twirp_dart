BINARY := protoc-gen-twirp_dart

TIMESTAMP := $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

LDFLAGS := -ldflags "-X main.Timestamp=${TIMESTAMP} -X main.Commit=${COMMIT} -X main.Branch=${BRANCH}"

all: clean test install

install:
	go install ${LDFLAGS} github.com/matthewtsmith/protoc-gen-twirp_dart

test:
	go test -v ./...

lint:
	go list ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

run: install
	mkdir -p example/ts_client && \
	protoc --proto_path=${GOPATH}/src:. --twirp_out=. --go_out=. --twirp_dart_out=package_name=haberdasher:./example/ts_client ./example/service.proto

build_linux:
	GOOS=linux GOARCH=amd64 go build -o ${BINARY} ${LDFLAGS} github.com/matthewtsmith/protoc-gen-twirp_dart

clean:
	-rm -f ${GOPATH}/bin/${BINARY}
