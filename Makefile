BINARY_NAME=kubectl-doktor
 
all: build
 
build:
	go build -o ${BINARY_NAME} cmd/kubectl-doktor.go
 
run:
	go build -o ${BINARY_NAME} cmd/kubectl-doktor.go
	./${BINARY_NAME}
 
clean:
	go clean
	rm ${BINARY_NAME}