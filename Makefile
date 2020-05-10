.PHONY: generate
generate:
	docker build -t n0stack/storage/protoc proto
	docker run -it --rm \
		-u $(UID):$(GID) \
		-v /etc/passwd:/etc/passwd:ro \
		-v /etc/group:/etc/group:ro \
		-v $(PWD):/src/n0stack/storage \
		-v $(PWD):/go/src/github.com/n0stack/storage \
		n0stack/storage/protoc
	go generate -v ./...

.PHONY: build
build:
	go build -o bin/api-generator api/generator/*
