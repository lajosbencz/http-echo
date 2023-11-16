# http-echo
Yet Another HTTP echo image for dummy payloads


## Config


### From Terminal

```
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


### From Environment

```bash 
LOG_JSON="0"
LISTEN_HOST="0.0.0.0"
LISTEN_PORT="8080"
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
  "uri": "/foo?baz=bax",
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
