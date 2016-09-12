package main

import(
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
	"encoding/json"
	"runtime"
	"os"
	"strings"
)

//Estructuras
type SimpleChaincode struct {
}

type UserObject struct {
	UserID    		string
	UserPass		string
	RecType   		string // Type = USER
	Name      		string
	UserType  		string //BY Buyer, SL Seller	
	CashBalance		string
}


//Variables globales
var gopath string
var ccPath string
var recType = []string{"USER"}
var tables = []string{"UserTable"}

//Funciones necesarias para el proyecto (main,init,invoke,query)
func main() {
	// maximize CPU usage for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Starting Item Auction Application chaincode BlueMix ver 0.25 Dated 2016-07-17 15.20.00 ")
	gopath = os.Getenv("GOPATH")
	if len(os.Args) == 2 && strings.EqualFold(os.Args[1], "DEV") {
		fmt.Println("----------------- STARTED IN DEV MODE -------------------- ")
		//set chaincode path for DEV MODE
		ccPath = fmt.Sprintf("%s/src/github.com/hyperledger/fabric/auction/art/artchaincode/", gopath)
	} else {
		fmt.Println("----------------- STARTED IN NET MODE -------------------- ")
		//set chaincode path for NET MODE
		ccPath = fmt.Sprintf("%s/src/github.com/ITPeople-Blockchain/auction/art/artchaincode/", gopath)
	}
	// Start the shim -- running the fabric
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Println("Error starting Item Fun Application chaincode: %s", err)
	}
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

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	// TODO - Include all initialization to be complete before Invoke and Query
	// Uses tables to delete tables if they exist and re-create them
	//myLogger.Info("[Trade and Auction Application] Init")
	fmt.Println("[Trade and Auction Application] Init")
	var err error
	for _, val := range tables {
		err = stub.DeleteTable(val)
		if err != nil {
			return nil, fmt.Errorf("Init(): DeleteTable of %s  Failed ", val)
		}
		err = InitLedger(stub, val)
		if err != nil {
			return nil, fmt.Errorf("Init(): InitLedger of %s  Failed ", val)
		}
	}
	// Update the ledger with the Application version
	err = stub.PutState("version", []byte(strconv.Itoa(23)))
	if err != nil {
		return nil, err
	}
	fmt.Println("Init() Initialization Complete  : ", args)
	return []byte("Init(): Initialization Complete"), nil
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

//Helper functios
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

func InvokeFunction(fname string) func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	InvokeFunc := map[string]func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error){
		"CreateUser":	CreateUser,		
	}
	return InvokeFunc[fname]
}


//////////////////////////////////////////////////////////////
// Query Functions based on Function name
//
//////////////////////////////////////////////////////////////
func QueryFunction(fname string) func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	QueryFunc := map[string]func(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error){		
		"GetUser":               GetUser,
	}
	return QueryFunc[fname]
}

////////////////////////////////////////////////////////////////////////////
// Open a User Registration Table if one does not exist
// Register users into this table
////////////////////////////////////////////////////////////////////////////
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


