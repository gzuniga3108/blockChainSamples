/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (	
	"fmt"
	"errors"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"strconv"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//Custom structs
type Account struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Balance 	float64 `json:"balance"`
}

type Game struct{
	ID   string `json:"id"`
	Name string `json:"name"`
	Price  float64 `json:"price"`
	Status string `json:"status"`
}



// Init callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will transfer X units from A to B upon invoke
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	//var err error	
	return []byte("{\"Success\":\"Deploy completed\"}"), nil
}
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function == "createAccount"{
		return t.createAccount(stub,args)
	}else if function == "purchaseCredit"{
		return t.purchaseCredit(stub,args)
	}else if function == "purchaseGame"{
		return t.purchaseGame(stub,args)
	}else if function == "addGame"{
		return t.addGame(stub,args)
	}else if function == "deleteGame"{
		return	t.deleteGame(stub,args)
	}
	return nil,errors.New("Cannot find function")		
}

// Query callback representing the query of a chaincode
// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {	
	if function == "readAccountState" {											
		return t.readAccountState(stub,args)
	}
	if function == "readGameInformation"{
		return t.readGameInformation(stub,args)
	}
	return nil, errors.New("Received unknown function query")
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

//Custom functions for putting state
func (t *SimpleChaincode) createAccount(stub *shim.ChaincodeStub,args []string) ([]byte, error){		
	fmt.Println("Creating account")
	if len(args) != 2 {		
		return nil,errors.New("Incorrect number of arguments. Expecting 2")
	}
	var account = Account{ID: args[0], Name:args[1], Balance: 0}	
	accountBytes, err := json.Marshal(&account)
	if(err != nil){
		return nil,err
	}
	err = stub.PutState(account.ID, accountBytes)
	if(err != nil){
		return nil,err	
	}
	return []byte("{\"Success:\",\"Account created succesfully\"}"),nil
}

func (t *SimpleChaincode) purchaseCredit(stub *shim.ChaincodeStub,args []string)([]byte, error){	
	var id string	
	var account Account	
	if len(args) != 2{
		return nil,errors.New("Expecting receive 2 arguments, 1 received")
	}
	id = args[0]
	amount,err := strconv.ParseFloat(args[1],64)	
	accountBytes, err := stub.GetState(id)
	if err != nil {		
		return nil,errors.New("Account not found " + id)
	}
	err = json.Unmarshal(accountBytes, &account)
	if err != nil {		
		return nil, errors.New("Error unmarshalling user account " + id)
	}	
	account.Balance = account.Balance + amount	
	accountBytes2, err := json.Marshal(&account)
	if(err != nil){
		return nil,err
	}
	err = stub.PutState(account.ID, accountBytes2)
	if(err != nil){
		return nil,err	
	}
	return []byte("{\"Success:\",\"Purchase completed succesfully\"}"),nil	
}

func (t *SimpleChaincode) purchaseGame(stub *shim.ChaincodeStub,args []string)([]byte, error){	
	var game Game
	var account Account
	var idAccount,id string	
	var amount float64
	var jsonResp string
	if len(args) != 3{
		return nil,errors.New("Expecting 2 arguments")
	}
	idAccount = args[0]
	id = args[1]
	quantity,err := strconv.ParseFloat(args[2], 64)
	accountBytes,err := stub.GetState(idAccount)
	if err != nil{
		return nil,errors.New("Error retrieving account information")
	}
	err = json.Unmarshal(accountBytes,&account)
	if(err != nil){
		return nil,errors.New("Error unmarshalling account information")
	}
	gameBytes,err := stub.GetState(id)
	if err != nil{
		return nil,errors.New("Error getting game information");
	}
	err = json.Unmarshal(gameBytes,&game);
	if err != nil{
		return nil,errors.New("Error unmarshalling game information");
	}	

	amount = game.Price * quantity;
	if amount > account.Balance{
		return nil,errors.New("Not enough money for buying games")
	}

	account.Balance = account.Balance - amount
	accountBytes2,err := json.Marshal(&account)
	err = stub.PutState(account.ID,accountBytes2)
	if err != nil {
		return nil,errors.New("Error saving the state of the account")	
	}
	jsonResp = "\"Success\":\"Account status updated succesfully\"";
	return []byte(jsonResp),nil
}
func (t *SimpleChaincode) addGame(stub *shim.ChaincodeStub,args []string)([]byte, error){	
	if len(args) != 3{
		return nil,errors.New("Expecting 3 arguments")
	}
	var game Game	
	var prefix string	
	price,err := strconv.ParseFloat(args[3],64)	
	if err != nil{
		return nil,errors.New("Error parsing game price")	
	}
	game = Game{ID: args[0], Name:args[1], Price: price, Status: "Active"}	
	prefix = "gm"
	gameBytes, err := json.Marshal(&game)
	if err != nil{
		return nil,errors.New("Error while marshaling game information")
	}
	err = stub.PutState(prefix+"-"+game.ID,gameBytes)
	if err != nil{
		return nil,errors.New("Error saving game information")
	}
	return []byte("\"Success\":\"Game information saved succesfully\""),nil		
}
func (t *SimpleChaincode) deleteGame(stub *shim.ChaincodeStub,args []string)([]byte, error){	
	return nil,nil
}

//Custom functions for getting state
func (t *SimpleChaincode) readAccountState(stub *shim.ChaincodeStub,args []string)([]byte, error){
	var id,jsonResp string	
	var err error
	if len(args) != 1 {
		return nil,errors.New("Incorrect number of arguments, expecting 1")
	}
	id = args[0]
	jsonResp =  args[0]
	valAsbytes,err := stub.GetState(id)
	if err != nil{
		jsonResp = "{\"Error\":\"Failed to get state for "+ id +"\"}"
		return nil, errors.New(jsonResp)
	}	
	return valAsbytes,nil
}

func (t *SimpleChaincode) readGameInformation(stub *shim.ChaincodeStub,args []string)([]byte,error){
	var id	string
	if len(args) != 1{
		return nil,errors.New("Expecting 1 argument")
	}
	id = args[0]
	gameBytes,err := stub.GetState(id)
	if err != nil{
		return nil,errors.New("Error retrieving information of the game with id "+id)
	}
	return gameBytes,nil
}