package jsonapi

//go:generate easyjson -all -snake_case $GOPATH/src/github.com/gsmcwhirter/eso-discord/pkg/discordapi/jsonapi/gateway_response.go

// GatewayResponse TODOC
//easyjson:json
type GatewayResponse struct {
	URL    string
	Shards int
}
