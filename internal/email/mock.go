package email

import (
	"context"
	"errors"
)

type Mock struct {
	Fail bool
}

func (m *Mock) SendInvite(ctx context.Context, userID, teamID int64) error {
	if m.Fail {
		return errors.New("email service unavailable")
	}

	return nil
}
