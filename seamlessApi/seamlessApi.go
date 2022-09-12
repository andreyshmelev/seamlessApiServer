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
	jsonrpc                   = "2.0"
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
	mutx                sync.Mutex
	userBalances        map[callerId]balance
	transactionRefsList map[callerId][]string
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
	BetType string
	WinType string
}

type getBalanceResponseParams struct {
	Balance        balance
	FreeRoundsLeft freeRoundsLeft
}

type getBalanceResponse struct {
	Jsonrpc  string
	Method   string
	Params   getBalanceResponseParams
	CallerId callerId
}

type withdrawAndDepositResponse struct {
	Jsonrpc  string
	Method   string
	Params   withdrawAndDepositResponseParams
	CallerId callerId
}

type withdrawAndDepositResponseParams struct {
	NewBalance     balance
	TransactionId  transactionId
	FreeRoundsLeft freeRoundsLeft
}

func GetBalance(wd *getBalanceRpc) ([]byte, error) {

	if _, ok := uB.userBalances[wd.Params.CallerId]; !ok {
		randBal := rand.Intn(300) * 100
		uB.userBalances[wd.Params.CallerId] = balance(randBal)
	}

	cId, bal, _ := uB.getUserBalance(wd.Params.CallerId)

	resp := getBalanceResponse{
		Jsonrpc:  jsonrpc,
		Method:   getBalanceMethod,
		CallerId: cId,
		Params: getBalanceResponseParams{
			Balance: bal,
		},
	}

	return json.Marshal(resp)

}

func WithdrawAndDeposit(body []byte, w http.ResponseWriter) (b balance, error string) {

	wrpc := &withdrawAndDepositRpc{}
	err := json.Unmarshal(body, wrpc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userId := wrpc.Params.CallerId
	if _, ok := uB.userBalances[userId]; !ok {
		randBal := rand.Intn(300) * 100
		uB.userBalances[userId] = balance(randBal)
	}

	//wd.Params.CallerId
	uB.updBalance(wrpc.Params.CallerId, wrpc.Params.Withdraw, wrpc.Params.Deposit)

	resp := withdrawAndDepositResponse{
		Jsonrpc:  jsonrpc,
		Method:   withdrawAndDepositMethod,
		CallerId: userId,
		Params: withdrawAndDepositResponseParams{
			NewBalance:     84,
			TransactionId:  "TransactionId to generate",
			FreeRoundsLeft: 0,
		},
	}

	fmt.Println("WithdrawAndDeposit ", wrpc, b, err)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	jsonResp, _ := json.Marshal(resp)

	fmt.Println("jsonResp ", string(jsonResp))

	_, e := w.Write(jsonResp)

	if e != nil {
		fmt.Println("error", e)
	}

	return b, "false"
}

func RollbackTransaction(rb *rollbackTransactionRpc) int {
	fmt.Println("RollbackTransaction ", rb)
	return 0
}
func NewServer() {

	uB = userBalancesContainer{
		userBalances:        make(map[callerId]balance),
		transactionRefsList: make(map[callerId][]string),
	}

	handler := http.HandlerFunc(handler)
	http.Handle("/mascot/seamless", handler)
	http.ListenAndServe(":8080", nil)

	/*	mux := http.NewServeMux()
		mux.HandleFunc("/mascot/seamless", handler)

		err := http.ListenAndServe(":8080", mux)
		log.Fatal(err)
	*/
}

func handleRequest(w http.ResponseWriter, r *http.Request) {

	return
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

var ig int = 0

func handler(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		return
	}

	rd := base{}
	json.Unmarshal(body, &rd)

	requestMethod := rd.Method

	fmt.Println("requestMethod ", requestMethod)
	switch requestMethod {
	case getBalanceMethod:

		brpc := &getBalanceRpc{}
		err = json.Unmarshal(body, brpc)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		jsonResp, err := GetBalance(brpc)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}

		fmt.Println("jsonResp ", string(jsonResp))

		_, e := w.Write(jsonResp)

		if e != nil {
			fmt.Println("error", e)
		}

	case withdrawAndDepositMethod:

		WithdrawAndDeposit(body, w)

	case rollbackTransactionMethod:

		rrpc := &rollbackTransactionRpc{}
		err = json.Unmarshal(body, rrpc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		RollbackTransaction(rrpc)
	}

}
