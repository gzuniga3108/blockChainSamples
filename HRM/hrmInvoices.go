package main
import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

///////////////////////////// GLOBAL VARIABLES ///////////////////////////////////
var globalInvoiceKey = "HRMINV"
var appTables = []string{"InvoiceTable"} 
var recType = []string{"INVOICE"}

///////////////////////////// OBJECTS STRUCTURES /////////////////////////////////
type SimpleChaincode struct {
}

type InvoiceObject struct {
	InvoiceID 	string //PRIMARYKEY	
	Issuer    	string 
	Receptor  	string 
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
		"CreateInvoice":	CreateInvoice,
	}
	return InvokeFunc[fname]
}

func QueryFunction(fname string) func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	QueryFunc := map[string]func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error){			
		"GetAllInvoices":				GetAllInvoices,
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

func GetNumberOfKeys(tname string) int {
	switch tname{		
		case "InvoiceTable":
			return 2	//InvoiceId
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
	keys := []string{globalInvoiceKey,args[0]}
	err = UpdateLedger(stub,"InvoiceTable",keys,invoiceBytes)
	if err != nil{
		return nil,errors.New("Error: Cannot save invoice")
	}
	return []byte("Item created successfully"),nil
}

func GetAllInvoices(stub shim.ChaincodeStubInterface,function string,args []string)([]byte,error){
	if len(args) >= 1{
		args[1] = args[0]
		args[0] = globalInvoiceKey
	}
	if len(args) < 1 {
		args[0] = globalInvoiceKey
	}
	rows, err := GetList(stub, "InvoiceTable", args)
	if err != nil {
		return nil, fmt.Errorf("GetAllInvoices() operation failed. Error marshaling JSON: %s", err)
	}
	nCol := GetNumberOfKeys("InvoiceTable")
	tlist := make([]InvoiceObject, len(rows))
	for i := 0; i < len(rows); i++ {
		ts := rows[i].Columns[nCol].GetBytes()
		uo, err := JsontoInvoice(ts)
		if err != nil {
			fmt.Println("GetAllInvoices() Failed : Ummarshall error")
			return nil, fmt.Errorf("GetAllInvoices() operation failed. %s", err)
		}
		tlist[i] = uo
	}
	jsonRows, _ := json.Marshal(tlist)	
	return jsonRows, nil
}

func InvoiceToJson(oInvoice InvoiceObject)([]byte,error){
	invoiceBytes,err := json.Marshal(oInvoice)
	if err != nil{
		return nil,errors.New("Error:Cannot get invoice  bytes")
	}
	return invoiceBytes,nil
}

func JsontoInvoice(invoiceBytes []byte)(InvoiceObject,error){
	oInvoice := InvoiceObject{}
	err := json.Unmarshal(invoiceBytes,&oInvoice)
	if err != nil{
		return oInvoice,errors.New("Error: Cannot create invoice object")
	}
	return oInvoice,nil
}
