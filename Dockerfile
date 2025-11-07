FROM golang:1.25 AS builder
ENV CGO_ENABLED 0
ADD . /app
WORKDIR /app
RUN go build -ldflags "-s -w" -v -o dorc .

FROM alpine:3
RUN apk update && \
    apk add openssl && \
    rm -rf /var/cache/apk/* \
    && mkdir /app

WORKDIR /app

ADD Dockerfile /Dockerfile

COPY --from=builder /app/dorc /usr/local/bin/dorc

RUN chown nobody /usr/local/bin/dorc \
    && chmod 500 /usr/local/bin/dorc

USER nobody

ENTRYPOINT ["/usr/local/bin/dorc"]