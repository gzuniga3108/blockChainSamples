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
var appTables = []string{"UserTable","ItemTable","TransactionTable","UserDetailTable","ItemCatTable","ItemOwnerTable","TransPrevOwnerTable","TransNewOwnerTable","InvoiceTable"} 
var recType = []string{"USER","ITEM","TRANS","INVOICE"}


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

type TransactionObject struct{
	TransactionId 	string
	ItemId 			string
	PrevOwner		string
	NewOwner		string
	PaidQty			string
	TransType		string
	RecType 		string //TRANS
}

type InvoiceObject struct {
	InvoiceID 	string //PRIMARYKEY	
	Issuer    	string //KEY
	Receptor  	string //KEY
	Amount	  	string
	Xml		  	string
	PaymentDay  string
	Status      string
	RecType   	string //INVOICE	
}
//////////////////////// BASIC FUNCTIONS ////////////////////////////////////////
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {	
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

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
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

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
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
func InitLedger(stub shim.ChaincodeStubInterface, tableName string) error {
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

func InvokeFunction(fname string) func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	InvokeFunc := map[string]func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error){		
		"CreateUser":		CreateUser,
		"CreateItem":		CreateItem,
		"UpdateUser":		UpdateUser,
		"UpdateItem":		UpdateItem,
		"NewTransaction": 	NewTransaction,
		"CreateInvoice":	CreateInvoice,

	}
	return InvokeFunc[fname]
}

func QueryFunction(fname string) func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	QueryFunc := map[string]func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error){			
		"GetUser":					GetUser,
		"GetItem":					GetItem,
		"GetTransaction":			GetTransaction,
		"GetUserList":				GetUserList,
		"GetItemListByCat": 		GetItemListByCat,
		"GetItemListByOwner":		GetItemListByOwner,
		"GetTransListPrevOwner":	GetTransListPrevOwner,
		"GetTransListNewOwner":		GetTransListNewOwner,
		"GetInvoice":				GetInvoice,
	}
	return QueryFunc[fname]
}

