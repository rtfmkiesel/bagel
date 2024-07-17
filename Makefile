NAME = "bagel"

.PHONY: make
make: bin

.PHONY: bin
bin:
	go mod tidy
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ${NAME}

.PHONY: docker
docker:
	docker build -t ${NAME} .

.PHONY: docker-run
docker-run: docker
	docker run --rm -p 8080:8080 ${NAME}