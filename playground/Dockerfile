FROM golang:alpine as base

RUN apk add binutils
RUN strip /usr/local/go/bin/go
RUN rm -r /usr/local/go/src/cmd
RUN rm -r /usr/local/go/src/vendor
RUN find /usr/local/go/src -name "testdata" -exec rm -r {} +

WORKDIR /go/src/app
COPY . .

FROM scratch
COPY --from=base /go/src/app/ /
COPY --from=base /usr/local/go/bin/go /bin/go
COPY --from=base /usr/local/go/src /lib/go/src

EXPOSE 8080
CMD ["/playground"]
