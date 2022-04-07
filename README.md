# Wrike Token Daemon

From the [Wrike FAQ](https://developers.wrike.com/faq/) - if you're developing a UI-less/background application, the
conventional browser-driven way of doing OAuth2 becomes a tedious thing to handle during development.
The `wrike-token-daemon` takes an authorization code on startup, and periodically generates access tokens.

## Running

### Docker

```shell
git clone https://github.com/clowre/wrike-http-daemon
docker build -t wriked .
docker run -d -p 80:8080 --name wriked wriked -I <your wrike client ID> -S <your wrike client secret> http -P 8080
```

Set the redirect URI of your app to `http://localhost/set-code`, and then log into the app by
visiting `https://login.wrike.com/oauth2/authorize/v4?client_id=<your client ID>&response_type=code`.

The endpoint `http://localhost/get` should then return the tokens sent by Wrike!

### No Docker

```shell
go install github.com/clowre/wrike-http-daemon/cmd/wriked@latest
wriked --help
```

### Programmatically

```go
package main

import (
	"log"

	"github.com/clowre/wrike-token-daemon"
)

func main() {

	daemon := wrikedaemon.NewDaemon("client_id", "client_secret")
	go daemon.StartPolling()

	daemon.SetAuthCode("access code returned by wrike")

	tok, err := daemon.Get()
	if err != nil {
		log.Printf("token resolution error: %v", err)
		return
	}

	log.Printf("got access token: %s", tok.AccessToken)
}
```