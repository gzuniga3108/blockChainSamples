package main

import (
	//"bufio"
	//"bytes"
	//"crypto/aes"
	//"crypto/cipher"
	//"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
	//"github.com/op/go-logging"
	/*"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"	
	"strings"
	"time"*/
	// "github.com/errorpkg"
)

/////////////////////////// OBJECTS STRUCTURES ////////////////////////////////////////////////////////
type SimpleChaincode struct {
}

type Table struct{
	Name string
	Keys int
}

type UserObject struct{
	UserId 		string
	Name    	string
	Password 	string
	RecType 	string
	UserType 	string
	CashBalance string
}

type ItemObject struct {
	ItemID         string
	RecType        string
	ItemDesc       string
	ItemDetail     string // Could included details such as who created the Art work if item is a Painting
	ItemDate       string
	ItemType       string
	ItemSubject    string	
	ItemImage	   string
	CurrentOwnerID string // This is validated for a user registered record
	ItemStatus	   string
}

///////////////////////// GLOBAL VARIABLES ////////////////////////////////////////////////////////////
//Tables that will be used in the application
var appTables = []Table{Table{"UserTable",1}, Table{"UserCatTable",3}, Table{"ItemTable",1}, Table{"ItemCatTable",4}, Table{"ItemHistoryTable",4},Table{"TransTable",2}}
//Record types to store in tables
var recType = []string{"ARTINV", "USER", "BID", "AUCREQ", "POSTTRAN", "OPENAUC", "CLAUC", "XFER", "VERIFY"}
var globalKey = "2016"

///////////////////////// BASIC FUNCTIOS ///////////////////////////////////////////////////////////////
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {	
	fmt.Println("Init application")
	var err error
	for i := 0; i <len(appTables); i++ {
		err = stub.DeleteTable(appTables[i].Name)
		if err != nil {
			return nil, fmt.Errorf("Init(): DeleteTable of %s  Failed ", appTables[i].Name)
		}
		err = InitLedger(stub, appTables[i])
		if err != nil {
			return nil, fmt.Errorf("Init(): InitLedger of %s  Failed ", appTables[i].Name)
		}
	}	
	err = stub.PutState("version", []byte(strconv.Itoa(23)))
	if err != nil {
		return nil, err
	}
	fmt.Println("Init() Initialization Complete  : ", args)
	return []byte("Init(): Initialization Complete"), nil
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var err error
	var buff []byte
	if ChkReqType(args) == true {
		InvokeRequest := InvokeFunction(function)
		if InvokeRequest != nil {
			buff, err = InvokeRequest(stub, function, args)
		}
	} else {
		fmt.Println("Invoke() Invalid recType : ", args, "\n")
		return nil, errors.New("Invoke() : Invalid recType : " + args[0])
	}
	return buff, err
}

func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var err error
	var buff []byte
	fmt.Println("ID Extracted and Type = ", args[0])
	fmt.Println("Args supplied : ", args)
	if len(args) < 1 {
		fmt.Println("Query() : Include at least 1 arguments Key ")
		return nil, errors.New("Query() : Expecting Transation type and Key value for query")
	}
	QueryRequest := QueryFunction(function)
	if QueryRequest != nil {
		buff, err = QueryRequest(stub, function, args)
	} else {
		fmt.Println("Query() Invalid function call : ", function)
		return nil, errors.New("Query() : Invalid function call : " + function)
	}
	if err != nil {
		fmt.Println("Query() Object not found : ", args[0])
		return nil, errors.New("Query() : Object not found : " + args[0])
	}
	return buff, err
}
/////////////////////// END OF BASIC FUNCTIONS /////////////////////////////////////////////////////////////////////

/////////////////////////// GENERAL FUNCTIONS ///////////////////////////////////////////////////////////////////////
func InitLedger(stub *shim.ChaincodeStub, tableObject Table) error {
	nKeys := tableObject.Keys
	if nKeys < 1 {
		fmt.Println("At least 1 Key must be provided \n")
		fmt.Println("Failed creating Table ", tableObject.Name)
		return errors.New("Failed creating Table " + tableObject.Name)
	}
	var columnDefsForTbl []*shim.ColumnDefinition
	for i := 0; i < nKeys; i++ {
		columnDef := shim.ColumnDefinition{Name: "keyName" + strconv.Itoa(i), Type: shim.ColumnDefinition_STRING, Key: true}
		columnDefsForTbl = append(columnDefsForTbl, &columnDef)
	}
	columnLastTblDef := shim.ColumnDefinition{Name: "Details", Type: shim.ColumnDefinition_BYTES, Key: false}
	columnDefsForTbl = append(columnDefsForTbl, &columnLastTblDef)
	// Create the Table (Nil is returned if the Table exists or if the table is created successfully
	err := stub.CreateTable(tableObject.Name, columnDefsForTbl)
	if err != nil {
		fmt.Println("Failed creating Table ", tableObject.Name)
		return errors.New("Failed creating Table " + tableObject.Name)
	}
	return err
}


