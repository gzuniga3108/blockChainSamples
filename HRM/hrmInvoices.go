///////////////////////////////////////////////////////////////////////
// Author : Genaro Zuniga
// Date   : 20/12/2016	
///////////////////////////////////////////////////////////////////////
package main

import (	
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"	
	//"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io"
)

/////////////////////////// OBJECTS STRUCTURES ////////////////////////////////////////////////////////
type SimpleChaincode struct {
}
type Table struct{
	Name string
	Keys int
}
type InvoiceObject struct {
	InvoiceID 	string //PRIMARYKEY
	Amount	  	string	
	Issuer    	string
	Receptor  	string
	Xml		  	string
	PaymentDay  string
	AES_Key     []byte
	OwnerComp 	string	
	Status      string
	RecType   	string //INVOICE	
	GlobalKey   string
}

///////////////////////// GLOBAL VARIABLES ////////////////////////////////////////////////////////////
//Tables that will be used in the application
var appTables = []Table{Table{"InvoiceTable",1}}
//Record types to store in tables
var recType = []string{"INVOICE"}
//Global key to get all records
var globalkey = "HRM";

const (
	AESKeyLength = 32 // AESKeyLength is the default AES key length
	NonceSize    = 24 // NonceSize is the default NonceSize
)
///////////////////////// BASIC FUNCTIOS ///////////////////////////////////////////////////////////////
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
func InitLedger(stub shim.ChaincodeStubInterface, tableObject Table) error {
	nKeys := tableObject.Keys
	if nKeys < 1 {
		fmt.Println("At least 1 Key must be provided \n")
		fmt.Println("Failed c|reating Table ", tableObject.Name)
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


func InvokeFunction(fname string) func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	InvokeFunc := map[string]func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error){
		"PostInvoice": PostInvoice,
	}
	return InvokeFunc[fname]
}


func QueryFunction(fname string) func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	QueryFunc := map[string]func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error){
		//Eg
		//"GetItem":               GetItem,
		//"GetUser":               GetUser,
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

func UpdateLedger(stub shim.ChaincodeStubInterface, tableName string, keys []string, args []byte) error {
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

func GetNumberOfKeys(tname string) int {
	TableMap := map[string]int{
		"InvoiceTable": 1,
	}
	return TableMap[tname]
}

/////////////////////////// END OF GENERAL FUNCTIONS ////////////////////////////////////////////////////////////////


////////////////////////////////// CUSTOM FUNCTIONS /////////////////////////////////////////////////////////////////
func PostInvoice(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	invoiceObject, err := CreateInvoiceObject(args[0:])
	if err != nil {
		fmt.Println("PostInvoice(): Cannot create invoice object \n")
		return nil, err
	}
	
	// Convert Invoice Object to JSON
	buff, err := ARtoJSON(invoiceObject) //
	if err != nil {
		fmt.Println("{\"status\":\"error\",\"message\":\"Cannot create invoice object buffer\"}")
		return nil, errors.New("{\"status\":\"error\",\"message\":\"Cannot create invoice object buffer\"}")
	} else {
		// Update the ledger with the Buffer Data
		keys := []string{args[0]}
		err = UpdateLedger(stub, "InvoiceTable", keys, buff)
		if err != nil {
			fmt.Println("{\"status\":\"error\",\"message\":\"Error while inserting record\"}")
			return buff, err
		}				
	}

	secret_key, _ := json.Marshal(invoiceObject.AES_Key)
	fmt.Println(string(secret_key))
	return secret_key, nil
}

func CreateInvoiceObject(args []string) (InvoiceObject, error) {
	var err error
	var invoice InvoiceObject
	// Check there are 7 arguments
	if len(args) != 6 {
		fmt.Println("CreateInvoiceItem(): Incorrect number of arguments. Expecting 6")
		return invoice, errors.New("{\"status\":\"error\",\"message\":\"Incorrect number of arguments. Expecting 6\"}")
	}
	// Validate Invoice ID is an integer
	_, err = strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("CreateInvoiceItem():Inovoice ID must be an integer")
		return invoice, errors.New("{\"status\":\"error\",\"message\":\"Inovoice ID must be an integer\"}")
	}
	AES_key, _ := GenAESKey()
	//AES_enc := Encrypt(AES_key, []byte(args[4]))
	invoice = InvoiceObject{args[0], args[1], args[2], args[3], args[4],"",AES_key,"",args[4],args[5],globalkey}
	fmt.Println("CreateInvoiceObject(): Invoice Object created: ID# ", invoice.InvoiceID, "\n AES Key: ", invoice.AES_Key)
	return invoice, nil
}


////////////////////////////////////////// ECNRYPTION FUNCTIONS /////////////////////////////////////////////////
func GenAESKey() ([]byte, error) {
	return GetRandomBytes(AESKeyLength)
}

func GetRandomBytes(len int) ([]byte, error) {
	key := make([]byte, len)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func Encrypt(key []byte, ba []byte) []byte {
	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	// Empty array of 16 + ba length
	// Include the IV at the beginning
	ciphertext := make([]byte, aes.BlockSize+len(ba))
	// Slice of first 16 bytes
	iv := ciphertext[:aes.BlockSize]
	// Write 16 rand bytes to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)
	// Encrypt bytes from ba to ciphertext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], ba)
	return ciphertext
}

/////////////////// DATA PASRSING FUNCTIONS ///////////////////////////////////////////////////////////////
func ARtoJSON(ar InvoiceObject) ([]byte, error) {
	ajson, err := json.Marshal(ar)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return ajson, nil
}