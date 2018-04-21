package client

type DiscordClient interface {
}

type discordClient struct {
}

func NewDiscordClient() DiscordClient {
	return discordClient{}
}
