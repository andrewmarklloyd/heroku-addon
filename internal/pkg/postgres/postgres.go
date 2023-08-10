package postgres

import (
	"database/sql"
	"fmt"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/account"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/crypto"
	_ "github.com/lib/pq"
)

const (
	createTableAccountStmt = `CREATE TABLE IF NOT EXISTS account(
		uuid text PRIMARY KEY,
		email text,
		name text,
		accounttype text,
		accesstoken bytea,
		refreshtoken bytea
		);`
	createTableInstancesStmt = `CREATE TABLE IF NOT EXISTS instance(
		id text PRIMARY KEY,
		accountid text,
		plan text,
		name text,
		CONSTRAINT fk_accountid
			FOREIGN KEY(accountid)
			REFERENCES account(uuid)
		);`
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
		return postgresClient, fmt.Errorf("executing create table account statement: %w", err)
	}

	_, err = db.Exec(createTableInstancesStmt)
	if err != nil {
		return postgresClient, fmt.Errorf("executing create table instances statement: %w", err)
	}

	return postgresClient, nil
}

func (c *Client) CreateOrUpdateAccount(cryptoUtil crypto.Util, account account.Account) error {
	accessEnc, err := cryptoUtil.Encrypt([]byte(account.AccessToken))
	if err != nil {
		return fmt.Errorf("encrypting access token: %w", err)
	}

	refreshEnc, err := cryptoUtil.Encrypt([]byte(account.RefreshToken))
	if err != nil {
		return fmt.Errorf("encrypting refresh token: %w", err)
	}

	// TODO: ensure excluded.* is encrypted
	stmt := "INSERT INTO account(uuid, email, name, accounttype, accesstoken, refreshtoken) VALUES($1, $2, $3, $4, $5, $6) ON CONFLICT (uuid) DO UPDATE SET email = excluded.email, name = excluded.name, accounttype = excluded.accounttype, accesstoken = excluded.accesstoken, refreshtoken = excluded.refreshtoken;"
	_, err = c.sqlDB.Exec(stmt, account.UUID, account.Email, account.Name, account.AccountType, string(accessEnc), string(refreshEnc))
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

func (c *Client) DeleteInstance(accountid, uuid string) error {
	stmt := "DELETE FROM instance WHERE accountid = $1 AND id = $2;"
	_, err := c.sqlDB.Exec(stmt, accountid, uuid)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetAccountFromEmail(cryptoUtil crypto.Util, email, accountType string) (account.Account, error) {
	var accounts []account.Account
	var acct account.Account
	stmt := `SELECT * FROM account WHERE email = $1 AND accounttype = $2 LIMIT 1`
	rows, err := c.sqlDB.Query(stmt, email, accountType)
	if err != nil {
		return acct, fmt.Errorf("executing select query: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a account.Account
		err := rows.Scan(&a.UUID, &a.Email, &a.Name, &a.AccountType, &a.AccessToken, &a.RefreshToken)
		if err != nil {
			return acct, err
		}

		accessToken, err := cryptoUtil.Decrypt([]byte(a.AccessToken))
		if err != nil {
			return acct, err
		}

		refreshToken, err := cryptoUtil.Decrypt([]byte(a.RefreshToken))
		if err != nil {
			return acct, err
		}

		a.AccessToken = string(accessToken)
		a.RefreshToken = string(refreshToken)
		accounts = append(accounts, a)
	}

	if len(accounts) == 0 {
		return account.Account{}, &AccountNotFound{
			Email: email,
		}
	}

	if len(accounts) > 1 {
		return account.Account{}, fmt.Errorf("more than 1 account was returned for email %s", email)
	}

	return accounts[0], nil
}

func (c *Client) CreateOrUpdateInstance(instance account.Instance) error {
	stmt := "INSERT INTO instance(id, accountid, plan, name) VALUES($1, $2, $3, $4);"
	_, err := c.sqlDB.Exec(stmt, instance.Id, instance.AccountID, instance.Plan, instance.Name)
	if err != nil {
		return fmt.Errorf("writing instance: %w", err)
	}

	return nil
}

func (c *Client) GetInstances(accountID string) ([]account.Instance, error) {
	instances := []account.Instance{}
	stmt := `SELECT * FROM instance WHERE accountid = $1;`
	rows, err := c.sqlDB.Query(stmt, accountID)
	if err != nil {
		return instances, fmt.Errorf("executing select query: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		var i account.Instance
		err := rows.Scan(&i.Id, &i.AccountID, &i.Plan, &i.Name)
		if err != nil {
			return instances, err
		}
		instances = append(instances, i)
	}
	return instances, nil
}
