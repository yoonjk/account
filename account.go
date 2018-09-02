package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/satori/go.uuid"
)

// AccountChaincode example simple Chaincode implementation
type AccountChaincode struct {
}

type vendor struct {
	VendorCode    string     `json:"vendorCode"`
	RegistDate    string     `json:"registDate"`
	AccountNumber string     `json:"accountNumber"`
	AccStatus     string     `json:"accStatus"`
	AccProcStatus string     `json:"accProcStatus"`
	Apprvrs       []approver `json:"apprvrs"`
}

type approver struct {
	ApprvrID   string `json:"apprvrID"`
	ApprvrDate string `json:"apprvrDate"`
}

// main
// ================================
func main() {
	err := shim.Start(new(AccountChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// genUUIDv4
// ==============================
func genUUIDv4() string {
	id, _ := uuid.NewV4()
	fmt.Printf("github.com/satori/go.uuid:   %s\n", id)
	return id.String()
}

// Init initializes chaincode
// ===========================
func (t *AccountChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke Chaincode
// ===========================
func (t *AccountChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "initAcc" {
		return t.initAcc(stub, args)
	} else if function == "addApprover" {
		return t.addApprover(stub, args)
	} else if function == "queryAccount" {
		return t.queryAccount(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// getSelector
//============================
func getSelector(stub shim.ChaincodeStubInterface, queryStr string) ([]byte, error) {
	iterator, err := stub.GetQueryResult(queryStr)

	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.WriteString("[")
	bool := false

	for iterator.HasNext() {
		queryResponse, err := iterator.Next();

		if (err !=nil) {
			return nil, err
		}

		if bool == true {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")

		bool = true
	}

	buffer.WriteString("]")
	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())
	return buffer.Bytes(), nil
}

// queryAccount
//==========================
func (t *AccountChaincode) queryAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	queryStr := args[0]

	queryResults, err := getSelector(stub, queryStr)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

// func (t *AccountChaincode) getSelector(stub shim.ChaincodeStubInterface, queryStr string) ([]byte, error) {
// 	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryStr)

// 	resultsIterator, err := stub.GetQueryResult(queryStr)

// 	if err != nil {
// 		return nil, err
// 	}

// 	var buffer bytes.Buffer

// 	buffer.WriteString("[")
// 	bool := false

// 	for resultsIterator.HasNext() {
// 		queryResponse, err := resultsIterator.Next()

// 		if err != nil {
// 			return nil, err
// 		}

// 		if bool == true {
// 			buffer.WriteString(",")
// 		}

// 		buffer.WriteString("{\"Key\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString(queryResponse.Key)
// 		buffer.WriteString("\"")

// 		buffer.WriteString(", \"Record\":")
// 		buffer.WriteString(string(queryResponse.Value))
// 		buffer.WriteString("}")

// 		bool = true
// 	}

// 	buffer.WriteString("]")
// 	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())
// 	return buffer.Bytes(), nil
// }

// InitAccount Insert Account Data
// ==============================
func (t *AccountChaincode) initAcc(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("====initAcc ") //er
	if len(args) != 7 {
		return shim.Error("====Invalid Parameters: Expecting 7 parameters")
	}
	var vendorJSON vendor

	vendorID := fmt.Sprintf("acc-%s", genUUIDv4())
	vendorCode := args[0]
	registDate := args[1]
	accountNumber := args[2]
	accStatus := args[3]
	accProcStatus := args[4]
	apprvrID := args[5]
	apprvrDate := args[6]

	vendorCodeAsBytes, err := stub.GetState(vendorID)

	if err != nil {
		return shim.Error("Failed to get vendor: " + err.Error())
	} else if vendorCodeAsBytes != nil {
		return shim.Error("exists vendorId:" + vendorID)
	}

	vendorJSON.VendorCode = vendorCode
	vendorJSON.AccProcStatus = accProcStatus
	vendorJSON.AccStatus = accStatus
	vendorJSON.RegistDate = registDate
	vendorJSON.AccountNumber = accountNumber
	vendorJSON.Apprvrs = append(vendorJSON.Apprvrs, approver{
		ApprvrID:   apprvrID,
		ApprvrDate: apprvrDate,
	})

	vendorJSONBytes, err := json.Marshal(vendorJSON)

	err = stub.PutState(vendorID, vendorJSONBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	indexName := "vendor~name"
	vendorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{vendorJSON.AccountNumber, vendorJSON.VendorCode})

	if err != nil {
		return shim.Error(err.Error())
	}

	value := []byte{0x00}

	err = stub.PutState(vendorNameIndexKey, value)
	fmt.Println("- end init Account")

	return shim.Success(nil)
}

// getAccount read Account Data
// ==================================================
func (t *AccountChaincode) getAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var vendorCode, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	vendorCode = args[0]
	valAsbytes, err := stub.GetState(vendorCode) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + vendorCode + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + vendorCode + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// delete delete account and index
// =============================
func (t *AccountChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var jsonResp string
	var vendorJSON vendor

	if len(args) != 1 {
		return shim.Error("Invaild parameter.Expecting 1 paramter")
	}

	vendorCode := args[0]

	vendorAsBytes, err := stub.GetState(vendorCode)

	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + vendorCode + "\"}"
		return shim.Error(jsonResp)
	} else if vendorAsBytes == nil {
		jsonResp = "{\"Error\":\"Account does not exist: " + vendorCode + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(vendorCode)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	err = json.Unmarshal([]byte(vendorAsBytes), &vendorJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	indexName := "vendor~name"
	vendorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{vendorJSON.AccountNumber, vendorJSON.VendorCode})

	//  Delete index entry to state.
	err = stub.DelState(vendorNameIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	return shim.Success(nil)
}

func (t *AccountChaincode) addApprover(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var vendorJSON vendor

	vendorCode := args[0]
	vendorAsBytes, err := stub.GetState(vendorCode)

	if err != nil {
		return shim.Error("Failed to get Vendor")
	} else if vendorAsBytes == nil {
		return shim.Error("Vendor not found:" + vendorCode)
	}

	apprvrID := args[1]
	apprvrDate := args[2]

	err = json.Unmarshal([]byte(vendorAsBytes), &vendorJSON)

	if err != nil {
		return shim.Error(err.Error())
	}

	vendorJSON.Apprvrs = append(vendorJSON.Apprvrs, approver{
		ApprvrID:   apprvrID,
		ApprvrDate: apprvrDate,
	})

	vendorJSONAsBytes, err := json.Marshal(vendorJSON)
	err = stub.PutState(vendorCode, vendorJSONAsBytes)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(vendorJSONAsBytes)
}


