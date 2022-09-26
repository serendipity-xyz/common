FROM golang:1.18

WORKDIR /go/src/serendipity-xyz/common

COPY . .

RUN go mod tidy

ENTRYPOINT ["go", "test", "-v", "./..."]