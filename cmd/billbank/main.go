package main

import (
	"fmt"
	"os"
	"time"

	_ "embed"

	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/db"
)

func main() {
	defer func() {
		err := os.Remove("hello.db")
		if err != nil {
			panic(err)
		}
	}()
	sdb := db.NewSqliteDb("hello", lib.USD)
	defer sdb.Close()

	err := sdb.CreateMonth(time.Date(2024, 8, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		panic(err)
	}

	accNum := "182383418234"
	pass := "johnny"
	notes := "protected back account notes"

	bankConfig := db.BankAccountConfig{
		Name:          "CitiBank",
		Password:      &pass,
		AccountNumber: &accNum,
		Notes:         &notes,
	}

	sdb.CreateBankAccount(bankConfig)

	dBankInfo, err := sdb.QueryDecryptedBankAccount(1, pass)
	if err != nil {
		panic(err)
	}
	fmt.Println(dBankInfo.Name, "|", *dBankInfo.AccountNumber, "|", *dBankInfo.Notes)

	bInfo, err := sdb.QueryAllBankAccounts()
	if err != nil {
		panic(err)
	}
	fmt.Println("Query All Bank Accounts:", bInfo)

	amount := lib.NewCurrency("768.21", lib.USD)
	sdb.CreateIncome("paycheck", amount, db.BIWEEKLY)

	sdb.CreateIncomeHistory(1, 1)

	ih, err := sdb.QueryIncomeHistory(db.QueryMap{
		db.BY_INCOME_ID: 1,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(ih)
}
