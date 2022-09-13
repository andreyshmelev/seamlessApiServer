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

type balance int
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
type rolledBack bool
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
	CallerId   callerId   `json:"callerId"`
	PlayerName playerName `json:"playerName"`
	Currency   currency   `json:"currency"`
	//not requred
	GameId               gameId               `json:"gameId"`
	SessionId            sessionId            `json:"sessionId"`
	SessionAlternativeId sessionAlternativeId `json:"sessionAlternativeId"`
	BonusId              bonusId              `json:"bonusId"`
}
type withdrawAndDepositParams struct {
	//requred
	CallerId       callerId       `json:"callerId"`
	PlayerName     playerName     `json:"playerName"`
	Withdraw       withdraw       `json:"withdraw"`
	Deposit        deposit        `json:"deposit"`
	Currency       currency       `json:"currency"`
	TransactionRef transactionRef `json:"transactionRef"`
	//not requred
	GameId               gameId               `json:"gameId"`
	Source               source               `json:"source"`
	Reason               reason               `json:"reason"`
	SessionId            sessionId            `json:"sessionId"`
	SessionAlternativeId sessionAlternativeId `json:"sessionAlternativeId"`
	SpinDetails          spinDetails          `json:"spinDetails"`
	BonusId              bonusId              `json:"bonusId"`
	ChargeFreerounds     chargeFreerounds     `json:"chargeFreerounds"`
	GameRoundRef         gameRoundRef         `json:"gameRoundRef"`
}

type rollbackStruct struct {
	tRef        transactionRef
	rolledBack  bool
	withdr      withdraw
	dep         deposit
	freeRndLeft chargeFreerounds
	uId         callerId
}

type rollbackTransactionParams struct {
	//requred
	CallerId       callerId       `json:"callerId"`
	PlayerName     playerName     `json:"playerName"`
	TransactionRef transactionRef `json:"transactionRef"`
	//not requred
	GameId               gameId               `json:"gameId"`
	SessionId            sessionId            `json:"sessionId"`
	SessionAlternativeId sessionAlternativeId `json:"sessionAlternativeId"`
	RoundId              roundId              `json:"roundId"`
}

type base struct {
	Jsonrpc string
	Method  string
	Params  string
	Id      id
}

type errorStr struct {
	Jsonrpc string   `json:"jsonrpc"`
	Id      id       `json:"id"`
	Error   errorPar `json:"error"`
}

type errorPar struct {
	Code    int    `json:"code"`
	Message string `json:"messsage"`
}

type getBalanceRpc struct {
	Jsonrpc string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  getBalanceParams `json:"params"`
	Id      id               `json:"id"`
}

type withdrawAndDepositRpc struct {
	Jsonrpc string                   `json:"jsonrpc"`
	Method  string                   `json:"method"`
	Params  withdrawAndDepositParams `json:"params"`
	Id      id                       `json:"id"`
}

type rollbackTransactionRpc struct {
	Jsonrpc string                    `json:"jsonrpc"`
	Method  string                    `json:"method"`
	Params  rollbackTransactionParams `json:"params"`
	Id      id                        `json:"id"`
}

type spinDetailsObject struct {
	BetType string `json:"betType"`
	WinType string `json:"winType"`
}

type getBalanceResponseParams struct {
	Balance        balance        `json:"balance"`
	FreeRoundsLeft freeRoundsLeft `json:"freeRoundsLeft"`
}

type getBalanceResponse struct {
	Jsonrpc string                   `json:"jsonrpc"`
	Method  string                   `json:"method"`
	Result  getBalanceResponseParams `json:"result"`
	Id      id                       `json:"id"`
}

type withdrawAndDepositResponse struct {
	Jsonrpc string                           `json:"jsonrpc"`
	Method  string                           `json:"method"`
	Result  withdrawAndDepositResponseParams `json:"result"`
	Id      id                               `json:"id"`
}

type withdrawAndDepositResponseParams struct {
	NewBalance     balance        `json:"newBalance"`
	TransactionId  transactionId  `json:"transactionId"`
	FreeRoundsLeft freeRoundsLeft `json:"freeRoundsLeft"`
}

