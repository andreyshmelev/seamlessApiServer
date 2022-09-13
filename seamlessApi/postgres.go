package seamlessApi

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "1234"
	dbname   = "sDB"
)

func checkDb() {

	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	CheckError(err)

	rows, err := db.Query(`SELECT "userId","balance", "freeRoundsLeft" FROM "users"`)
	CheckError(err)

	defer rows.Close()

	for rows.Next() {
		var uId callerId
		var ba balance
		var fr freeRoundsLeft

		err = rows.Scan(&uId, &ba, &fr)
		CheckError(err)

		fmt.Println(uId, ba, fr)
	}
}

func getBalanceDB(cId callerId) (balance, freeRoundsLeft, error) {

	// default for test
	var bal balance = 35000
	var fr freeRoundsLeft = 2

	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		fmt.Println(err)
		return -1, -1, err
	}

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return -1, -1, err
	}

	rows, err := db.Query(`SELECT "userId","balance", "freeRoundsLeft" FROM "users" 
	where "userId" = $1`, int(cId))

	if err != nil {
		fmt.Println(err)
		return -1, -1, err
	}
	defer rows.Close()

	for rows.Next() {

		err = rows.Scan(&cId, &bal, &fr)

		fmt.Println("&uId, &ba, &fr", cId, bal, fr)

		if err != nil {
			fmt.Println(err)
			return -1, -1, err
		}
		fmt.Println(cId, bal, fr)

		return bal, fr, nil
	}
	// если нет пользователя то добавить пользователя по умолчанию (для тестов)

	insertStmt := `INSERT into "users"( "userId","balance", "freeRoundsLeft") values($1,35000,2)`
	db.Exec(insertStmt, int(cId))

	return bal, fr, nil
}

func createTransactionDB(de deposit, wd withdraw, rb rolledBack, transw transactionRef, cId callerId) error {
	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		fmt.Println(err)
		return err
	}

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return err
	}

	rows, err := db.Query(`INSERT into "transactions"( "transactionRef","rolledBack", "callerId", "withdraw", "deposit") values($1,$2,$3,$4,$5)`, transw, rb, cId, wd, de)

	if err != nil {
		fmt.Println("!!err", err)
		return err
	}
	defer rows.Close()
	return err
}
func getTransactionDB(tref transactionRef) (transw transactionRef, rb rolledBack, cId callerId, wd withdraw, de deposit, e error) {

	// default for test

	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		fmt.Println(err)
		return "", false, -1, -1, -1, err
	}

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return "", false, -1, -1, -1, err
	}

	var t transactionRef
	rows, err := db.Query(`SELECT "transactionRef","rolledBack","callerId", "withdraw" , "deposit" FROM "transactions" 
	where "transactionRef" = $1`, string(tref))

	if err != nil {
		fmt.Println(err)
		return "", false, -1, -1, -1, err
	}
	defer rows.Close()

	for rows.Next() {

		err = rows.Scan(&t, &rb, &cId, &wd, &de)

		fmt.Println("&t,&rb, &cId,&wd, &de ", t, rb, cId, wd, de)

		if err != nil {
			fmt.Println(err)
			return "", false, -1, -1, -1, err
		}

		return t, rb, cId, wd, de, nil
	}

	return t, rb, cId, wd, de, nil
}

func updateBalanceDB(cId callerId, ba balance, fr freeRoundsLeft) error {

	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		fmt.Println(err)
		return err
	}

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return err
	}

	insertStmt := `update "users" set "balance" = $1,"freeRoundsLeft" = $2 where "userId" = $3`
	_, err = db.Exec(insertStmt, ba, fr, cId)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("!!!", err)
	}
}
