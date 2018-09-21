FROM golang:alpine as base

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o server ./server

FROM scratch
COPY --from=base /go/src/app/ /

EXPOSE 8080
CMD ["/server"]

