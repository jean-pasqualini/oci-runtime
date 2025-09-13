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
docker-run-unconfined: docker-build
	docker run --privileged --security-opt seccomp=unconfined --rm -w /app -v $(PWD):/app -it oci-container:build bash
docker-run-least: docker-build
	docker run --cap-add CAP_SYS_ADMIN --cap-add CAP_NET_ADMIN	--security-opt seccomp=$(PWD)/docker-default-patched.json --rm -w /app -v $(PWD):/app -it oci-container:build bash
build:
	go build ./...
test:
	go test ./...
run-shim:
	go build -tags $(LEVEL) -o /tmp/oci-runtime -- ./cmd/oci-runtime
	go run -tags $(LEVEL) ./cmd/container-shim
create:
	go run -tags $(LEVEL) ./cmd/oci-runtime create --root /tmp/state --bundle /app/bundle cid
start:
	go run -tags $(LEVEL) ./cmd/oci-runtime start --root /tmp/state cid
run:
	go run -tags $(LEVEL) ./cmd/oci-runtime run --root /tmp/state --bundle /app/bundle cid
lint:
	golangci-lint run  ./...
lint-diff:
	golangci-lint run --new-from-rev HEAD~
# temp stuffs
build-tmp:
	go build -gcflags="all=-N -l" -tags $(LEVEL) -o /tmp/oci-runtime -- ./cmd/oci-runtime
debug-tmp: build-tmp
	dlv exec /tmp/oci-runtime -- run --root /tmp/state --bundle /app/bundle cid