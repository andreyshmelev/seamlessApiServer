package seamlessApi

import (
	"encoding/json"
	"fmt"
)

type apifunc func() int

type balance int
type newBalance int
type transactionId string
type freeRoundsLeft int

// common parameters
type callerId int
type playerName string
type currency string
type gameId string
type sessionId string
type sessionAlternativeId string

// getBalance parameters
type id int
type bonusId string

// withdrawAndDeposit parameters
type withdraw int
type deposit int
type transactionRef string
type gameRoundRef string
type source string
type reason string
type spinDetails spinDetailsObject
type chargeFreerounds int

// rollbackTransaction parameters
type roundId string

type Params struct {
	CallerId   int
	PlayerName string
	Currency   string
	GameId     string
}

type Rpc struct {
	Jsonrpc string
	Method  string
	Params  Params
	Id      int
}

type spinDetailsObject struct {
	betType string
	winType string
}

func GetBalance() int {
	fmt.Println("  GetBalance")

	return 789
}

func WithdrawAndDeposit(cId callerId, w withdraw, dep deposit) int {
	fmt.Println("  WithdrawAndDeposit")

	return 0
}

func RollbackTransaction() int {
	fmt.Println("  RollbackTransaction")

	return 0
}
func NewServer() {

	rJson := `{
		"jsonrpc": "2.0",
		"method": "getBalance",
		"Params": {
		  "callerId": 1,
		  "playerName": "player1",
		  "currency": "EUR",
		  "gameId": "riot"
		},
		"id": 0
	  }`

	var r Rpc
	json.Unmarshal([]byte(rJson), &r)

	fmt.Println(r)
}
