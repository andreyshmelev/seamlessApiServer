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

type result string

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
	mutx                    sync.Mutex
	userBalances            map[callerId]balance
	userFreeRoundsRemaining map[callerId]freeRoundsLeft
	transactionRefsList     map[callerId][]string
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
	Jsonrpc string
	Method  string
	Result  getBalanceResponseParams
	Id      id
}

type withdrawAndDepositResponse struct {
	Jsonrpc string
	Method  string
	Result  withdrawAndDepositResponseParams
	Id      id
}

type withdrawAndDepositResponseParams struct {
	NewBalance     balance
	TransactionId  transactionId
	FreeRoundsLeft freeRoundsLeft
}

type rollbackTransactionResponse struct {
	Jsonrpc  string
	Method   string
	Id       id
	Result   rollbackTransactionResponseParams
	CallerId callerId
}

type rollbackTransactionResponseParams struct {
	Result result
}

var uB userBalancesContainer

func (c *userBalancesContainer) updBalance(cId callerId, wdraw withdraw, depo deposit, cfree chargeFreerounds) (b balance, err bool) {
	c.mutx.Lock()
	defer c.mutx.Unlock()
	if _, ok := uB.userBalances[cId]; !ok {
		randBal := rand.Intn(300) * 100
		uB.userBalances[cId] = balance(randBal)
	}
	if _, ok := uB.userFreeRoundsRemaining[cId]; !ok {
		randBal := rand.Intn(2)
		uB.userFreeRoundsRemaining[cId] = freeRoundsLeft(randBal)
	}

	freeRundsLeft := int(c.userFreeRoundsRemaining[cId])
	if wdraw > 0 && freeRundsLeft >= int(cfree) {
		c.userFreeRoundsRemaining[cId]-- // reduce free rounds and ignore debiting
		c.userBalances[cId] += balance(depo)
		return c.userBalances[cId], false
	} else {
		if (c.userBalances[cId]) >= balance(wdraw) {
			c.userBalances[cId] -= balance(wdraw) //  decrease  the amount from the withdraw field.
			c.userBalances[cId] += balance(depo)
			return b, false
		} else {
			c.userBalances[cId] = 0
			return 0, true
		}
	}

}

func (c *userBalancesContainer) getUserBalance(cId callerId) (balance, freeRoundsLeft, bool) {
	c.mutx.Lock()
	defer c.mutx.Unlock()
	bal := c.userBalances[cId]
	fr := c.userFreeRoundsRemaining[cId]
	return bal, fr, false
}

func GetBalance(wd *getBalanceRpc) ([]byte, error) {

	userId := wd.Params.CallerId

	if _, ok := uB.userBalances[userId]; !ok {
		randBal := rand.Intn(300) * 100
		uB.userBalances[userId] = balance(randBal)
	}

	if _, ok := uB.userFreeRoundsRemaining[userId]; !ok {
		randBal := rand.Intn(5)
		uB.userFreeRoundsRemaining[userId] = freeRoundsLeft(randBal)
	}

	bal, frleft, _ := uB.getUserBalance(wd.Params.CallerId)

	resp := getBalanceResponse{
		Jsonrpc: jsonrpc,
		Method:  getBalanceMethod,
		Id:      id(userId),
		Result: getBalanceResponseParams{
			Balance: bal,
		},
	}

	if frleft > 0 {
		resp.Result = getBalanceResponseParams{
			Balance:        bal,
			FreeRoundsLeft: frleft,
		}
	}

	return json.Marshal(resp)

}

func WithdrawAndDeposit(body []byte, wdraw http.ResponseWriter) (b balance, error string) {

	wrpc := &withdrawAndDepositRpc{}
	err := json.Unmarshal(body, wrpc)
	if err != nil {
		http.Error(wdraw, err.Error(), http.StatusBadRequest)
		return
	}

	userId := wrpc.Params.CallerId
	wd := wrpc.Params.Withdraw
	de := wrpc.Params.Deposit
	fr := wrpc.Params.ChargeFreerounds

	uB.updBalance(userId, wd, de, fr)

	bal, frleft, _ := uB.getUserBalance(userId)

	generatedTransId := transactionId("TransactionId to generate")

	resp := withdrawAndDepositResponse{
		Jsonrpc: jsonrpc,
		Method:  withdrawAndDepositMethod,
		Id:      id(userId),
		Result: withdrawAndDepositResponseParams{
			NewBalance:    bal,
			TransactionId: generatedTransId,
		},
	}

	if frleft > 0 {
		resp.Result = withdrawAndDepositResponseParams{
			NewBalance:     bal,
			TransactionId:  generatedTransId,
			FreeRoundsLeft: frleft,
		}
	}

	wdraw.WriteHeader(http.StatusCreated)
	wdraw.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	jsonResp, _ := json.Marshal(resp)

	fmt.Println("jsonResp ", string(jsonResp))

	_, e := wdraw.Write(jsonResp)

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
		userBalances:            make(map[callerId]balance),
		userFreeRoundsRemaining: make(map[callerId]freeRoundsLeft),
		transactionRefsList:     make(map[callerId][]string),
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

func handleRequest(wdraw http.ResponseWriter, r *http.Request) {

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

func handler(wdraw http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if r.Method != http.MethodPost {
		http.Error(wdraw, "Method not allowed", http.StatusMethodNotAllowed)
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
			http.Error(wdraw, err.Error(), http.StatusBadRequest)
			return
		}

		jsonResp, err := GetBalance(brpc)

		wdraw.WriteHeader(http.StatusCreated)
		wdraw.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}

		fmt.Println("jsonResp ", string(jsonResp))

		_, e := wdraw.Write(jsonResp)

		if e != nil {
			fmt.Println("error", e)
		}

	case withdrawAndDepositMethod:

		WithdrawAndDeposit(body, wdraw)

	case rollbackTransactionMethod:

		rrpc := &rollbackTransactionRpc{}
		err = json.Unmarshal(body, rrpc)
		if err != nil {
			http.Error(wdraw, err.Error(), http.StatusBadRequest)
			return
		}
		RollbackTransaction(rrpc)
	}

}
