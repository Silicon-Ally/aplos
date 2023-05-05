// Package aplos provides basic support for the Aplos API, see https://www.aplos.com/api
// This package is very much still under development.
package aplos

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/oauth2"
)

// LoadPrivateKeyFromFile loads a base64-encoded, PKCS8-formatted RSA key file
// from disk. This is the format returned from the Aplos UI when creating and
// downloading an API key.
func LoadPrivateKeyFromFile(fp string) (*rsa.PrivateKey, error) {
	// One could use os.Open + base64.NewDecoder to stream the file, but for a key
	// file, which is a fixed size, there's no harm in just loading the whole thing
	// into memory straight away.
	b64EncDat, err := os.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	dat, err := base64.StdEncoding.DecodeString(string(b64EncDat))
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode: %w", err)
	}

	return LoadPrivateKey(dat)
}

// LoadPrivateKey parses PKCS8-formatted bytes into an RSA key.
func LoadPrivateKey(dat []byte) (*rsa.PrivateKey, error) {
	key, err := x509.ParsePKCS8PrivateKey(dat)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key as PKCS8: %w", err)
	}

	k, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key was not an RSA key, was %T", key)
	}

	return k, nil
}

// Client is an authenticated API client for connecting to Aplos.
type Client struct {
	http *http.Client
}

// Transaction represents a single transaction recorded in a register.
type Transaction struct {
	ID             int
	Memo           string
	Date           Date
	IDNumber       int `json:"id_number"`
	Created        Time
	Amount         float64
	InClosedPeriod bool `json:"in_closed_period"`

	// Lines is only populated in the "get single transaction details" endpoint, e.g. GET /.../v1/transactions/{transactionID}
	Lines []TransactionLine
}

// TransactionLine is a single line in a larger transaction, like a journal entry.
type TransactionLine struct {
	ID      int
	Amount  float64
	Account Account
	Fund    Fund
}

type Account struct {
	AccountNumber int `json:"account_number"`
	Name          string

	// Populated in ListAccounts
	Category     string
	AccountGroup *AccountGroup `json:"account_group"`
	IsEnabled    bool          `json:"is_enabled"`
	Type         string
	Activity     string
}

type AccountGroup struct {
	ID   int
	Name string
	Seq  int
}

type Fund struct {
	ID   int
	Name string
}

type getTransactionResponse struct {
	Version string
	Status  int
	Data    getTransactionResponseData
}

type getTransactionResponseData struct {
	Transaction Transaction
}

func (c *Client) Transaction(ctx context.Context, id int) (*Transaction, error) {
	resp, err := ctxhttp.Get(ctx, c.http, "https://www.aplos.com/hermes/api/v1/transactions/"+strconv.Itoa(id))
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer resp.Body.Close()

	var gResp getTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return nil, fmt.Errorf("failed to decode get transaction response: %w", err)
	}

	return &gResp.Data.Transaction, nil
}

type listAccountsResponse struct {
	Version string
	Status  int
	Data    listAccountsResponseData
}

type listAccountsResponseData struct {
	Accounts []Account
}

type listAccountsOpts struct {
	accountName *string
}

func WithAccountName(acctName string) ListAccountOption {
	return func(o *listAccountsOpts) {
		o.accountName = &acctName
	}
}

type ListAccountOption func(*listAccountsOpts)

// Accounts returns a list of accounts satisfying the given options.
func (c *Client) Accounts(ctx context.Context, opts ...ListAccountOption) ([]Account, error) {
	o := &listAccountsOpts{}
	for _, opt := range opts {
		opt(o)
	}

	q := url.Values{}
	if o.accountName != nil {
		q.Add("f_name", *o.accountName)
	}

	resp, err := ctxhttp.Get(ctx, c.http, "https://www.aplos.com/hermes/api/v1/accounts?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer resp.Body.Close()

	var lResp listAccountsResponse
	if err := json.NewDecoder(resp.Body).Decode(&lResp); err != nil {
		return nil, fmt.Errorf("failed to decode list accounts response: %w", err)
	}

	return lResp.Data.Accounts, nil
}

type listTransactionsResponse struct {
	Version string
	Status  int
	Data    listTransactionsResponseData
}

type listTransactionsResponseData struct {
	Transactions []Transaction
}

type listTransactionsOpts struct {
	accountNumber *int
	rangeStart    *Date
	rangeEnd      *Date
}

func WithAccountNumber(acctNumber int) ListTransactionOption {
	return func(o *listTransactionsOpts) {
		o.accountNumber = &acctNumber
	}
}

func WithRangeStart(year int, month time.Month, day int) ListTransactionOption {
	return func(o *listTransactionsOpts) {
		o.rangeStart = &Date{Year: year, Month: month, Day: day}
	}
}

func WithRangeEnd(year int, month time.Month, day int) ListTransactionOption {
	return func(o *listTransactionsOpts) {
		o.rangeEnd = &Date{Year: year, Month: month, Day: day}
	}
}

type ListTransactionOption func(*listTransactionsOpts)

// Transactions returns a list of transactions satisfying the given options.
func (c *Client) Transactions(ctx context.Context, opts ...ListTransactionOption) ([]Transaction, error) {
	o := &listTransactionsOpts{}
	for _, opt := range opts {
		opt(o)
	}

	q := url.Values{}
	if o.accountNumber != nil {
		q.Add("f_accountnumber", strconv.Itoa(*o.accountNumber))
	}
	if o.rangeStart != nil {
		q.Add("f_rangestart", o.rangeStart.String())
	}
	if o.rangeEnd != nil {
		q.Add("f_rangeend", o.rangeEnd.String())
	}

	resp, err := ctxhttp.Get(ctx, c.http, "https://www.aplos.com/hermes/api/v1/transactions?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer resp.Body.Close()

	var lResp listTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&lResp); err != nil {
		return nil, fmt.Errorf("failed to decode list transactions response: %w", err)
	}

	return lResp.Data.Transactions, nil
}

// New returns an Aplos API client initialized with the given key credentials.
// If the credentials are invalid (expired, mismatched, malformed, etc), this
// call with fail.
func New(clientID string, pk *rsa.PrivateKey) (*Client, error) {
	ts, err := newTokenSource(clientID, pk)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	return &Client{
		http: oauth2.NewClient(context.Background(), ts),
	}, nil
}

func newTokenSource(clientID string, key *rsa.PrivateKey) (oauth2.TokenSource, error) {
	t := &ts{key: key, clientID: clientID}
	tkn, err := t.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	return oauth2.ReuseTokenSource(tkn, t), nil
}

type ts struct {
	clientID string
	key      *rsa.PrivateKey
}

type authResponse struct {
	Version string
	Status  int
	Data    authResponseData
}

type authResponseData struct {
	Expires Time
	Token   string
}

// Token performs the Aplos authentication handshake of downloaded the
// encrypted access token for our Client ID and decrypting it with our private
// key credentials. For more details, see the Aplos API Authentication docs:
// https://www.aplos.com/api/authentication
func (t *ts) Token() (*oauth2.Token, error) {
	resp, err := http.Get("https://www.aplos.com/hermes/api/v1/auth/" + t.clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to query auth endpoint: %w", err)
	}
	defer resp.Body.Close()

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	encToken, err := base64.StdEncoding.DecodeString(authResp.Data.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode encrypted token: %w", err)
	}

	dec, err := rsa.DecryptPKCS1v15(nil, t.key, encToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	return &oauth2.Token{
		AccessToken: string(dec),
		TokenType:   "Bearer",
		Expiry:      authResp.Data.Expires.Time,
	}, nil
}
