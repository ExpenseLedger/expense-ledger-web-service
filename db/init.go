package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/expenseledger/web-service/config"
	dbconfig "github.com/expenseledger/web-service/config/database"
	"github.com/expenseledger/web-service/constant"
	"github.com/jmoiron/sqlx"

	// This is just a PostgreSQL driver for sqlx package
	_ "github.com/lib/pq"
)

// Table names
const (
	Transaction      = "transaction"
	AffectedWallet   = "affected_wallet"
	Category         = "category"
	Wallet           = "wallet"
	WalletTypes      = "wallet_type"
	TransactionTypes = "transaction_type"
	WalletRoles      = "wallet_role"
)

var conn *sqlx.DB

func init() {
	var (
		dbinfo string
		err    error
	)

	configs := config.GetConfigs()
	dbconfigs := dbconfig.GetConfigs()

	if configs.Mode == "PRODUCTION" {
		dbinfo = dbconfigs.DBURL
	} else {
		dbinfo = fmt.Sprintf(
			"user=%s password=%s dbname=%s port=%s sslmode=disable",
			dbconfigs.DBUser,
			dbconfigs.DBPswd,
			dbconfigs.DBName,
			dbconfigs.DBPort,
		)
	}

	conn, err = sqlx.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal("Error opening connection to the database", err)
	}
}

// Conn returns an SQL connection
func Conn() *sqlx.DB {
	return conn
}

// CreateTables creates (if not exists) all the required tables
func CreateTables() (err error) {
	err = createWalletTypeEnum()
	if err != nil {
		log.Println("Error creating enum:", WalletTypes, err)
		return
	}

	err = createTransactionTypeEnum()
	if err != nil {
		log.Println("Error creating enum:", TransactionTypes, err)
		return
	}

	err = createWalletRoleEnum()
	if err != nil {
		log.Println("Error creating enum:", WalletRoles, err)
		return
	}

	err = createWalletTable()
	if err != nil {
		log.Println("Error creating table:", Wallet, err)
		return
	}

	err = createCategoryTable()
	if err != nil {
		log.Println("Error creating table:", Category, err)
		return
	}

	err = createTransactionTable()
	if err != nil {
		log.Println("Error creating table:", Transaction, err)
		return
	}

	err = createAffectedWalletTable()
	if err != nil {
		log.Println("Error creating table:", AffectedWallet, err)
		return
	}

	err = createTriggerSetUpdatedAt(
		Wallet,
		Category,
		Transaction,
		AffectedWallet,
	)
	if err != nil {
		log.Println("Error creating trigger for updated_at", err)
		return
	}

	return
}

func createCategoryTable() (err error) {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", Category)
	query +=
		`
		name character varying(20),
		created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
		user_id character varying(128),
		PRIMARY KEY (user_id, name)
		);
		`
	_, err = conn.Exec(query)
	return
}

func createWalletTypeEnum() (err error) {
	walletType := constant.WalletTypes()
	query :=
		fmt.Sprintf(
			"CREATE TYPE %s AS ENUM ('%s', '%s', '%s');",
			WalletTypes,
			walletType.Cash,
			walletType.BankAccount,
			walletType.Credit,
		)
	_, err = conn.Exec(query)
	return filterError(err)
}

func createTransactionTypeEnum() (err error) {
	transactionType := constant.TransactionTypes()
	query :=
		fmt.Sprintf(
			"CREATE TYPE %s AS ENUM ('%s', '%s', '%s');",
			TransactionTypes,
			transactionType.Income,
			transactionType.Expense,
			transactionType.Transfer,
		)
	_, err = conn.Exec(query)
	return filterError(err)
}

func createWalletRoleEnum() (err error) {
	walletRole := constant.WalletRoles()
	query :=
		fmt.Sprintf(
			"CREATE TYPE %s AS ENUM ('%s', '%s');",
			WalletRoles,
			walletRole.SrcWallet,
			walletRole.DstWallet,
		)
	_, err = conn.Exec(query)
	return filterError(err)
}

func createWalletTable() (err error) {
	query := fmt.Sprintf(
		`
		CREATE TABLE IF NOT EXISTS %s (
			name character varying(20),
			type %s NOT NULL,
			balance NUMERIC(11, 2) NOT NULL DEFAULT 0.00,
			created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
			user_id character varying(128),
			PRIMARY KEY (name, user_id)
		);
		`,
		Wallet,
		WalletTypes,
	)

	_, err = conn.Exec(query)
	return
}

//wallet character varying(20) NOT NULL REFERENCES %s,

func createTransactionTable() (err error) {
	query := fmt.Sprintf(
		`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		CREATE TABLE IF NOT EXISTS %s (
			id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			amount NUMERIC(11, 2) NOT NULL DEFAULT 0.00 CHECK (amount >= 0),
			type %s NOT NULL,
			category character varying(20) NOT NULL ,
			description text NOT NULL DEFAULT '',
			occurred_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
			user_id character varying(128),
			FOREIGN KEY (category, user_id) REFERENCES %s (name, user_id)
		);
		`,
		Transaction,
		TransactionTypes,
		Category,
	)

	_, err = conn.Exec(query)
	return
}

func createAffectedWalletTable() (err error) {
	query := fmt.Sprintf(
		`
		CREATE TABLE IF NOT EXISTS %s (
			transaction_id uuid NOT NULL REFERENCES %s,
			wallet character varying(20) NOT NULL,
			role %s NOT NULL,
			created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
			user_id character varying(128),
			PRIMARY KEY (transaction_id, wallet, role),
			FOREIGN KEY (wallet, user_id) REFERENCES %s (name, user_id)
		);
		`,
		AffectedWallet,
		Transaction,
		WalletRoles,
		Wallet,
	)

	_, err = conn.Exec(query)
	return
}

func createTriggerSetUpdatedAt(tableNames ...string) (err error) {
	query := deleteExistingTriggers(tableNames)
	query += "CREATE EXTENSION IF NOT EXISTS moddatetime;"

	for _, tableName := range tableNames {
		query += fmt.Sprintf(
			`
			CREATE TRIGGER %s
			BEFORE UPDATE ON %s
			FOR EACH ROW
			EXECUTE PROCEDURE moddatetime (updated_at);
			`,
			"mdt_"+tableName,
			tableName,
		)
	}

	_, err = conn.Exec(query)
	return
}

func deleteExistingTriggers(tableNames []string) string {
	var query string

	for _, tableName := range tableNames {
		query += fmt.Sprintf(
			"DROP TRIGGER IF EXISTS %s ON %s;",
			"mdt_"+tableName,
			tableName,
		)
	}

	return query
}

func filterError(err error) error {
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return nil
	}
	return err
}
