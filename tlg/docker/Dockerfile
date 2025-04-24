FROM golang:alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /aidicti.top

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Install required dependencies
RUN apk add --no-cache git curl bash protobuf

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

ENV PATH="$PATH:/root/go/bin"

WORKDIR /aidicti.top/api

RUN protoc --go_out=../.. --go-grpc_out=../.. --plugin=$(which protoc-gen-go-grpc) cmn.proto uis.proto tlg.proto

WORKDIR /aidicti.top
# Build the binary
RUN go build -o tlg_srvc ./tlg/cmd/tlg/main.go



# Use a smaller base image for the final image
FROM alpine:latest

# Install ca-certificates
# RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /aidicti.top/* .

# Command to run the executable
# CMD ["/root/stt_srvc"]
ENTRYPOINT ["/root/tlg_srvc"]