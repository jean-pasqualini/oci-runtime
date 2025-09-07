.PHONY: build test lint run docker-run docker-build

docker-build:
	docker build -f Dockerfile.build -t oci-container:build .
docker-run: docker-build
# CAP_SYS_ADMIN to avoid the privileged but almost same
# --security-opt seccomp=$(PWD)/docker-default-patched.json
# --security-opt apparmor=unconfined
# --security-opt seccomp=unconfined
# --cap-add CAP_SYS_ADMIN
	docker run --cap-add CAP_SYS_ADMIN --security-opt seccomp=$(PWD)/docker-default-patched.json --rm -w /app -v $(PWD):/app -it oci-container:build bash
build:
	go build ./...
test:
	go test ./...
run:
	go run ./cmd/oci-runtime
lint:
	GOOS=linux golangci-lint run  ./...
lint-diff:
	GOOS=linux golangci-lint run --new-from-rev HEAD~