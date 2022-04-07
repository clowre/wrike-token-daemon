package wrikehttp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	wrikedaemon "github.com/clowre/wrike-token-daemon"
	"github.com/gorilla/mux"
)

func createHandler(daemon *wrikedaemon.Daemon) http.Handler {

	router := mux.NewRouter()
	router.Methods("GET").Path("/get").HandlerFunc(getTokenHandler(daemon))
	router.Methods("GET").Path("/set-code").HandlerFunc(setAuthCodeHandler(daemon))

	return router
}

func getTokenHandler(daemon *wrikedaemon.Daemon) http.HandlerFunc {

	var bo backoff.BackOff
	{
		bo = backoff.NewConstantBackOff(3 * time.Second)
		bo = backoff.WithMaxRetries(bo, 3)
	}

	return func(rw http.ResponseWriter, r *http.Request) {

		var token *wrikedaemon.Token
		err := backoff.Retry(func() error {
			tok, err := daemon.Get()
			if err != nil {
				return err
			}

			token = tok
			return nil
		}, bo)

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}

		rw.Header().Set("content-type", "application/json")
		json.NewEncoder(rw).Encode(token)
	}
}

func setAuthCodeHandler(daemon *wrikedaemon.Daemon) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		c := r.FormValue("code")
		if c == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("code is required"))
			return
		}

		daemon.SetAuthCode(c)
	}
}
