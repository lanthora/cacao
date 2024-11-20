# Cacao

Candy Server with WebUI

## Build from source
```bash
# frontend
cd cacao/frontend && npm install && npm run build

# server
cd cacao && go mod tidy && go build

# Optional for ssl expired
cd cacao/ssl && openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout api.key -out api.crt
```

## Run

#### default
```bash
# loglevel=[info] listen=[:80],https=[false],apikey=[.],apicrt=[.] storage=[.]
cacao

# loglevel=[debug] listen=[127.0.0.1:8080],https=[false],apikey=[.] storage=[/var/lib/cacao]
cacao --loglevel=debug --listen=127.0.0.1:8080 --storage=/var/lib/cacao
#
```
#### TLS
```bash
# loglevel=[debug] listen=[127.0.0.1:8080],https=[true],apikey=[./ssl/api.key],apicrt=[./ssl/api.crt] storage=[/var/lib/cacao]
cacao --loglevel=debug --listen=127.0.0.1:8080 --apikey=./ssl/api.key --apicrt=./ssl/api.crt --storage=/var/lib/cacao
```
