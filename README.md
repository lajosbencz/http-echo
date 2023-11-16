# http-echo
Yet Another HTTP echo image for dummy payloads


## Download

- Docker Hub
    - https://hub.docker.com/r/lajosbencz/http-echo
    - `docker pull lajosbencz/http-echo`
- GitHub Registry
    - https://github.com/lajosbencz/http-echo/pkgs/container/http-echo
    - `docker pull ghcr.io/lajosbencz/http-echo`
- GitHub Release
    - https://github.com/lajosbencz/http-echo/releases

## Config


### Arguments

```
-cors
      Allow CORS
-env
      Overwrite options from ENV
-host string
      Host to listen on (default "0.0.0.0")
-jwt
      Enable parsing of JWT
-jwt-header string
      JWT header name (default "Authorization")
-log-json
      Set log format to JSON
-port int
      Port to listen on (default 8080)
```


### Environment

⚠️ Environment variables will only be used if `-env` argument was enabled _(enabled by default for Docker image)_

```bash 
LOG_JSON="0"
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
curl -s -X POST -H "Content-Type: application/json" -H "${JWT_HEADER}: Bearer ${JWT_TOKEN}" http://localhost:${LISTEN_PORT}/foo?baz=bax -d '{"foo":"bar"}'
```


### Response

```json
{
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
