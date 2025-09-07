# Build
FROM golang:1.25 as build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/oci-runtime ./cmd/oci-runtime

# Runtime
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/oci-runtime /app/oci-runtime
ENTRYPOINT ["/app/oci-runtime"]