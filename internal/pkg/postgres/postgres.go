package postgres

import (
	"database/sql"
	"fmt"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/crypto"
	_ "github.com/lib/pq"
)

const (
	createTableAccountStmt = `CREATE TABLE IF NOT EXISTS account(uuid text PRIMARY KEY, owneremail text, accesstoken bytea, refreshtoken bytea);`
)

type Client struct {
	sqlDB *sql.DB
}

func NewPostgresClient(databaseURL string) (Client, error) {
	postgresClient := Client{}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return postgresClient, err
	}
	postgresClient.sqlDB = db

	_, err = db.Exec(createTableAccountStmt)
	if err != nil {
		return postgresClient, err
	}

	return postgresClient, nil
}

func (c *Client) CreateOrUpdateAccount(cryptoUtil crypto.Util, uuid, ownerEmail, accessToken, refreshToken string) error {
	accessEnc, err := cryptoUtil.Encrypt([]byte(accessToken))
	if err != nil {
		return fmt.Errorf("encrypting access token: %w", err)
	}

	refreshEnc, err := cryptoUtil.Encrypt([]byte(refreshToken))
	if err != nil {
		return fmt.Errorf("encrypting refresh token: %w", err)
	}

	// TODO: ensure excluded.* is encrypted
	stmt := "INSERT INTO account(uuid, owneremail, accesstoken, refreshtoken) VALUES($1, $2, $3, $4) ON CONFLICT (uuid) DO UPDATE SET owneremail = excluded.owneremail, accesstoken = excluded.accesstoken, refreshtoken = excluded.refreshtoken;"
	_, err = c.sqlDB.Exec(stmt, uuid, ownerEmail, string(accessEnc), string(refreshEnc))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteAccout(uuid string) error {
	stmt := "DELETE FROM account WHERE uuid = $1;"
	_, err := c.sqlDB.Exec(stmt, uuid)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetAccount(cryptoUtil crypto.Util, uuid string) (Account, error) {
	var accounts []Account
	var account Account
	stmt := `SELECT * FROM account WHERE uuid = $1 LIMIT 1`
	rows, err := c.sqlDB.Query(stmt, uuid)
	if err != nil {
		return account, fmt.Errorf("executing select query: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a Account
		err := rows.Scan(&a.UUID, &a.AccessToken, &a.RefreshToken)
		if err != nil {
			return account, err
		}

		accessToken, err := cryptoUtil.Decrypt([]byte(a.AccessToken))
		if err != nil {
			return account, err
		}

		refreshToken, err := cryptoUtil.Decrypt([]byte(a.RefreshToken))
		if err != nil {
			return account, err
		}

		accounts = append(accounts, Account{
			UUID:         a.UUID,
			AccessToken:  string(accessToken),
			RefreshToken: string(refreshToken),
		})
	}

	if len(accounts) == 0 {
		return Account{}, fmt.Errorf("no account was returned for uuid %s", uuid)
	}

	if len(accounts) > 1 {
		return Account{}, fmt.Errorf("more than 1 account was returned for uuid %s", uuid)
	}

	return accounts[0], nil
}
