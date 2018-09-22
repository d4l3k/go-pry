FROM golang:alpine as base

RUN apk add binutils git subversion mercurial
RUN strip /usr/local/go/bin/go
RUN rm /usr/local/go/bin/gofmt
RUN rm -r /usr/local/go/src/cmd
#RUN rm -r /usr/local/go/src/vendor
RUN rm -r /usr/local/go/api
RUN rm -r /usr/local/go/lib
RUN rm -r /usr/local/go/misc
RUN rm -r /usr/local/go/test
RUN rm -r /usr/share
RUN find /usr/local/go/src -name "testdata" -exec rm -r {} +
RUN find /usr/lib/python2.7 -name "*.pyo" -exec rm -r {} +
RUN find /usr/lib/python2.7 -name "*.pyc" -exec rm -r {} +
RUN rm -r /usr/local/go/pkg/linux_amd64
RUN strip /usr/local/go/pkg/tool/linux_amd64/*

WORKDIR /go/src/app
COPY . .

RUN CGO_ENABLED=0 go get ./server
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o playground ./server
RUN rm -r /go/src/github.com /go/src/golang.org /go/bin
RUN rm -r /root

EXPOSE 8080
CMD ["/go/src/app/playground"]

#WORKDIR /go/src/app
#COPY . .
#
#FROM scratch
#ENV LD_LIBRARY_PATH /lib/:/usr/lib/
#
#COPY --from=base /go/src/app/ /
#
#COPY --from=base /usr/local/go/bin/go /bin/
#COPY --from=base /usr/local/go/src /usr/local/go/
#
#COPY --from=base /lib/*.so* /lib/
#COPY --from=base /usr/lib/*.so* /usr/lib/
#
#COPY --from=base /usr/bin/git /bin/
#COPY --from=base /usr/bin/svn /bin/
#COPY --from=base /usr/bin/hg /bin/
#
#COPY --from=base /tmp/ /tmp/

#EXPOSE 8080
#CMD ["/playground"]
