package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// TODO: test migrations with these samples!

func (db *Database) MoneyTransfer(from, to *Account, amount int) (err error) {
	defer err2.Handle(&err)

	tx := try.To1(db.BeginTransaction())
	defer err2.Handle(&err, func() {
		if errRoll := tx.Rollback(); errRoll != nil {
			// with go 1.20: err = fmt.Errorf("%w: ROLLBACK ERROR: %w", err, errRoll)
			err = fmt.Errorf("%v: ROLLBACK ERROR: %w", err, errRoll)
		}
	})

	try.To(from.ReserveBalance(tx, amount))

	defer err2.Handle(&err, func() { // optional, following sample's wording
		err = fmt.Errorf("cannot %w", err)
	})

	try.To(from.Withdraw(tx, amount))
	try.To(to.Deposit(tx, amount))
	try.To(tx.Commit())

	return nil
}

func doDBMain() {
	err2.SetErrorTracer(os.Stderr)
	err2.SetErrorTracer(nil) // <- out-comment/rm to get automatic error traces

	defer err2.Catch("CATCH Warning: %s", "test-name")

	db, from, to := new(Database), new(Account), new(Account)

	// --- TODO: play with these lines to simulate different errors:
	db.errRoll = errRollback
	//db.err = errBegin              // tx fails
	from.balance = 1100            // no enough funds
	from.errWithdraw = errWithdraw // withdraw error
	to.errDeposit = errDeposit     // deposit error
	amount := 100                  // no enough funds
	// --- TODO: simulation variables end

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
