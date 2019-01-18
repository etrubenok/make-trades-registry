FROM 876031820668.dkr.ecr.ap-southeast-2.amazonaws.com/beewave-base-go-build-docker:latest as builder

RUN apk add musl-utils

ENV SERVICE make-trades-registry

RUN mkdir -p /go/src/github.com/etrubenok/$SERVICE
ADD . /go/src/github.com/etrubenok/$SERVICE
WORKDIR /go/src/github.com/etrubenok/$SERVICE
RUN rm -rf /go/src/github.com/etrubenok/$SERVICE/vendor

RUN eval "$(ssh-agent -s)" && \
    ssh-add /etrubenok-beewave-bitbucket && \
    dep ensure -v
RUN go test
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -tags static_all -ldflags '-extldflags "-static"' -o $GOPATH/bin/main .

FROM scratch
COPY --from=builder /lib/ld-musl-x86_64.so.1 /lib/
COPY --from=builder /go/bin/main /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
CMD ["./main", "-logtostderr=true"]