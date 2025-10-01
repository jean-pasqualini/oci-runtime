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
	docker run --privileged --security-opt seccomp=unconfined --rm -w /app -v $(PWD):/app -v go-mod-cache:/go/pkg/mod -v go-build-cache:/root/.cache/go-build -it oci-container:build bash
docker-run-least: docker-build
	docker run --cap-add CAP_SYS_ADMIN --cap-add CAP_NET_ADMIN	--security-opt seccomp=$(PWD)/docker-default-patched.json --security-opt apparmor=everything --rm -w /app -v $(PWD):/app -v go-mod-cache:/go/pkg/mod -v go-build-cache:/root/.cache/go-build -it oci-container:build bash
build:
	go build -buildvcs=false -tags $(LEVEL) -o /tmp/oci-runtime -- ./cmd/oci-runtime
test:
	go test ./...
run-shim:
	go build -tags $(LEVEL) -o /tmp/oci-runtime -- ./cmd/oci-runtime
	go run -tags $(LEVEL) ./cmd/container-shim
delete:
	rm -rf /tmp/state/cid/
create: delete
	go run -tags $(LEVEL) ./cmd/oci-runtime --root /tmp/state create --bundle /app/bundle cid
create-tty: delete
	go run -tags $(LEVEL) ./cmd/oci-runtime --root /tmp/state create --console-socket /tmp/console.socket --bundle /app/bundle cid
start:
	go run -tags $(LEVEL) ./cmd/oci-runtime --root /tmp/state start cid
run: delete
	go run -tags $(LEVEL) ./cmd/oci-runtime --root /tmp/state run --bundle /app/bundle cid
run-strace: build
	# trace go run and follow children, outputs per-pid logs in /tmp
	strace -s 200 -ttt \
	  -e trace=execve,execveat,open,openat,mount,umount2,clone,fork,vfork \
	  /tmp/oci-runtime --root /tmp/state run --bundle /app/bundle cid
run-strace-follow: build
	# trace go run and follow children, outputs per-pid logs in /tmp
	strace -ff -f -o /tmp/strace.log -s 200 -ttt \
	  -e trace=execve,execveat,open,openat,mount,umount2,clone,fork,vfork \
	  /tmp/oci-runtime --root /tmp/state run --bundle /app/bundle cid
lint:
	golangci-lint run  ./...
lint-diff:
	golangci-lint run --new-from-rev HEAD~
install-in-docker-desktop:
	docker run -w /app -v $(PWD):/app -v go-mod-cache:/go/pkg/mod -v go-build-cache:/root/.cache/go-build -v /bin:/host/bin --rm -it golang:1.25 ./install-in-docker-desktop.sh
# temp stuffs
build-tmp:
	go build -gcflags="all=-N -l" -tags $(LEVEL) -o /tmp/oci-runtime -- ./cmd/oci-runtime
debug-tmp: build-tmp
	dlv exec /tmp/oci-runtime -- run --root /tmp/state --bundle /app/bundle cid