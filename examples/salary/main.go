// Command salary provides an example of using the Aplos API to:
// 1. Look up a list of accounts with the name 'Salaries'
// 2. Load transactions that touch the 'Salaries' account
// 3. Sum up the relevant line items.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Silicon-Ally/aplos"
)

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("args cannot be empty")
	}

	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		clientID = fs.String("client_id", "", "Required. The Aplos Client ID of the API key to use for authentication, should be formatted like a UUID")
		keyPath  = fs.String("key_path", "", "Required. The RSA private key used to authenticate with the Aplos API.")
	)
	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %v", err)
	}

	if *clientID == "" {
		return errors.New("no --client_id was specified, but is required for authenticating")
	}

	if *keyPath == "" {
		return errors.New("no --key_path was specified, but is required for authenticating")
	}

	key, err := aplos.LoadPrivateKeyFromFile(*keyPath)
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	c, err := aplos.New(*clientID, key)
	if err != nil {
		return fmt.Errorf("failed to init Aplos client: %w", err)
	}

	ctx := context.Background()

	accts, err := c.Accounts(ctx, aplos.WithAccountName("Salaries"))
	if err != nil {
		return fmt.Errorf("failed to load 'Salaries' accounts: %w", err)
	}
	if len(accts) != 1 {
		return fmt.Errorf("found %d accounts with 'Salaries' in the name, expected exactly one", len(accts))
	}
	salaryAccountNumber := accts[0].AccountNumber

	txns, err := c.Transactions(ctx, aplos.WithAccountNumber(salaryAccountNumber))
	if err != nil {
		return fmt.Errorf("failed to load transactions: %w", err)
	}

	// We use a time.Ticker as a basic rate limiter so that we aren't hammering the
	// Aplos API.
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()

	var total float64
	for _, t := range txns {
		txn, err := c.Transaction(ctx, t.ID)
		if err != nil {
			return fmt.Errorf("failed to load transaction %d: %w", t.ID, err)
		}
		v, ok := sumLinesByAccount(txn.Lines, salaryAccountNumber)
		if !ok {
			return fmt.Errorf("transaction %d had no salary component", t.ID)
		}
		<-tick.C
		fmt.Printf("Salary for transaction %d: %f\n", t.ID, v)
		total += v
	}

	fmt.Println("Total: ", total)

	return nil
}

func sumLinesByAccount(lines []aplos.TransactionLine, acctNumber int) (float64, bool) {
	found := false
	var total float64
	for _, l := range lines {
		if l.Account.AccountNumber == acctNumber {
			total += l.Amount
			found = true
		}
	}
	return total, found
}
