package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	/*"time"
	"strings"*/
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

///////////////////////////// GLOBAL VARIABLES ///////////////////////////////////
var globalKey = "2016"
var appTables = []string{"UserTable","ItemTable"} 
var recType = []string{"USER"}


///////////////////////////// OBJECTS STRUCTURES /////////////////////////////////
type SimpleChaincode struct {
}

type UserObject struct{
	UserId			string
	Name 			string
	Password		string
	CashBalance		string
	UserType		string
	RecType			string //USER
}

type  ItemObject struct{
	ItemId		   	string
	Description	   	string
	Subject		   	string
	Price		   	string	
	OwnerId		   	string
	Status			string
	Image			string
	ItemType	   	string
	RecType 		string //ITEM
}

//////////////////////// BASIC FUNCTIONS ////////////////////////////////////////
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
		err = stub.DeleteTable(appTables[i])
		if err != nil {
			return nil, fmt.Errorf("Init(): DeleteTable of %s  Failed ", appTables[i])
		}
		err = InitLedger(stub, appTables[i])
		if err != nil {
			return nil, fmt.Errorf("Init(): InitLedger of %s  Failed ", appTables[i])
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

///////////////////// GENERAL FUNCTIONS //////////////////////////////////////////////////
func InitLedger(stub *shim.ChaincodeStub, tableName string) error {
	nKeys := GetNumberOfKeys(tableName)
	if nKeys < 1 {
		fmt.Println("At least one key required")		
		return errors.New("Failed creating Table " + tableName)
	}
	var columnDefsForTbl []*shim.ColumnDefinition
	for i := 0; i < nKeys; i++ {
		columnDef := shim.ColumnDefinition{Name: "keyName" + strconv.Itoa(i), Type: shim.ColumnDefinition_STRING, Key: true}
		columnDefsForTbl = append(columnDefsForTbl, &columnDef)
	}
	columnLastTblDef := shim.ColumnDefinition{Name: "Details", Type: shim.ColumnDefinition_BYTES, Key: false}
	columnDefsForTbl = append(columnDefsForTbl, &columnLastTblDef)	
	err := stub.CreateTable(tableName, columnDefsForTbl)
	if err != nil {
		fmt.Println("Failed creating Table ", tableName)
		return errors.New("Failed creating Table " + tableName)
	}
	return err
}

func InvokeFunction(fname string) func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	InvokeFunc := map[string]func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error){		
		"CreateUser":	CreateUser,
	}
	return InvokeFunc[fname]
}

func QueryFunction(fname string) func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	QueryFunc := map[string]func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error){			
		"GetUser":	GetUser,
	}
	return QueryFunc[fname]
}

func UpdateLedger(stub *shim.ChaincodeStub, tableName string, keys []string, args []byte) error {
	nKeys := GetNumberOfKeys(tableName)
	if nKeys < 1 {
		return errors.New("At least one key is required")
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

func GetNumberOfKeys(tname string) int {
	switch tname{
		case "UserTable":
			return 1 //UserId
		case "UserCatTable":
			return 3 //GlobalKey,UserType,UserId
		case "ItemTable":
			return 1
	}
	return 0
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
	case "ITEM":
		oItem,err := JsontoItem(Avalbytes)
		if err != nil{
			return err
		}
		fmt.Println("ProcessRequestType() : ", oItem)
		return err
	default:
		return errors.New("Unknown")
	}
	return nil
}


///////////////////////// USER'S FUNCTIONS ////////////////////////////////////////////////////////////////////////////////

func CreateUser(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var aUser UserObject
	if len(args) < 6{
		return nil,errors.New("Expecting 6 parameters")
	}
	aUser = UserObject{args[0], args[1], args[2], args[3], args[4], args[5]}
	userBytes,err := UsertoJSON(aUser)
	if err != nil{
		return nil,errors.New("Error creating userr bytes")
	}
	keys := []string{args[0]}
	err = UpdateLedger(stub,"UserTable",keys,userBytes)
	if err != nil{
		return nil,errors.New("Error: An error has ocured while creating the user")
	}
	return []byte("User created successfully"),nil
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

func UsertoJSON(user UserObject) ([]byte, error) {
	ajson, err := json.Marshal(user)
	if err != nil {
		fmt.Println("UsertoJSON error: ", err)
		return nil, err
	}
	fmt.Println("UsertoJSON created: ", ajson)
	return ajson, nil
}


//////////////////////////////////////////// ITEMS FUNCTION /////////////////////////////////////////////////////////
func CreateItem(stub *shim.ChaincodeStub, function string, args []string)([]byte,error){
	var oItem ItemObject
	if len(args) < 9{
		return nil,errors.New("Error: Expecting 9 parameters")
	}
	oItem = ItemObject{args[0],args[1],args[2],args[3],args[4],args[5],args[6],args[7],args[8]}
	itemBytes,err := ItemToJson(oItem) 
	if err != nil{
		return nil,err
	}
	keys := []string{args[0]}
	err = UpdateLedger(stub,"ItemTable",keys,itemBytes)
	if err != nil{
		return nil,errors.New("Error: Cannot save item")
	}
	return []byte("Item created successfully"),nil
}

func GetItem(stub *shim.ChaincodeStub,function string, args []string)([]byte,error){
	var err error
	Avalbytes,err := QueryLedger(stub,"ItemTable",args)
	if err != nil{
		return nil,errors.New("{\"Error\":\"Cannot retrieve item information\"}")
	}
	if Avalbytes == nil{
		return nil,errors.New("{\"Error\":\"Item information is incomplete\"}")
	}
	return Avalbytes,nil
}

func JsontoItem(itemBytes []byte)(ItemObject,error){
	oItem := ItemObject{}
	err := json.Unmarshal(itemBytes,&oItem)
	if err != nil{
		return oItem,errors.New("Error: Cannot create item object")
	}
	return oItem,nil
}

func ItemToJson(oItem ItemObject)([]byte,error){
	itemBytes,err := json.Marshal(oItem)
	if err != nil{
		return nil,errors.New("Error:Cannot get item  bytes")
	}
	return itemBytes,nil
}
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

