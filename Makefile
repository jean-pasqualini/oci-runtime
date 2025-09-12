.PHONY: build test lint run docker-run docker-build
export GOOS = linux
export LEVEL = medium

docker-build:
	docker build -f Dockerfile.build -t oci-container:build .

# CAP_SYS_ADMIN to avoid the privileged but almost same
# --security-opt seccomp=$(PWD)/docker-default-patched.json
# --security-opt apparmor=unconfined
# --security-opt seccomp=unconfined
# --cap-add CAP_SYS_ADMIN
# --security-opt seccomp=$(PWD)/docker-default-patched.json
docker-run-unconfined:
	docker run --privileged --security-opt seccomp=unconfined --rm -w /app -v $(PWD):/app -it oci-container:build bash
docker-run-least:
	docker run --cap-add CAP_SYS_ADMIN --cap-add CAP_NET_ADMIN	--security-opt seccomp=$(PWD)/docker-default-patched.json --rm -w /app -v $(PWD):/app -it oci-container:build bash
build:
	go build ./...
test:
	go test ./...
run-shim:
	go build -tags $(LEVEL) -o /tmp/oci-runtime -- ./cmd/oci-runtime
	go run ./cmd/container-shim
run:
	go run -tags $(LEVEL) ./cmd/oci-runtime run --root /tmp/state --bundle /app/bundle cid
lint:
	golangci-lint run  ./...
lint-diff:
	golangci-lint run --new-from-rev HEAD~