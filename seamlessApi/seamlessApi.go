package seamlessApi

import (
	"encoding/json"
	"fmt"
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

func (c *userBalancesContainer) updBalance(cId callerId, w withdraw, d deposit) (err bool) {
	c.mutx.Lock()
	defer c.mutx.Unlock()

	if (c.userBalances[cId]) >= balance(w) {
		c.userBalances[cId] -= balance(w)
		c.userBalances[cId] += balance(d)
		return false

	}

	c.userBalances[cId] = 0
	return true
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

type Rpc struct {
	Jsonrpc string
	Method  string
	Params  getBalanceParams
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

func GetBalance(rpc *Rpc) int {

	if _, ok := uB.userBalances[rpc.Params.CallerId]; !ok {
		randBal := rand.Intn(300) * 100
		uB.userBalances[rpc.Params.CallerId] = balance(randBal)
	}

	a, b, e := uB.getUserBalance(rpc.Params.CallerId)
	fmt.Println("GetBalance  ", a, b, e)

	return 0
}

func WithdrawAndDeposit(rpc *Rpc) int {

	if _, ok := uB.userBalances[rpc.Params.CallerId]; !ok {
		randBal := rand.Intn(300) * 100
		uB.userBalances[rpc.Params.CallerId] = balance(randBal)
	}

	//rpc.Params.CallerId
	b := uB.updBalance(rpc.Params.CallerId, 500, 100)
	fmt.Println("WithdrawAndDeposit ", rpc, b)
	return 0
}

func RollbackTransaction(rpc *Rpc) int {
	fmt.Println("RollbackTransaction ", rpc)
	return 0
}
func NewServer() {

	uB = userBalancesContainer{
		userBalances: make(map[callerId]balance),
	}
	http.HandleFunc("/mascot/seamless", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rpc := &Rpc{}
	err := json.NewDecoder(r.Body).Decode(rpc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestMethod := rpc.Method

	//fmt.Println("requestMethod ", requestMethod)

	switch requestMethod {
	case getBalanceMethod:
		go GetBalance(rpc)
	case withdrawAndDepositMethod:
		go WithdrawAndDeposit(rpc)
	case rollbackTransactionMethod:
		go RollbackTransaction(rpc)
	default:
		fmt.Println("Херовый метод ")

	}

	w.WriteHeader(http.StatusCreated)
}
