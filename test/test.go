package main

import (		
	"errors"
	"github.com/hyperledger/fabric/core/chaincode/shim"		
)

//Estructuras
type SimpleChaincode struct {
}

//Funciones principales
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		//fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {	
	return []byte("Success"), nil
}
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {	
	if len(args) != 2{
		return nil,errors.New("NEW ERROR")
	}
	return []byte("HOLA"),nil		
}
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub,function string, args []string) ([]byte,error) {
	return nil,nil	
}