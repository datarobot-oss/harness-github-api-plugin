FROM golang:1.22.5-alpine AS build

ADD . .

RUN go mod init harness-github-api-plugin
RUN go get .
RUN go build -v -o .

FROM scratch
COPY --from=build /go/harness-github-api-plugin /go/harness-github-api-plugin
COPY --from=build /etc/ssl /etc/ssl
COPY --from=build /bin/cp /bin/cp
COPY --from=build /lib/ld-musl-x86_64.so.1 /lib/ld-musl-x86_64.so.1
ENTRYPOINT ["/go/harness-github-api-plugin"]
