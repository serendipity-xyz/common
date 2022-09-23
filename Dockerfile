FROM golang:1.18

WORKDIR /go/src/serendipity-xyz/core/

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["go test -v ./..."]