type rollbackTransactionResponse struct {
	Jsonrpc string                            `json:"jsonrpc"`
	Method  string                            `json:"method"`
	Id      id                                `json:"id"`
	Result  rollbackTransactionResponseParams `json:"result"`
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

func (c *userBalancesContainer) rollbackBalance(uId callerId, wdraw withdraw, depo deposit, charge chargeFreerounds) (b balance, err bool) {
	c.mutx.Lock()
	defer c.mutx.Unlock()
	c.userFreeRoundsRemaining[uId] += freeRoundsLeft(charge)
	c.userBalances[uId] -= balance(depo)
	c.userBalances[uId] += balance(wdraw)
	return c.userBalances[uId], false
}

func GetBalance(body []byte, wdraw http.ResponseWriter) error {
	brpc := &getBalanceRpc{}
	err := json.Unmarshal(body, brpc)
	if err != nil {
		http.Error(wdraw, err.Error(), http.StatusBadRequest)
		return err
	}

	userId := brpc.Params.CallerId

	bal, fr, _ := getBalanceDB(userId)

	resp := getBalanceResponse{
		Jsonrpc: jsonrpc,
		Method:  getBalanceMethod,
		Id:      id(userId),
		Result: getBalanceResponseParams{
			Balance: bal,
		},
	}

	if fr > 0 {
		resp.Result = getBalanceResponseParams{
			Balance:        bal,
			FreeRoundsLeft: fr,
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
	return nil
}

func WithdrawAndDeposit(body []byte, wdraw http.ResponseWriter) (b balance, error string) {

	rbrpc := &withdrawAndDepositRpc{}
	err := json.Unmarshal(body, rbrpc)
	if err != nil {
		http.Error(wdraw, err.Error(), http.StatusBadRequest)
		return
	}

	cId := rbrpc.Params.CallerId
	wd := rbrpc.Params.Withdraw
	de := rbrpc.Params.Deposit
	freeRndLeft := rbrpc.Params.ChargeFreerounds
	tr := rbrpc.Params.TransactionRef

	t, _, _, _, _, _ := getTransactionDB(tr)
	// если  подобный transactionRefs присутствует то выходим
	if t == tr {
		http.Error(wdraw, "Operation already Rolled Back", http.StatusBadRequest)
		fmt.Println("Operation already Rolled Back")
		return
	}

	// если  подобный transactionRefs отсутствует то создаем

	createTransactionDB(de, wd, false, tr, cId)

	// если ставка больше чем депозит то ошибка нет денег

	ba, _, er := getBalanceDB(cId)

	if er != nil {
		http.Error(wdraw, er.Error(), http.StatusBadRequest)
		return
	}

	var errorCode int
	var errMessage string

	if int(wd) > int(ba) {
		errorCode = 1
		errMessage = "ErrNotEnoughMoneyCode"
	}
	if de < 0 {
		errorCode = 3
		errMessage = "ErrNegativeDepositCode"
	}
	if wd < 0 {
		errorCode = 4
		errMessage = "ErrNegativeWithdrawalCode"
	}
	if errorCode > 0 {

		resp := errorStr{
			Id:      id(cId),
			Jsonrpc: jsonrpc,

			Error: errorPar{
				Code:    errorCode,
				Message: errMessage,
			},
		}

		jsonResp, _ := json.Marshal(resp)

		fmt.Println("jsonResp ", string(jsonResp))

		_, e := wdraw.Write(jsonResp)

		if e != nil {
			fmt.Println("error", e)
		}
		return
	}

	gb, frl, er := calcBalance(cId, wd, de, freeRndLeft)

	if er != nil {
		http.Error(wdraw, er.Error(), http.StatusBadRequest)
		return
	}

	generatedTransId := transactionId(RandStringBytes(18))

	resp := withdrawAndDepositResponse{
		Jsonrpc: jsonrpc,
		Method:  withdrawAndDepositMethod,
		Id:      id(cId),
		Result: withdrawAndDepositResponseParams{
			NewBalance:    gb,
			TransactionId: generatedTransId,
		},
	}

	if freeRndLeft > 0 {
		resp.Result = withdrawAndDepositResponseParams{
			NewBalance:     gb,
			TransactionId:  generatedTransId,
			FreeRoundsLeft: freeRoundsLeft(frl),
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

func calcBalance(cId callerId, wdraw withdraw, depo deposit, charge chargeFreerounds) (balance, freeRoundsLeft, error) {
	ba, freeRndLeft, er := getBalanceDB(cId)
	if er != nil {
		return -1, -1, er
	}
	if wdraw > 0 && int(freeRndLeft) >= int(charge) && freeRndLeft > 0 && charge > 0 {
		freeRndLeft -= freeRoundsLeft(charge) // reduce free rounds and ignore debiting
		if freeRndLeft < 0 {
			freeRndLeft = 0
		}
		ba += balance(depo)
	} else {
		if ba >= balance(wdraw) {
			ba -= balance(wdraw) //  decrease  the amount from the withdraw field.
			ba += balance(depo)
		} else {
			ba = 0
			freeRndLeft = 0
			e := fmt.Errorf("Balance and withdraw conflict")
			return ba, freeRndLeft, e
		}
	}
	e := updateBalanceDB(cId, ba, freeRndLeft)
	return ba, freeRndLeft, e
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
		uB.rollbackBalance(entry.uId, entry.withdr, entry.dep, entry.freeRndLeft)
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

	checkDb()
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
