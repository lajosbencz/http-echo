# user to run process as
ARG USER=nobody

# default env vars for app
ARG LOG_JSON=0
ARG LISTEN_HOST=0.0.0.0
ARG LISTEN_HTTP=8080
ARG LISTEN_HTTPS=8443
ARG CORS_ENABLED=0
ARG JWT_ENABLED=0
ARG JWT_HEADER=Authorization


# build image
FROM golang:1.21-alpine3.18 as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /src/app

COPY . .

RUN go get -tags musl && \
    go test ./... -covermode=atomic -coverpkg=./... && \
    go build -tags musl --ldflags "-s -w" -o http-echo ./cmd/http-echo/


# final image
FROM scratch

ENV LOG_JSON=$LOG_JSON
ENV LISTEN_HOST=$LISTEN_HOST
ENV LISTEN_HTTP=$LISTEN_HTTP
ENV LISTEN_HTTPS=$LISTEN_HTTPS
ENV CORS_ENABLED=$CORS_ENABLED
ENV JWT_ENABLED=$JWT_ENABLED
ENV JWT_HEADER=$JWT_HEADER

COPY --from=builder /src/app/http-echo /http-echo

EXPOSE $PORT

USER $USER

ENTRYPOINT [ "/http-echo" ]
CMD [ "-env" ]