func InvokeFunction(fname string) func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	InvokeFunc := map[string]func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error){
		"CreateUser":			CreateUser,			
		"PostItem":           PostItem,
		/*"PostUser":           PostUser,
		"PostAuctionRequest": PostAuctionRequest,
		"PostTransaction":    PostTransaction,
		"PostBid":            PostBid,
		"OpenAuctionForBids": OpenAuctionForBids,
		"BuyItNow":           BuyItNow,
		"TransferItem":       TransferItem,
		"CloseAuction":       CloseAuction,
		"CloseOpenAuctions":  CloseOpenAuctions,*/
	}
	return InvokeFunc[fname]
}


func QueryFunction(fname string) func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	QueryFunc := map[string]func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error){		
		"GetUser":               GetUser,
		/*"GetItem":               GetItem,		
		"GetAuctionRequest":     GetAuctionRequest,
		"GetTransaction":        GetTransaction,
		"GetBid":                GetBid,
		"GetLastBid":            GetLastBid,
		"GetHighestBid":         GetHighestBid,
		"GetNoOfBidsReceived":   GetNoOfBidsReceived,
		"GetListOfBids":         GetListOfBids,
		"GetItemLog":            GetItemLog,
		"GetItemListByCat":      GetItemListByCat,
		"GetUserListByCat":      GetUserListByCat,
		"GetListOfInitAucs":     GetListOfInitAucs,
		"GetListOfOpenAucs":     GetListOfOpenAucs,
		"ValidateItemOwnership": ValidateItemOwnership,
		"IsItemOnAuction":       IsItemOnAuction,
		"GetVersion":            GetVersion,*/
	}
	return QueryFunc[fname]
}

func ChkReqType(args []string) bool {
	for _, rt := range args {
		for _, val := range recType {
			if val == rt {
				return true
			}
		}
	}
	return false
}

func GetNumberOfKeys(tname string) int {
	for i:=0; i<len(appTables); i++{
		if appTables[i].Name == tname{
			return appTables[i].Keys
		}
	}
	return 0
}

func UpdateLedger(stub *shim.ChaincodeStub, tableName string, keys []string, args []byte) error {
	nKeys := GetNumberOfKeys(tableName)
	if nKeys < 1 {
		fmt.Println("Atleast 1 Key must be provided \n")
	}
	var columns []*shim.Column
	for i := 0; i < nKeys; i++ {
		col := shim.Column{Value: &shim.Column_String_{String_: keys[i]}}
		columns = append(columns, &col)
	}
	lastCol := shim.Column{Value: &shim.Column_Bytes{Bytes: []byte(args)}}
	columns = append(columns, &lastCol)
	row := shim.Row{columns}
	ok, err := stub.InsertRow(tableName, row)
	if err != nil {
		return fmt.Errorf("UpdateLedger: InsertRow into "+tableName+" Table operation failed. %s", err)
	}
	if !ok {
		return errors.New("UpdateLedger: InsertRow into " + tableName + " Table failed. Row with given key " + keys[0] + " already exists")
	}

	fmt.Println("UpdateLedger: InsertRow into ", tableName, " Table operation Successful. ")
	return nil
}

func QueryLedger(stub *shim.ChaincodeStub, tableName string, args []string) ([]byte, error) {
	var columns []shim.Column
	nCol := GetNumberOfKeys(tableName)
	for i := 0; i < nCol; i++ {
		colNext := shim.Column{Value: &shim.Column_String_{String_: args[i]}}
		columns = append(columns, colNext)
	}
	row, err := stub.GetRow(tableName, columns)
	fmt.Println("Length or number of rows retrieved ", len(row.Columns))
	if len(row.Columns) == 0 {
		jsonResp := "{\"Error\":\"Failed retrieving data " + args[0] + ". \"}"
		fmt.Println("Error retrieving data record for Key = ", args[0], "Error : ", jsonResp)
		return nil, errors.New(jsonResp)
	}	
	Avalbytes := row.Columns[nCol].GetBytes()	
	fmt.Println("QueryLedger() : Successful - Proceeding to ProcessRequestType ")
	err = ProcessQueryResult(stub, Avalbytes, args)
	if err != nil {
		fmt.Println("QueryLedger() : Cannot create object  : ", args[1])
		jsonResp := "{\"QueryLedger() Error\":\" Cannot create Object for key " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}
	return Avalbytes, nil
}

