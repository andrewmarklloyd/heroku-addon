package account

import (
	"context"
	"fmt"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/heroku"
)

func CreateAccount(ctx context.Context, tokenResp heroku.TokenResponse, resourceUUID string) error {
	fmt.Println("UUID:", resourceUUID)
	fmt.Println("tokens:", tokenResp.AccessToken, tokenResp.RefreshToken)
	return nil
}
