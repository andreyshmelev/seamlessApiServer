package seamlessApi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

type apifunc func() int

type balance int
type newBalance int
type transactionId string
type freeRoundsLeft int

const (
	getBalanceMethod          = "getBalance"
	withdrawAndDepositMethod  = "withdrawAndDeposit"
	rollbackTransactionMethod = "rollbackTransaction"
)

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

type userBalancesContainer struct {
	mutx         sync.Mutex
	userBalances map[callerId]balance
}

var uB userBalancesContainer

func (c *userBalancesContainer) updBalance(cId callerId, w withdraw, d deposit) (b balance, err bool) {
	c.mutx.Lock()
	defer c.mutx.Unlock()

	if (c.userBalances[cId]) >= balance(w) {
		c.userBalances[cId] -= balance(w)
		c.userBalances[cId] += balance(d)
		return b, false

	}

	c.userBalances[cId] = 0
	return 0, true
}

func (c *userBalancesContainer) getUserBalance(cId callerId) (callerId, balance, bool) {
	c.mutx.Lock()
	defer c.mutx.Unlock()
	bal := c.userBalances[cId]
	return cId, bal, false
}

type getBalanceParams struct {
	//requred
	CallerId   callerId
	PlayerName playerName
	Currency   currency
	//not requred
	GameId               gameId
	SessionId            sessionId
	SessionAlternativeId sessionAlternativeId
	BonusId              bonusId
}
type withdrawAndDepositParams struct {
	//requred
	CallerId       callerId
	PlayerName     playerName
	Withdraw       withdraw
	Deposit        deposit
	Currency       currency
	TransactionRef transactionRef
	//not requred
	GameId               gameId
	Source               source
	Reason               reason
	SessionId            sessionId
	SessionAlternativeId sessionAlternativeId
	SpinDetails          spinDetails
	BonusId              bonusId
	ChargeFreerounds     chargeFreerounds
}

type rollbackTransactionParams struct {
	//requred
	CallerId       callerId
	PlayerName     playerName
	TransactionRef transactionRef
	//not requred
	GameId               gameId
	SessionId            sessionId
	SessionAlternativeId sessionAlternativeId
	RoundId              roundId
}

type base struct {
	Jsonrpc string
	Method  string
	Params  string
	Id      id
}

type getBalanceRpc struct {
	Jsonrpc string
	Method  string
	Params  getBalanceParams
	Id      id
}

type withdrawAndDepositRpc struct {
	Jsonrpc string
	Method  string
	Params  withdrawAndDepositParams
	Id      id
}

type rollbackTransactionRpc struct {
	Jsonrpc string
	Method  string
	Params  rollbackTransactionParams
	Id      id
}

type spinDetailsObject struct {
	betType string
	winType string
}

type getBalanceResponseResult struct {
	balance        balance
	freeRoundsLeft freeRoundsLeft
}

type getBalanceResponse struct {
	Jsonrpc string
	Method  string
	Params  getBalanceResponseResult
	Id      id
}

func GetBalance(wd *getBalanceRpc) balance {

	if _, ok := uB.userBalances[wd.Params.CallerId]; !ok {
		randBal := rand.Intn(300) * 100
		uB.userBalances[wd.Params.CallerId] = balance(randBal)
	}

	a, b, e := uB.getUserBalance(wd.Params.CallerId)
	fmt.Println("GetBalance  ", a, b, e)

	return b
}

func WithdrawAndDeposit(wd *withdrawAndDepositRpc) balance {

	if _, ok := uB.userBalances[wd.Params.CallerId]; !ok {
		randBal := rand.Intn(300) * 100
		uB.userBalances[wd.Params.CallerId] = balance(randBal)
	}

	//wd.Params.CallerId
	b, err := uB.updBalance(wd.Params.CallerId, wd.Params.Withdraw, wd.Params.Deposit)
	fmt.Println("WithdrawAndDeposit ", wd, b, err)
	return b
}

func RollbackTransaction(rb *rollbackTransactionRpc) int {
	fmt.Println("RollbackTransaction ", rb)
	return 0
}
func NewServer() {

	uB = userBalancesContainer{
		userBalances: make(map[callerId]balance),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mascot/seamless", handler)

	err := http.ListenAndServe(":8080", mux)
	log.Fatal(err)

}

func handler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	rd := base{}
	json.Unmarshal(body, &rd)

	requestMethod := rd.Method

	switch requestMethod {
	case getBalanceMethod:

		brpc := &getBalanceRpc{}
		err = json.Unmarshal(body, brpc)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}
		go GetBalance(brpc)

	case withdrawAndDepositMethod:

		wrpc := &withdrawAndDepositRpc{}
		err = json.Unmarshal(body, wrpc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		go WithdrawAndDeposit(wrpc)

	case rollbackTransactionMethod:

		rrpc := &rollbackTransactionRpc{}
		err = json.Unmarshal(body, rrpc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		go RollbackTransaction(rrpc)
	}
	w.WriteHeader(http.StatusCreated)
}
