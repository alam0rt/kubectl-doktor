BINARY_NAME=kubectl-doktor
 
all: test build
 
build:
	go build -o ${BINARY_NAME} cmd/kubectl-doktor.go
 
run:
	go build -o ${BINARY_NAME} cmd/kubectl-doktor.go
	./${BINARY_NAME}
 
clean:
	go clean
	rm ${BINARY_NAME}

test:
	go test -v ./...