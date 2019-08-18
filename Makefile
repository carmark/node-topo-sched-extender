BIN_DIR=_output/cmd/bin

all: init build

build: init
	GOOS=linux GOARCH=amd64 go build -o ${BIN_DIR}/node-topology-sched ./cmd/node-topology-sched

verify:
	hack/verify-gofmt.sh

init:
	mkdir -p ${BIN_DIR}
clean:
	rm -fr ${BIN_DIR}

.PHONY: clean
