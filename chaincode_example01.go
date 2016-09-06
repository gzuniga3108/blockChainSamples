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
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//Custom structs
type User struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	balance 	float64 `json:"balance"`
}

// Init callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will transfer X units from A to B upon invoke
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	//var err error	
	return []byte("Code deployed"), nil
}
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	// Transaction makes payment of X units from A to B
	//var err error
	/*X, err = strconv.Atoi(args[0])
	Aval = Aval - X
	Bval = Bval + X
	ts, err2 := stub.GetTxTimestamp()
	if err2 != nil {
		fmt.Printf("Error getting transaction timestamp: %s", err2)
	}
	fmt.Printf("Transaction Time: %v,Aval = %d, Bval = %d\n", ts, Aval, Bval)*/
	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	return nil, nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}