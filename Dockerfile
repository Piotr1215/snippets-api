# Lean base image
FROM golang:alpine as builder
# Enable go modules
ENV GO111MODULE=on

# Install git. (alpine image does not have git in it)
RUN apk update && apk add --no-cache git

# Setup workdir in the contianer
WORKDIR /src

# Copy go mod and sum files
COPY go.mod ./
COPY go.sum ./

# Download all dependencies.
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /usr/local/bin/api .

RUN chmod +x /usr/local/bin/api

# Finally our multi-stage to build a small image
# Start a new stage from scratch
FROM scratch

# Copy the Pre-built binary file
COPY --from=builder /usr/local/bin/api .

# Signal exposing port 8080
EXPOSE 8080

# Run executable
CMD ["./api"]

