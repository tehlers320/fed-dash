FROM golang:alpine as builder
RUN apk add --no-cache git 
WORKDIR /build
COPY go.mod /build
RUN go mod download

FROM builder as compiler
COPY *.go /build/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

FROM alpine
WORKDIR /app
COPY --from=compiler /build/main /app
CMD ["/app/main"]
