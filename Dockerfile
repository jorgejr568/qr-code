FROM golang:1.22-alpine as builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .
CMD ["go", "run", "main.go"]
#
#FROM alpine:latest
#WORKDIR /app
#COPY --from=builder /build/main .
#
#EXPOSE 8080
#CMD ["./main"]