func InitLedger(stub *shim.ChaincodeStub, tableName string) error {
	// Generic Table Creation Function - requires Table Name and Table Key Entry
	// Create Table - Get number of Keys the tables supports
	// This version assumes all Keys are String and the Data is Bytes
	// This Function can replace all other InitLedger function in this app such as InitItemLedger()
	nKeys := GetNumberOfKeys(tableName)
	if nKeys < 1 {
		fmt.Println("Atleast 1 Key must be provided \n")
		fmt.Println("Auction_Application: Failed creating Table ", tableName)
		return errors.New("Auction_Application: Failed creating Table " + tableName)
	}
	var columnDefsForTbl []*shim.ColumnDefinition
	for i := 0; i < nKeys; i++ {
		columnDef := shim.ColumnDefinition{Name: "keyName" + strconv.Itoa(i), Type: shim.ColumnDefinition_STRING, Key: true}
		columnDefsForTbl = append(columnDefsForTbl, &columnDef)
	}
	columnLastTblDef := shim.ColumnDefinition{Name: "Details", Type: shim.ColumnDefinition_BYTES, Key: false}
	columnDefsForTbl = append(columnDefsForTbl, &columnLastTblDef)

	// Create the Table (Nil is returned if the Table exists or if the table is created successfully
	err := stub.CreateTable(tableName, columnDefsForTbl)
	if err != nil {
		fmt.Println("Auction_Application: Failed creating Table ", tableName)
		return errors.New("Auction_Application: Failed creating Table " + tableName)
	}
	return err
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

	//fmt.Println("User Query Response:", row)
	//jsonResp := "{\"Owner\":\"" + string(row.Columns[nCol].GetBytes()) + "\"}"
	//fmt.Println("User Query Response:%s\n", jsonResp)
	Avalbytes := row.Columns[nCol].GetBytes()

	// Perform Any additional processing of data
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

	// Identify Record Type by scanning the args for one of the recTypes
	// This is kind of a post-processor once the query fetches the results
	// RecType is the style of programming in the punch card days ..
	// ... well

	var dat map[string]interface{}

	if err := json.Unmarshal(Avalbytes, &dat); err != nil {
		panic(err)
	}

	var recType string
	recType = dat["RecType"].(string)
		switch recType {
		case "USER":
			ur, err := JSONtoUser(Avalbytes) //
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

//Custom functios
//////////////////////////////////////////////////////////
// Converts an User Object to a JSON String
//////////////////////////////////////////////////////////
func UsertoJSON(user UserObject) ([]byte, error) {

	ajson, err := json.Marshal(user)
	if err != nil {
		fmt.Println("UsertoJSON error: ", err)
		return nil, err
	}
	fmt.Println("UsertoJSON created: ", ajson)
	return ajson, nil
}

//////////////////////////////////////////////////////////
// Converts an User Object to a JSON String
//////////////////////////////////////////////////////////
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

func CreateUserObject(args []string) (UserObject, error) {
	var err error
	var aUser UserObject
	// Check there are 5 Arguments
	if len(args) != 6 {
		fmt.Println("CreateUserObject(): Incorrect number of arguments. Expecting 10 ")
		return aUser, errors.New("CreateUserObject() : Incorrect number of arguments. Expecting 10 ")
	}
	// Validate UserID is an integer
	_, err = strconv.Atoi(args[0])
	if err != nil {
		return aUser, errors.New("CreateUserObject() : User ID should be an integer")
	}
	aUser = UserObject{args[0], args[1], args[2], args[3], args[4], args[5]}
	fmt.Println("CreateUserObject() : User Object : ", aUser)
	return aUser, nil
}


func CreateUser(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	record, err := CreateUserObject(args[0:]) //
	if err != nil {
		return nil, err
	}
	buff, err := UsertoJSON(record) //
	if err != nil {
		fmt.Println("PostuserObject() : Failed Cannot create object buffer for write : ", args[0])
		return nil, errors.New("PostUser(): Failed Cannot create object buffer for write : " + args[0])
	} else {
		// Update the ledger with the Buffer Data
		// err = stub.PutState(args[0], buff)
		keys := []string{args[0],args[1]}
		err = UpdateLedger(stub, "UserTable", keys, buff)
		if err != nil {
			fmt.Println("PostUser() : write error while inserting record")
			return nil, err
		}

		// Post Entry into UserCatTable - i.e. User Category Table
		keys = []string{"2016", args[0], args[1]}
		err = UpdateLedger(stub, "UserCatTable", keys, buff)
		if err != nil {
			fmt.Println("PostUser() : write error while inserting recordinto UserCatTable \n")
			return nil, err
		}
	}
	return buff, err
}

func GetUser(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var err error
	// Get the Object and Display it
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

/////////////////////////////////////////////////////////////////////////////////////////////////////
// A Map that holds TableNames and the number of Keys
// This information is used to dynamically Create, Update
// Replace , and Query the Ledger
// In this model all attributes in a table are strings
// The chain code does both validation
//              "UserTable":        1, Key: UserID
func GetNumberOfKeys(tname string) int {
	TableMap := map[string]int{
		"UserTable":        2,
	}
	return TableMap[tname]
}
