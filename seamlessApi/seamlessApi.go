package seamlessApi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
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
const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

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
	transactionRefsList     map[transactionRef]rollbackStruct
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

type rollbackStruct struct {
	tRef       transactionRef
	rolledBack bool
	withdr     withdraw
	dep        deposit
	fr         chargeFreerounds
	uId        callerId
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
	Jsonrpc string
	Method  string
	Id      id
	Result  rollbackTransactionResponseParams
}

type rollbackTransactionResponseParams struct {
	Result result `json:"Result,omitempty"`
}

var uB userBalancesContainer

func RandStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (c *userBalancesContainer) updBalance(uId callerId, wdraw withdraw, depo deposit, cfree chargeFreerounds) (b balance, err bool) {
	c.mutx.Lock()
	defer c.mutx.Unlock()

	if _, ok := uB.userBalances[uId]; !ok {
		randBal := rand.Intn(300) * 100
		uB.userBalances[uId] = balance(randBal)
	}
	if _, ok := uB.userFreeRoundsRemaining[uId]; !ok {
		randBal := rand.Intn(2)
		uB.userFreeRoundsRemaining[uId] = freeRoundsLeft(randBal)
	}

	freeRundsLeft := int(c.userFreeRoundsRemaining[uId])
	if wdraw > 0 && freeRundsLeft >= int(cfree) {
		c.userFreeRoundsRemaining[uId] -= freeRoundsLeft(cfree) // reduce free rounds and ignore debiting
		if c.userFreeRoundsRemaining[uId] < 0 {
			c.userFreeRoundsRemaining[uId] = 0
		}
		c.userBalances[uId] += balance(depo)
		return c.userBalances[uId], false
	} else {
		if (c.userBalances[uId]) >= balance(wdraw) {
			c.userBalances[uId] -= balance(wdraw) //  decrease  the amount from the withdraw field.
			c.userBalances[uId] += balance(depo)
			return b, false
		} else {
			c.userBalances[uId] = 0
			return 0, true
		}
	}

}

func (c *userBalancesContainer) rollbackBalance(uId callerId, wdraw withdraw, depo deposit, cfree chargeFreerounds) (b balance, err bool) {
	c.mutx.Lock()
	defer c.mutx.Unlock()

	c.userFreeRoundsRemaining[uId] += freeRoundsLeft(cfree)

	c.userBalances[uId] -= balance(depo)
	c.userBalances[uId] += balance(wdraw)
	return c.userBalances[uId], false
}

func (c *userBalancesContainer) getUserBalance(uId callerId) (balance, freeRoundsLeft, bool) {
	c.mutx.Lock()
	defer c.mutx.Unlock()
	bal := c.userBalances[uId]
	fr := c.userFreeRoundsRemaining[uId]
	return bal, fr, false
}

func GetBalance(body []byte, wdraw http.ResponseWriter) (error string) {

	brpc := &getBalanceRpc{}
	err := json.Unmarshal(body, brpc)
	if err != nil {
		http.Error(wdraw, err.Error(), http.StatusBadRequest)
		return "true"
	}

	userId := brpc.Params.CallerId

	if _, ok := uB.userBalances[userId]; !ok {
		randBal := rand.Intn(300) * 100
		uB.userBalances[userId] = balance(randBal)
	}

	if _, ok := uB.userFreeRoundsRemaining[userId]; !ok {
		randBal := rand.Intn(5)
		uB.userFreeRoundsRemaining[userId] = freeRoundsLeft(randBal)
	}

	bal, frleft, _ := uB.getUserBalance(brpc.Params.CallerId)

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
	return "false"
}

func WithdrawAndDeposit(body []byte, wdraw http.ResponseWriter) (b balance, error string) {

	rbrpc := &withdrawAndDepositRpc{}
	err := json.Unmarshal(body, rbrpc)
	if err != nil {
		http.Error(wdraw, err.Error(), http.StatusBadRequest)
		return
	}

	userId := rbrpc.Params.CallerId
	wd := rbrpc.Params.Withdraw
	de := rbrpc.Params.Deposit
	fr := rbrpc.Params.ChargeFreerounds
	tr := rbrpc.Params.TransactionRef

	if _, ok := uB.transactionRefsList[tr]; !ok {
		// если  подобный transactionRefs отсутствует то создаем

		entry := rollbackStruct{}
		entry.dep = de
		entry.withdr = wd
		entry.rolledBack = false
		entry.tRef = tr
		entry.uId = userId
		entry.fr = fr
		uB.transactionRefsList[tr] = entry

	} else {

		http.Error(wdraw, "Operation already Rolled Back", http.StatusBadRequest)
		fmt.Println("Operation already Rolled Back")

		return
	}

	uB.updBalance(userId, wd, de, fr)

	bal, frleft, _ := uB.getUserBalance(userId)

	generatedTransId := transactionId(RandStringBytes(18))

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

func RollbackTransaction(body []byte, wdraw http.ResponseWriter) (b balance, error string) {

	rbrpc := &rollbackTransactionRpc{}
	err := json.Unmarshal(body, rbrpc)
	if err != nil {
		http.Error(wdraw, err.Error(), http.StatusBadRequest)
		return
	}

	tr := rbrpc.Params.TransactionRef

	if _, ok := uB.transactionRefsList[tr]; !ok {
		// если  подобный transactionRefs отсутствует то создаем и помечаем как откаченый

		entry := rollbackStruct{}
		entry.rolledBack = true
		entry.tRef = tr
		uB.transactionRefsList[tr] = entry

	} else {
		entry := uB.transactionRefsList[tr]
		uB.rollbackBalance(entry.uId, entry.withdr, entry.dep, entry.fr)
		return
	}

	userId := rbrpc.Params.CallerId

	resp := rollbackTransactionResponse{
		Jsonrpc: jsonrpc,
		Method:  rollbackTransactionMethod,
		Id:      id(userId),
		Result:  rollbackTransactionResponseParams{},
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
func NewServer() {

	uB = userBalancesContainer{
		userBalances:            make(map[callerId]balance),
		userFreeRoundsRemaining: make(map[callerId]freeRoundsLeft),
		transactionRefsList:     make(map[transactionRef]rollbackStruct),
	}

	handler := http.HandlerFunc(handler)
	http.Handle("/mascot/seamless", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(wdraw http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(wdraw, err.Error(), http.StatusMethodNotAllowed)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(wdraw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rd := base{}
	json.Unmarshal(body, &rd)
	requestMethod := rd.Method

	switch requestMethod {
	case getBalanceMethod:
		GetBalance(body, wdraw)
	case withdrawAndDepositMethod:
		WithdrawAndDeposit(body, wdraw)
	case rollbackTransactionMethod:
		RollbackTransaction(body, wdraw)
	}
}
