ARG USER=nobody
ARG LISTEN_PORT=8080
ARG JWT_HEADER=

# build image
FROM golang:1.21-alpine3.18 as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /src

COPY . .

RUN go mod download && \
    go build --ldflags "-s -w" -o http-echo ./cmd/http-echo/


# final image
FROM scratch

ENV LISTEN_PORT=$LISTEN_PORT
ENV JWT_HEADER=$JWT_HEADER

COPY --from=builder /src/http-echo /http-echo

EXPOSE $PORT

USER $USER

ENTRYPOINT [ "/http-echo" ]
