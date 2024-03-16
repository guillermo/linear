package linear

import (
	"log"
	"net/http"
	"os"

	gqlclient "git.sr.ht/~emersion/gqlclient"
)

//go:generate sh -c "go run git.sr.ht/~emersion/gqlclient/cmd/gqlintrospect@latest https://api.linear.app/graphql > linear.graphql"
//go:generate sh -c "go run git.sr.ht/~emersion/gqlclient/cmd/gqlclientgen@latest -s linear.graphql -q queries.graphql -o linear_gen.go -n linear"

type AuthHeader struct{ Header string }

func (h AuthHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", h.Header)
	return http.DefaultTransport.RoundTrip(req)
}

func DefaultClient() *gqlclient.Client {
	key := os.Getenv("LINEAR_KEY")
	if key == "" {
		log.Fatal("LINEAR_KEY must be set")
	}
	return gqlclient.New(
		"https://api.linear.app/graphql",
		&http.Client{Transport: AuthHeader{Header: key}},
	)
}
