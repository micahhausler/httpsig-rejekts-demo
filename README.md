# httpsig-rejekts-demo

A demonstration of HTTP Signature authentication (RFC-9421) using GitHub public keys to create signatures.
This project showcases how to implement secure, signature-based authentication with GitHub as a key disccovery mechanism.

## Installation

### Client

To install the client locally:

```bash
go install github.com/micahhausler/httpsig-rejekts-demo@latest
```

## Usage

The client can be used to make a signed `curl` command to authenticate to the server

```bash
# Set your GitHub username and key path
export GH_USERNAME=your-username
export GH_KEY=~/.ssh/id_ecdsa

# Make a request
httpsig-rejekts-demo --key $GH_KEY --username $GH_USERNAME
```

Will emit something like 
```
curl -X 'POST' \
    'https://rejekts.dev.micahhausler.com/hello' 
    -H 'Content-Type: application/json' \
    -H 'Signature-Input: sig1=("@method" "@target-uri" "content-type" "x-github-username");keyid="7829c799f966275fa9a01ae111e6dd249522611c8df502fcaed17dca039cf1aeeeb2e3bc95e23f4f3326195a14a55aeadbd75f761c501dbb6cb5a3874756ff88";alg="ecdsa-p256-sha256";tag="foo";nonce="Zfy0dMDu62h-6EUKSKUGrXLGihlAp-O-2GpsrCSzaWU=";created=1743327105' \
    -H 'Signature: sig1=:kmG5zcLdDKrUUJRWBqxmF2323eh2K8n9U8yP6pWTblxy4xRcj8zrlPyklO5C/IcQp1EiGPSjYjhmeRi0eCrBZg==:' \
    -H 'X-Github-Username: micahhausler'
```

Once you execute the `curl` command or add `--execute` to the client, the server will then lookup the signing keys (`https://github.com/<username>.keys`) and validate the request is signed with one of that user's keys.

```json
{"message": "hello, micahhausler!"}
```

## Development

### Building

```bash
# Build both server and client
make build

# Build only server
make bin/server

# Build only client
make bin/client
```

### Running Locally

To run locally, you'll need to use a local port and scheme
```bash
make server 

# Run the client
make client CLIENT_ARGS= -host localhost -port=8080 -scheme=http
```

### Server

The server is deployed as a Kubernetes application. See the `deploy/` for an example deployment behind an ingress.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details