func ProcessQueryResult(stub *shim.ChaincodeStub, Avalbytes []byte, args []string) error {	
	var dat map[string]interface{}
	if err := json.Unmarshal(Avalbytes, &dat); err != nil {
		panic(err)
	}
	var recType string
	recType = dat["RecType"].(string)
	switch recType {	
	case "USER":
		ur, err := JSONtoUser(Avalbytes)
		if err != nil {
			return err
		}
		fmt.Println("ProcessRequestType() : ", ur)
		return err
	default:
		return errors.New("Unknown")
	}
	return nil
}

func ReplaceLedgerEntry(stub *shim.ChaincodeStub, tableName string, keys []string, args []byte) error {

	nKeys := GetNumberOfKeys(tableName)
	if nKeys < 1 {
		fmt.Println("Atleast 1 Key must be provided \n")
	}
	var columns []*shim.Column
	for i := 0; i < nKeys; i++ {
		col := shim.Column{Value: &shim.Column_String_{String_: keys[i]}}
		columns = append(columns, &col)
	}
	lastCol := shim.Column{Value: &shim.Column_Bytes{Bytes: []byte(args)}}
	columns = append(columns, &lastCol)
	row := shim.Row{columns}
	ok, err := stub.ReplaceRow(tableName, row)
	if err != nil {
		return fmt.Errorf("ReplaceLedgerEntry: Replace Row into "+tableName+" Table operation failed. %s", err)
	}
	if !ok {
		return errors.New("ReplaceLedgerEntry: Replace Row into " + tableName + " Table failed. Row with given key " + keys[0] + " already exists")
	}

	fmt.Println("ReplaceLedgerEntry: Replace Row in ", tableName, " Table operation Successful. ")
	return nil
}

func GetList(stub *shim.ChaincodeStub, tableName string, args []string) ([]shim.Row, error) {
	var columns []shim.Column
	nKeys := GetNumberOfKeys(tableName)
	nCol := len(args)
	if nCol < 1 {
		fmt.Println("Atleast 1 Key must be provided \n")
		return nil, errors.New("GetList failed. Must include at least key values")
	}
	for i := 0; i < nCol; i++ {
		colNext := shim.Column{Value: &shim.Column_String_{String_: args[i]}}
		columns = append(columns, colNext)
	}
	rowChannel, err := stub.GetRows(tableName, columns)
	if err != nil {
		return nil, fmt.Errorf("GetList operation failed. %s", err)
	}
	var rows []shim.Row
	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				rows = append(rows, row)
				//If required enable for debugging
				//fmt.Println(row)
			}
		}
		if rowChannel == nil {
			break
		}
	}

	fmt.Println("Number of Keys retrieved : ", nKeys)
	fmt.Println("Number of rows retrieved : ", len(rows))
	return rows, nil
}

/////////////////////////// END OF GENERAL FUNCTIONS ////////////////////////////////////////////////////////////////


/////////////////////// USER FUNCTIONS ////////////////////////////////////////////////////////////////////////////
func CreateUser(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	record, err := CreateUserObject(args[0:])
	if err != nil {
		return nil, err
	}
	buff, err := UsertoJSON(record)
	if err != nil {
		fmt.Println("PostuserObject() : Failed Cannot create object buffer for write : ", args[1])
		return nil, errors.New("PostUser(): Failed Cannot create object buffer for write : " + args[1])
	} else {
		//Taken UserID as the key
		keys := []string{args[0]}
		err = UpdateLedger(stub, "UserTable", keys, buff)
		if err != nil {
			fmt.Println("PostUser() : write error while inserting record")
			return nil, err
		}
		//Saving user by categorie (Using UserType and UserId)
		keys = []string{globalKey, args[4], args[0]}
		err = UpdateLedger(stub, "UserCatTable", keys, buff)
		if err != nil {
			fmt.Println("PostUser() : write error while inserting recordinto UserCatTable \n")
			return nil, err
		}
	}
	return buff, err
}

func CreateUserObject(args []string) (UserObject, error) {	
	var aUser UserObject	
	if len(args) != 6 {
		fmt.Println("CreateUserObject(): Incorrect number of arguments. Expecting 6 ")
		return aUser, errors.New("CreateUserObject() : Incorrect number of arguments. Expecting 6 ")
	}	
	aUser = UserObject{args[0], args[1], args[2], args[3], args[4], args[5]}
	fmt.Println("CreateUserObject() : User Object : ", aUser)
	return aUser, nil
}

