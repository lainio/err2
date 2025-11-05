package main

import (
	"errors"
	"fmt"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func (db *Database) MoneyTransfer(from, to *Account, amount int) (err error) {
	defer err2.Handle(&err)

	tx := try.To1(db.BeginTransaction())
	defer err2.Handle(&err, func(err error) error {
		if errRoll := tx.Rollback(); errRoll != nil {
			// with go 1.20< we can wrap two errors as below:
			// err = fmt.Errorf("%w: ROLLBACK ERROR: %w", err, errRoll)

			// with go 1.18 (err2 minimum need) we cannot wrap two errors
			// same time.
			err = fmt.Errorf("%v\nROLLBACK ERROR: %w", err, errRoll)

			// NOTE: that this is a good sample how difficult error handling
			// can be. Now we select to wrap rollback error and use original
			// as a main error message, no wrapping for it.
		}
		return err
	})

	try.To(from.ReserveBalance(tx, amount))
	try.To(from.Withdraw(tx, amount))
	try.To(to.Deposit(tx, amount))
	try.To(tx.Commit())

	return nil
}

func doDBMain() {
	defer err2.Catch("CATCH Warning: %s", "test-name")

	db, from, to := new(Database), new(Account), new(Account)

	// --- play with these lines to simulate different errors:
	db.errRoll = errRollback
	//db.err = errBegin              // tx fails
	from.balance = 1100            // no enough funds
	from.errWithdraw = errWithdraw // withdraw error
	to.errDeposit = errDeposit     // deposit error
	amount := 100                  // no enough funds
	// --- simulation variables end

	try.To(db.MoneyTransfer(from, to, amount))

	fmt.Println("all ok")
}

type (
	Database struct {
		err     error
		errRoll error
	}

	Tx struct {
		errBegint   error
		errCommit   error
		errRollback error
		db          *Database
	}

	Account struct {
		errWithdraw error
		errDeposit  error
		balance     int
		reserved    int
	}
)

var (
	errWithdraw = errors.New("AML error, FBI freeze")
	errDeposit  = errors.New("AML error")
	errBegin    = errors.New("tx begin error")
	errCommit   = errors.New("tx commit error")
	errRollback = errors.New("tx Rollback error")
	err         = errors.New("")
)

func (a *Account) Withdraw(_ *Tx, amount int) (err error) {
	defer err2.Handle(&err)
	if a.errWithdraw == errWithdraw {
		return errWithdraw
	}
	a.reserved -= amount
	a.balance -= amount
	return nil
}

func (a *Account) Deposit(_ *Tx, _ int) (err error) {
	defer err2.Handle(&err)
	if a.errDeposit == errDeposit {
		return errDeposit
	}
	return nil
}

func (t *Tx) Commit() (err error) {
	defer err2.Handle(&err)

	if t.errCommit == errCommit {
		return errCommit
	}
	return nil
}

func (t *Tx) Rollback() (err error) {
	defer err2.Handle(&err)

	if t.errRollback == errRollback {
		return errRollback
	}
	return nil
}

func (a *Account) ReserveBalance(_ *Tx, amount int) (err error) {
	defer err2.Handle(&err, nil)

	total := a.balance - a.reserved
	if total >= amount {
		a.reserved += amount
		return nil
	}
	return fmt.Errorf("no funds")
}

func (db *Database) BeginTransaction() (tx *Tx, err error) {
	defer err2.Handle(&err)

	if db.err == errBegin {
		return nil, errBegin
	}
	return &Tx{errRollback: db.errRoll}, nil
}