func UpdateLedger(stub shim.ChaincodeStubInterface, tableName string, keys []string, args []byte) error {
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

func ReplaceLedgerEntry(stub shim.ChaincodeStubInterface, tableName string, keys []string, args []byte) error {
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

func GetNumberOfKeys(tname string) int {
	switch tname{
		case "UserTable":
			return 1 //UserId
		case "UserDetailTable":
			return 3 //GlobalKey,UserType,UserId
		case "ItemTable": 
			return 1 //ItemId
		case "ItemCatTable":
			return 4 //GlobalKey,Status,ItemType,ItemId
		case "ItemOwnerTable":
			return 4 //GlobalKey,Owner,Status,ItemId
		case "TransactionTable": 
			return 1 //TransactionID			
		case "TransPrevOwnerTable":
			return 4
		case "TransNewOwnerTable":
			return 4			
		case "InvoiceTable":
			return 4			
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

func QueryLedger(stub shim.ChaincodeStubInterface, tableName string, args []string) ([]byte, error) {
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

func ProcessQueryResult(stub shim.ChaincodeStubInterface, Avalbytes []byte, args []string) error {	
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
	case "TRANS":
		oTransaction,err := JsonToTransaction(Avalbytes)
		if err != nil{
			return err
		}
		fmt.Println("ProcessRequestType() : ", oTransaction)
		return err
	case "INVOICE":
		oInvoice,err := JsonToInvoice(Avalbytes)
		if err != nil{
			return err
		}
		fmt.Println("ProcessRequestType() : ", oInvoice)
		return err	
	default:
		return errors.New("Unknown")
	}
	return nil
}

func GetList(stub shim.ChaincodeStubInterface, tableName string, args []string) ([]shim.Row, error) {
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

///////////////////////// USER'S FUNCTIONS ////////////////////////////////////////////////////////////////////////////////

func CreateUser(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
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
	//Saving user details table(
	keys = []string{globalKey,args[4],args[0]}
	err = UpdateLedger(stub,"UserDetailTable",keys,userBytes)
	if err != nil{
		return nil,err
	}
	return []byte("User created successfully"),nil
}
func GetUser(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
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

func GetUserList(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) < 1 {
		fmt.Println("GetUserList(): Incorrect number of arguments. Expecting 1 ")		
		return nil, errors.New("CreateUserObject(): Incorrect number of arguments. Expecting 1 ")
	}
	rows, err := GetList(stub, "UserDetailTable", args)
	if err != nil {
		return nil, fmt.Errorf("GetUserList() operation failed. Error marshaling JSON: %s", err)
	}
	nCol := GetNumberOfKeys("UserDetailTable")
	tlist := make([]UserObject, len(rows))
	for i := 0; i < len(rows); i++ {
		ts := rows[i].Columns[nCol].GetBytes()
		uo, err := JSONtoUser(ts)
		if err != nil {
			fmt.Println("GetUserList() Failed : Ummarshall error")
			return nil, fmt.Errorf("GetUserList() operation failed. %s", err)
		}
		tlist[i] = uo
	}
	jsonRows, _ := json.Marshal(tlist)	
	return jsonRows, nil
}

func  UpdateUser(stub shim.ChaincodeStubInterface,function string,args []string)([]byte,error){
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
	err = ReplaceLedgerEntry(stub,"UserTable",keys,userBytes)
	if err != nil{
		return nil,errors.New("Error: An error has ocured while creating the user")
	}
	//Saving user details table(
	keys = []string{globalKey,args[4],args[0]}
	err = UpdateLedger(stub,"UserDetailTable",keys,userBytes)
	if err != nil{
		return nil,err
	}
	return []byte("User updated successfully"),nil	
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


///////////////////////////////////////////// INVOICE'S FUNCTIONS ////////////////////////////////////////////////////
func CreateInvoice(stub shim.ChaincodeStubInterface, function string, args []string)([]byte,error){
	var oInvoice InvoiceObject
	if len(args) < 7{
		return nil,errors.New("Error: Expecting 7 parameters")
	}
	oInvoice = InvoiceObject{args[0],args[1],args[2],args[3],args[4],"--",args[5],args[6]}
	invoiceBytes,err := InvoiceToJson(oInvoice) 
	if err != nil{
		return nil,err
	}
	keys := []string{globalKey,args[0],args[1],args[2]}
	err = UpdateLedger(stub,"InvoiceTable",keys,invoiceBytes)
	if err != nil{
		return nil,errors.New("Error: Cannot save invoice")
	}
	return []byte("Item created successfully"),nil
}

func GetInvoice(stub shim.ChaincodeStubInterface,function string, args []string)([]byte,error){
	var err error
	Avalbytes,err := QueryLedger(stub,"InvoiceTable",args)
	if err != nil{
		return nil,errors.New("{\"Error\":\"Cannot retrieve invoice information\"}")
	}
	if Avalbytes == nil{
		return nil,errors.New("{\"Error\":\"Invoice information is incomplete\"}")
	}
	return Avalbytes,nil
}

func InvoiceToJson(oInvoice InvoiceObject)([]byte,error){
	invoiceBytes,err := json.Marshal(oInvoice)
	if err != nil{
		return nil,errors.New("Error:Cannot get invoice  bytes")
	}
	return invoiceBytes,nil
}

func JsonToInvoice(invoiceBytes []byte)(InvoiceObject,error){
	oInvoice := InvoiceObject{}
	err := json.Unmarshal(invoiceBytes,&oInvoice)
	if err != nil{
		return oInvoice,errors.New("Error: Cannot create item object")
	}
	return oInvoice,nil
}
//////////////////////////////////////////// ITEM'S FUNCTION /////////////////////////////////////////////////////////
func CreateItem(stub shim.ChaincodeStubInterface, function string, args []string)([]byte,error){
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
	//Insert into ItemCatTable
	//GlobalKey,Status,ItemType,ItemId
	keys = []string{globalKey,args[5],args[7],args[0]}
	err = UpdateLedger(stub,"ItemCatTable",keys,itemBytes)
	if err != nil{
		return nil,err
	}
	//Insert into ItemOwnerTable
	keys = []string{globalKey,args[4],args[5],args[0]}
	err = UpdateLedger(stub,"ItemOwnerTable",keys,itemBytes)
	if err != nil{
		return nil,err
	}
	return []byte("Item created successfully"),nil
}

func GetItem(stub shim.ChaincodeStubInterface,function string, args []string)([]byte,error){
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

func UpdateItem(stub shim.ChaincodeStubInterface, function string, args []string)([]byte,error){
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
	err = ReplaceLedgerEntry(stub,"ItemTable",keys,itemBytes)
	if err != nil{
		return nil,errors.New("Error: Cannot save item")
	}
	//Insert into ItemCatTable
	//GlobalKey,Status,ItemType,ItemId
	keys = []string{globalKey,args[5],args[7],args[0]}
	err = UpdateLedger(stub,"ItemCatTable",keys,itemBytes)
	if err != nil{
		return nil,err
	}
	//Insert into ItemOwnerTable
	keys = []string{globalKey,args[4],args[5],args[0]}
	err = UpdateLedger(stub,"ItemOwnerTable",keys,itemBytes)
	if err != nil{
		return nil,err
	}
	return []byte("Item created successfully"),nil
}

func GetItemListByCat(stub shim.ChaincodeStubInterface,function string,args []string)([]byte,error){
	if len(args) < 1 {
		fmt.Println("GetItemListByCat(): Incorrect number of arguments. Expecting 1 ")		
		return nil, errors.New("GetItemListByCat(): Incorrect number of arguments. Expecting 1 ")
	}
	rows, err := GetList(stub, "ItemCatTable", args)
	if err != nil {
		return nil, fmt.Errorf("GetItemListByCat() operation failed. Error marshaling JSON: %s", err)
	}
	nCol := GetNumberOfKeys("ItemCatTable")
	tlist := make([]ItemObject, len(rows))
	for i := 0; i < len(rows); i++ {
		ts := rows[i].Columns[nCol].GetBytes()
		uo, err := JsontoItem(ts)
		if err != nil {
			fmt.Println("GetItemListByCat() Failed : Ummarshall error")
			return nil, fmt.Errorf("GetItemListByCat() operation failed. %s", err)
		}
		tlist[i] = uo
	}
	jsonRows, _ := json.Marshal(tlist)	
	return jsonRows, nil
}

func GetItemListByOwner(stub shim.ChaincodeStubInterface,function string,args []string)([]byte,error){
	if len(args) < 1 {
		fmt.Println("GetItemListByOwner(): Incorrect number of arguments. Expecting 1 ")		
		return nil, errors.New("GetItemListByOwner(): Incorrect number of arguments. Expecting 1 ")
	}
	rows, err := GetList(stub, "ItemOwnerTable", args)
	if err != nil {
		return nil, fmt.Errorf("GetItemListByOwner() operation failed. Error marshaling JSON: %s", err)
	}
	nCol := GetNumberOfKeys("ItemOwnerTable")
	tlist := make([]ItemObject, len(rows))
	for i := 0; i < len(rows); i++ {
		ts := rows[i].Columns[nCol].GetBytes()
		uo, err := JsontoItem(ts)
		if err != nil {
			fmt.Println("GetItemListByOwner() Failed : Ummarshall error")
			return nil, fmt.Errorf("GetItemListByOwner() operation failed. %s", err)
		}
		tlist[i] = uo
	}
	jsonRows, _ := json.Marshal(tlist)	
	return jsonRows, nil
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
////////////////////////// TRANSACTION'S FUNCTION //////////////////////////////////////////////////////////////
func NewTransaction (stub shim.ChaincodeStubInterface,function string, args []string)([]byte,error){
	var oTransaction TransactionObject
	if len(args) != 7{
		return nil,errors.New("Error: Expecting 6 arguments")
	}
	oTransaction = TransactionObject{args[0],args[1],args[2],args[3],args[4],args[5],args[6]}
	transactionBytes,err := TransactionToJson(oTransaction)
	if err != nil{
		return nil,errors.New("Error: Cannot get transaction bytes")
	}
	 keys := []string{args[0]}
	 err = UpdateLedger(stub,"TransactionTable",keys,transactionBytes)
	 if err != nil{
	 	return nil,errors.New("Error: An error has ocured while saving the transaction")
	 }
	 //Save record for prev owner
	 keys = []string{globalKey,args[2],args[5],args[0]}
	 err = UpdateLedger(stub,"TransPrevOwnerTable",keys,transactionBytes)
	 if err != nil{
	 	return nil,err
	 }
	 //Save record for new owner
	 keys = []string{globalKey,args[3],args[5],args[0]}
	 err = UpdateLedger(stub,"TransNewOwnerTable",keys,transactionBytes)
	 if err != nil{
	 	return nil,err
	 }
	 return []byte("Transation has been successfully saved"),nil
}

func GetTransaction(stub shim.ChaincodeStubInterface,function string, args []string)([]byte,error) {
	var err error
	Avalbytes,err := QueryLedger(stub,"TransactionTable",args)
	if err != nil{
		return nil,errors.New("{\"Error\":\"Cannot retrieve transaction information\"}")
	}
	if Avalbytes == nil{
		return nil,errors.New("{\"Error\":\"Transaction information is incomplete\"}")
	}
	return Avalbytes,nil
}

func GetTransListNewOwner(stub shim.ChaincodeStubInterface,function string, args []string)([]byte,error){
	if len(args) < 1 {
		fmt.Println("GetTransListNewOwner(): Incorrect number of arguments. Expecting 1 ")		
		return nil, errors.New("GetTransListNewOwner(): Incorrect number of arguments. Expecting 1 ")
	}
	rows, err := GetList(stub, "TransNewOwnerTable", args)
	if err != nil {
		return nil, fmt.Errorf("GetTransListNewOwner() operation failed. Error marshaling JSON: %s", err)
	}
	nCol := GetNumberOfKeys("TransNewOwnerTable")
	tlist := make([]TransactionObject, len(rows))
	for i := 0; i < len(rows); i++ {
		ts := rows[i].Columns[nCol].GetBytes()
		uo, err := JsonToTransaction(ts)
		if err != nil {
			fmt.Println("GetTransListNewOwner() Failed : Ummarshall error")
			return nil, fmt.Errorf("GetTransListNewOwner() operation failed. %s", err)
		}
		tlist[i] = uo
	}
	jsonRows, _ := json.Marshal(tlist)	
	return jsonRows, nil
}


func GetTransListPrevOwner(stub shim.ChaincodeStubInterface,function string, args []string)([]byte,error){
	if len(args) < 1 {
		fmt.Println("GetTransListPrevOwner(): Incorrect number of arguments. Expecting 1 ")		
		return nil, errors.New("GetTransListPrevOwner(): Incorrect number of arguments. Expecting 1 ")
	}
	rows, err := GetList(stub, "TransPrevOwnerTable", args)
	if err != nil {
		return nil, fmt.Errorf("GetTransListPrevOwner() operation failed. Error marshaling JSON: %s", err)
	}
	nCol := GetNumberOfKeys("TransPrevOwnerTable")
	tlist := make([]TransactionObject, len(rows))
	for i := 0; i < len(rows); i++ {
		ts := rows[i].Columns[nCol].GetBytes()
		uo, err := JsonToTransaction(ts)
		if err != nil {
			fmt.Println("GetTransListPrevOwner() Failed : Ummarshall error")
			return nil, fmt.Errorf("GetTransListPrevOwner() operation failed. %s", err)
		}
		tlist[i] = uo
	}
	jsonRows, _ := json.Marshal(tlist)	
	return jsonRows, nil
}

func JsonToTransaction(transactionBytes []byte)(TransactionObject,error){	
	oTransaction := TransactionObject{}
	err := json.Unmarshal(transactionBytes,&oTransaction)
	if err != nil{
		return oTransaction,errors.New("Error: Error creating transaction object")
	}
	return oTransaction,nil
}


func TransactionToJson(oTransaction TransactionObject) ([]byte,error){
	transactionBytes,err := json.Marshal(oTransaction)
	if err != nil{
		return nil,errors.New("Error: Cannot get transaction bytes")
	}
	return transactionBytes,nil
}