# http-echo

Yet Another HTTP echo server for dummy payloads, focusing on small image size (`~5MB`)


## Download

#### [Docker Hub](https://hub.docker.com/r/lajosbencz/http-echo)

```bash
docker pull lajosbencz/http-echo
```

#### [GitHub Registry](https://github.com/lajosbencz/http-echo/pkgs/container/http-echo)

```bash
docker pull ghcr.io/lajosbencz/http-echo
```

#### [GitHub Release Binaries](https://github.com/lajosbencz/http-echo/releases)


## Config

### Arguments

```
-cors
      Allow CORS
-env
      Overwrite options from ENV
-host string
      Host to listen on (default "0.0.0.0")
-http int
      HTTP port to listen on (default 8080)
-https int
      HTTPS port to listen on, 0 turns it off (default 8443)
-jwt
      Enable parsing of JWT
-jwt-header string
      JWT header name (default "Authorization")
-log-json
      Set log format to JSON
-log-level int
      Logging level of Zerolog (default 1)
```

### Environment

⚠️ Environment variables will only be used if `-env` argument was enabled _(enabled by default for Docker image)_

```bash 
LOG_JSON="0"
LOG_LEVEL="1"
LISTEN_HOST="0.0.0.0"
LISTEN_PORT="8080"
CORS_ENABLED="0"
JWT_ENABLED="0"
JWT_HEADER="Authorization"
```


## Example

### Request

```bash
JWT_TOKEN="..."
curl -s \
-X POST \
-H "X-Set-Response-Status-Code: 401" \
-H "X-Set-Response-Delay-Ms: 1000" \
-H "Content-Type: application/json" \
-H "${JWT_HEADER}: Bearer ${JWT_TOKEN}" \
-d '{"foo":"bar"}' \
http://localhost:${LISTEN_PORT}/foo?baz=bax
```

### Response

```json
{
  "status_code": 401,
  "hostname": "localhost:8080",
  "path": "/foo",
  "method": "POST",
  "query": {
    "baz": [
      "bax"
    ]
  },
  "headers": {
    "Accept": [
      "*/*"
    ],
    "Authorization": [
      "Bearer ..."
    ],
    "Content-Length": [
      "13"
    ],
    "Content-Type": [
      "application/json"
    ],
    "User-Agent": [
      "curl/7.68.0"
    ]
  },
  "json": {
    "foo": "bar"
  },
  "jwt": {
    "header": {
      "alg": "HS256"
    },
    "payload": {
      "name": "Joe Coder"
    },
    "signature": "..."
  }
}
```
