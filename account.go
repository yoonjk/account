package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// AccountChaincode example simple Chaincode implementation
type AccountChaincode struct {
}

// account
type account struct {
	AccountNumber string `json:"accountNumber"`
	Seq           int    `json:"seq"`
	BankCode      string `json:"bankCode"`
	CtryCode      string `json:"ctryCode"`
	Owner         string `json:"owner"`
}

type vendor struct {
	VendorCode  string    `json:"vendorCode"`
	BizRegistNo string    `json:"bizRegistNo"`
	Accounts    []account `json:"accounts"`
}

// main
// ================================
func main() {
	err := shim.Start(new(AccountChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
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

	if function == "initAccount" {
		return t.initAccount(stub, args)
	}

	return shim.Success(nil)
}

// // queryAccount inquire Account
// // ============================
// func (t *AccountChaincode) queryAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

// 	// "queryString"
// 	if len(args) < 1 {
// 		return shim.Error("Incorrect number of arguments. Expecting 1")
// 	}
// 	queryStr := args[0]

// 	queryResults, err := getQueryResultForQueryString(stub, queryStr)

// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	return shim.Success(queryResults)
// }

// func (t *AccountChaincode) getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryStr string) ([]byte, error) {
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
func (t *AccountChaincode) initAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 7 {
		return shim.Error("Invalid Parameters: Expecting 7 parameters")
	}
	var vendorJSON vendor

	vendorCode := args[0]
	bizRegistNo := args[1]
	bankCode := args[4]
	ctryCode := args[5]
	owner := args[6]
	accountNumber := args[2]

	seq, err := strconv.Atoi(args[3])

	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	vendorCodeAsBytes, err := stub.GetState(vendorCode)

	if err != nil {
		return shim.Error("Failed to get vendor: " + err.Error())
	} else if vendorCodeAsBytes != nil {
		fmt.Println("account exists:" + vendorCode)
		err = json.Unmarshal([]byte(vendorCodeAsBytes), &vendorJSON)

		if err != nil {
			return shim.Error("Failed to get Vendor:" + err.Error())
		}
		size := len(vendorJSON.Accounts)
		fmt.Println("account exists:" + vendorCode + ",len:" + string(size))
	}

	vendorJSON.VendorCode = vendorCode
	vendorJSON.BizRegistNo = bizRegistNo
	vendorJSON.Accounts = append(vendorJSON.Accounts, account{
		AccountNumber: accountNumber,
		Seq:           seq,
		BankCode:      bankCode,
		CtryCode:      ctryCode,
		Owner:         owner,
	})
	vendorJSONBytes, err := json.Marshal(vendorJSON)

	err = stub.PutState(vendorCode, vendorJSONBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	indexName := "vendor~name"
	vendorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{vendorJSON.BizRegistNo, vendorJSON.VendorCode})

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
	vendorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{vendorJSON.BizRegistNo, vendorJSON.VendorCode})

	//  Delete index entry to state.
	err = stub.DelState(vendorNameIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	return shim.Success(nil)
}

// func (t *AccountChaincode) transferAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

// 	if len(args) < 2 {
// 		return shim.Error("Incorrect number of arguments. Expecting 2")
// 	}

// 	accountNumber := args[0]
// 	newOwner := args[1]

// 	accountAsBytes, err := stub.GetState(accountNumber)
// 	if err != nil {
// 		return shim.Error("Failed to get account:" + err.Error())
// 	} else if accountAsBytes == nil {
// 		return shim.Error("Account does not exist")
// 	}

// 	accountJSON := account{}

// 	err = json.Unmarshal([]byte(accountAsBytes), &accountJSON)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	accountJSON.Owner = newOwner

// 	accountAsJSONBytes, _ := json.Marshal(accountJSON)

// 	stub.PutState(accountNumber, accountAsJSONBytes)

// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	fmt.Println("- end transferAccount (success)")

// 	return shim.Success(nil)
// }

// func (t *AccountChaincode) getAccountByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {
// 	var buffer bytes.Buffer

// 	if len(args) < 2 {
// 		return shim.Error("Incorrect number of arguments. Expecting 2")
// 	}

// 	startKey := args[0]
// 	endKey := args[1]

// 	resultsIterator, err := stub.GetStateByRange(startKey, endKey)

// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	defer resultsIterator.Close()

// 	buffer.WriteString("[")

// 	bArrayMemberAlreadyWritten := false
// 	for resultsIterator.HasNext() {
// 		queryResponse, err := resultsIterator.Next()
// 		if err != nil {
// 			return shim.Error(err.Error())
// 		}
// 		// Add a comma before array members, suppress it for the first array member
// 		if bArrayMemberAlreadyWritten == true {
// 			buffer.WriteString(",")
// 		}
// 		buffer.WriteString("{\"Key\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString(queryResponse.Key)
// 		buffer.WriteString("\"")

// 		buffer.WriteString(", \"Record\":")
// 		// Record is a JSON object, so we write as-is
// 		buffer.WriteString(string(queryResponse.Value))
// 		buffer.WriteString("}")
// 		bArrayMemberAlreadyWritten = true
// 	}
// 	buffer.WriteString("]")

// 	fmt.Printf("- getAccountByRange queryResult:\n%s\n", buffer.String())

// 	return shim.Success(buffer.Bytes())
// }

// func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

// 	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

// 	resultsIterator, err := stub.GetQueryResult(queryString)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resultsIterator.Close()

// 	// buffer is a JSON array containing QueryRecords
// 	var buffer bytes.Buffer
// 	buffer.WriteString("[")

// 	bArrayMemberAlreadyWritten := false
// 	for resultsIterator.HasNext() {
// 		queryResponse, err := resultsIterator.Next()
// 		if err != nil {
// 			return nil, err
// 		}
// 		// Add a comma before array members, suppress it for the first array member
// 		if bArrayMemberAlreadyWritten == true {
// 			buffer.WriteString(",")
// 		}
// 		buffer.WriteString("{\"Key\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString(queryResponse.Key)
// 		buffer.WriteString("\"")

// 		buffer.WriteString(", \"Record\":")
// 		// Record is a JSON object, so we write as-is
// 		buffer.WriteString(string(queryResponse.Value))
// 		buffer.WriteString("}")
// 		bArrayMemberAlreadyWritten = true
// 	}
// 	buffer.WriteString("]")

// 	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

// 	return buffer.Bytes(), nil
// }

// func (t *AccountChaincode) getHistoryForAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
// 	if len(args) < 1 {
// 		return shim.Error("Incorrect number of arguments. Expecting 1")
// 	}

// 	accountNumber := args[0]

// 	fmt.Printf("- start getHistoryForAccount: %s\n", accountNumber)

// 	resultsIterator, err := stub.GetHistoryForKey(accountNumber)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	defer resultsIterator.Close()

// 	// buffer is a JSON array containing historic values for the marble
// 	var buffer bytes.Buffer
// 	buffer.WriteString("[")

// 	bArrayMemberAlreadyWritten := false
// 	for resultsIterator.HasNext() {
// 		response, err := resultsIterator.Next()
// 		if err != nil {
// 			return shim.Error(err.Error())
// 		}
// 		// Add a comma before array members, suppress it for the first array member
// 		if bArrayMemberAlreadyWritten == true {
// 			buffer.WriteString(",")
// 		}
// 		buffer.WriteString("{\"TxId\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString(response.TxId)
// 		buffer.WriteString("\"")

// 		buffer.WriteString(", \"Value\":")
// 		// if it was a delete operation on given key, then we need to set the
// 		//corresponding value null. Else, we will write the response.Value
// 		//as-is (as the Value itself a JSON marble)
// 		if response.IsDelete {
// 			buffer.WriteString("null")
// 		} else {
// 			buffer.WriteString(string(response.Value))
// 		}

// 		buffer.WriteString(", \"Timestamp\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
// 		buffer.WriteString("\"")

// 		buffer.WriteString(", \"IsDelete\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString(strconv.FormatBool(response.IsDelete))
// 		buffer.WriteString("\"")

// 		buffer.WriteString("}")
// 		bArrayMemberAlreadyWritten = true
// 	}
// 	buffer.WriteString("]")

// 	fmt.Printf("- getHistoryForAccount returning:\n%s\n", buffer.String())

// 	return shim.Success(buffer.Bytes())
// }

// func (t *AccountChaincode) transferMarblesBasedOnColor(stub shim.ChaincodeStubInterface, args []string) pb.Response {
// 	//   0       1
// 	// "color", "bob"
// 	if len(args) < 2 {
// 		return shim.Error("Incorrect number of arguments. Expecting 2")
// 	}

// 	color := args[0]
// 	newOwner := strings.ToLower(args[1])
// 	fmt.Println("- start transferMarblesBasedOnColor ", color, newOwner)

// 	indexName := "color~name"
// 	coloredAccountResultsIterator, err := stub.GetStateByPartialCompositeKey(indexName, []string{color})

// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	defer coloredAccountResultsIterator.Close()

// 	// Iterate through result set and for each marble found, transfer to newOwner
// 	var i int
// 	for i = 0; coloredAccountResultsIterator.HasNext(); i++ {
// 		responseRange, err := coloredAccountResultsIterator.Next()

// 		if err != nil {
// 			return shim.Error(err.Error())
// 		}

// 		// get the color and name from color~name composite key
// 		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
// 		if err != nil {
// 			return shim.Error(err.Error())
// 		}

// 		returnedColor := compositeKeyParts[0]
// 		returnedAccountNumber := compositeKeyParts[1]

// 		fmt.Printf("- found a marble from index:%s color:%s name:%s\n", objectType, returnedColor, returnedAccountNumber)

// 		// Now call the transfer function for the found marble.
// 		// Re-use the same function that is used to transfer individual Account
// 		response := t.transferAccount(stub, []string{returnedAccountNumber, newOwner})
// 		// if the transfer failed break out of loop and return error
// 		if response.Status != shim.OK {
// 			return shim.Error("Transfer failed: " + response.Message)
// 		}
// 	}

// 	responsePayload := fmt.Sprintf("Transferred %d %s account to %s", i, color, newOwner)
// 	fmt.Println("- end transferAccountBasedOnColor: " + responsePayload)
// 	return shim.Success([]byte(responsePayload))
// }
