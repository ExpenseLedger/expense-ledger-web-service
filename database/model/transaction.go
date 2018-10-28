package model

import (
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// Transaction the structure represents a stored transaction on database
type Transaction struct {
	ID          string          `db:"id"`
	SrcWallet   string          `db:"src_wallet"`
	DstWallet   *string         `db:"dst_wallet"`
	Amount      decimal.Decimal `db:"amount"`
	Type        string          `db:"type"`
	Category    string          `db:"category"`
	Description string          `db:"description"`
	OccuredAt   *time.Time      `db:"occured_at"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

// Transactions is defined just to be used as a receiver
type Transactions []Transaction

// InsertExpense ...
func (wallet *Wallet) InsertExpense(tx *Transaction) error {
	var txQuery, walletQuery string

	if tx.OccuredAt != nil {
		txQuery =
			`
			INSERT INTO transaction
			(src_wallet, amount, type, category, description, occured_at)
			VALUES
			(:src_wallet, :amount, :type, :category, :description, :occured_at)
			RETURNING *;
			`
	} else {
		txQuery =
			`
			INSERT INTO transaction
			(src_wallet, amount, type, category, description)
			VALUES
			(:src_wallet, :amount, :type, :category, :description)
			RETURNING *;
			`
	}

	walletQuery =
		`
		UPDATE wallet
		SET balance=balance-$1
		WHERE name=$2
		RETURNING *;
		`

	dbTx, err := db.Beginx()
	if err != nil {
		log.Println("Error beginning a transaction", err)
		return err
	}

	namedStmt, err := dbTx.PrepareNamed(txQuery)
	if err != nil {
		log.Println("Error inserting a transaction", err)

		if err := dbTx.Rollback(); err != nil {
			log.Println("Error rolling back a transaction", err)
			return err
		}

		return err
	}

	if err := namedStmt.Get(tx, tx); err != nil {
		log.Println("Error inserting a transaction", err)

		if err := dbTx.Rollback(); err != nil {
			log.Println("Error rolling back a transaction", err)
			return err
		}

		return err
	}

	stmt, err := dbTx.Preparex(walletQuery)
	if err != nil {
		log.Println("Error updating a wallet", err)

		if err := dbTx.Rollback(); err != nil {
			log.Println("Error rolling back a transaction", err)
			return err
		}

		return err
	}

	if err := stmt.Get(wallet, tx.Amount, wallet.Name); err != nil {
		log.Println("Error updating a wallet", err)

		if err := dbTx.Rollback(); err != nil {
			log.Println("Error rolling back a transaction", err)
			return err
		}

		return err
	}

	if err := dbTx.Commit(); err != nil {
		log.Println("Error committing a transaction", err)

		if err := dbTx.Rollback(); err != nil {
			log.Println("Error rolling back a transaction", err)
			return err
		}

		return err
	}

	return nil
}

// DeleteAll ...
func (transactions *Transactions) DeleteAll() (int, error) {
	query :=
		`
		DELETE FROM transaction
		RETURNING *;
		`

	stmt, err := db.Preparex(query)
	if err != nil {
		log.Println("Error deleting all transactions", err)
		return 0, err
	}

	if err := stmt.Select(transactions); err != nil {
		log.Println("Error deleting all transactions", err)
		return 0, err
	}

	return len(*transactions), nil
}
