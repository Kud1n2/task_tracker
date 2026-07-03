package email

import "context"

type Client struct{}

func New() *Client {
	return &Client{}
}

func (c *Client) SendInvite(ctx context.Context, userID, teamID int64) error {
	return nil
}
