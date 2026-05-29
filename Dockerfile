FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o yanzi ./cmd/yanzi

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/yanzi .
EXPOSE 8080
CMD ["./yanzi", "serve", "--host", "0.0.0.0"]
