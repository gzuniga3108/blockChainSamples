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
	Name   string `json:"id"`
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
	return nil, errors.New("Received unknown function query")
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

//Custom functions for putting state
func (t *SimpleChaincode) createAccount(stub *shim.ChaincodeStub,args []string)([]byte, error){	
	var err error	
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
	return nil,nil
}
func (t *SimpleChaincode) purchaseGame(stub *shim.ChaincodeStub,args []string)([]byte, error){	
	return nil,nil
}
func (t *SimpleChaincode) addGame(stub *shim.ChaincodeStub,args []string)([]byte, error){	
	return nil,nil
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