func UsertoJSON(user UserObject) ([]byte, error) {
	ajson, err := json.Marshal(user)
	if err != nil {
		fmt.Println("UsertoJSON error: ", err)
		return nil, err
	}
	fmt.Println("UsertoJSON created: ", ajson)
	return ajson, nil
}

func JSONtoUser(user []byte) (UserObject, error) {
	ur := UserObject{}
	err := json.Unmarshal(user, &ur)
	if err != nil {
		fmt.Println("JSONtoUser error: ", err)
		return ur, err
	}
	fmt.Println("JSONtoUser created: ", ur)
	return ur, err
}

func GetUser(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var err error	
	Avalbytes, err := QueryLedger(stub, "UserTable", args)
	if err != nil {
		fmt.Println("GetUser() : Failed to Query Object ")
		jsonResp := "{\"Error\":\"Failed to get  Object Data for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}
	if Avalbytes == nil {
		fmt.Println("GetUser() : Incomplete Query Object ")
		jsonResp := "{\"Error\":\"Incomplete information about the key for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}

	fmt.Println("GetUser() : Response : Successfull -")
	return Avalbytes, nil
}
/////////////////////////////////// END OF USER FUNCTIONS ////////////////////////////////////////////////////////



	
/////////////////////////////////////// ITEMS FUNCTIONS //////////////////////////////////////////////////////
func PostItem(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	itemObject, err := CreateItemObject(args[0:])
	if err != nil {
		fmt.Println("PostItem(): Cannot create item object \n")
		return nil, err
	}	
	// Convert Item Object to JSON
	buff, err := ItemtoJSON(itemObject) //
	if err != nil {
		fmt.Println("PostItem() : Failed Cannot create object buffer for write : ", args[1])
		return nil, errors.New("PostItem(): Failed Cannot create object buffer for write : " + args[1])
	} else {		
		keys := []string{args[0]}
		err = UpdateLedger(stub, "ItemTable", keys, buff)
		if err != nil {
			fmt.Println("PostItem() : write error while inserting record\n")
			return buff, err
		}		
		keys = []string{globalKey, args[5], args[0], args[9]}
		err = UpdateLedger(stub, "ItemCatTable", keys, buff)
		if err != nil {
			fmt.Println("PostItem() : Write error while inserting record into ItemCatTable \n")
			return buff, err
		}
	}	
	return []byte("Item posted successfully"), nil
}

func CreateItemObject(args []string) (ItemObject, error) {
	var err error
	var myItem ItemObject	
	if len(args) != 10 {
		fmt.Println("CreateItemObject(): Incorrect number of arguments. Expecting 10 ")
		return myItem, errors.New("CreateItemObject(): Incorrect number of arguments. Expecting 10 ")
	}
	// Validate ItemID is an integer
	_, err = strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("CreateItemObject():ID should be an integer create failed! ")
		return myItem, errors.New("CreateItemObject():ID should be an integer create failed!")
	}
	// Append the AES Key, The Encrypted Image Byte Array and the file type
	myItem = ItemObject{args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],args[0]}	
	return myItem, nil
}

func ItemtoJSON(item ItemObject) ([]byte, error) {
	ijson, err := json.Marshal(item)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return ijson, nil
}


func JSONtoItem(data []byte) (ItemObject, error) {
	ar := ItemObject{}
	err := json.Unmarshal([]byte(data), &ar)
	if err != nil {
		fmt.Println("Unmarshal failed : ", err)
	}
	return ar, err
}

func GetItemListByCat(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {	
	if len(args) < 1 {
		fmt.Println("GetItemListByCat(): Incorrect number of arguments. Expecting 1 ")		
		return nil, errors.New("CreateItemObject(): Incorrect number of arguments. Expecting 1 ")
	}
	rows, err := GetList(stub, "ItemCatTable", args)
	if err != nil {
		return nil, fmt.Errorf("GetItemListByCat() operation failed. Error GetList: %s", err)
	}
	nCol := GetNumberOfKeys("ItemCatTable")
	tlist := make([]ItemObject, len(rows))
	for i := 0; i < len(rows); i++ {
		ts := rows[i].Columns[nCol].GetBytes()
		io, err := JSONtoItem(ts)
		if err != nil {
			fmt.Println("() Failed : Ummarshall error")
			return nil, fmt.Errorf("GetItemListByCat() operation failed. %s", err)
		}		
		tlist[i] = io
	}
	jsonRows, _ := json.Marshal(tlist)
	//fmt.Println("All Items : ", jsonRows)
	return jsonRows, nil
}

