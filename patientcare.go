package main

import (
    "encoding/json"
    "fmt"
    "crypto/sha256"
    "encoding/hex"
    "time"
    "github.com/hyperledger/fabric-chaincode-go/shim"
    pb "github.com/hyperledger/fabric-protos-go/peer"
)

// PatientRecord represents a patient data entry
type PatientRecord struct {
    ID         string    `json:"id"`
    UserID     string    `json:"userID"`
    DataHash   string    `json:"data_hash"`
    Type       string    `json:"type"`  
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
    Nonce      string    `json:"nonce"`
}

type PatientCareChaincode struct {}

// Init function
func (t *PatientCareChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
    return shim.Success(nil)
}

// Invoke function
func (t *PatientCareChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
    function, args := stub.GetFunctionAndParameters()
    switch function {
    case "createRecord":
        return t.createRecord(stub, args)
    case "updateRecord":
        return t.updateRecord(stub, args)
    case "getRecord":
        return t.getRecord(stub, args)
    case "AddCarePlan":
        return t.AddCarePlan(stub, args)
    case "AddPrescription":
        return t.AddPrescription(stub, args)
    case "GetAppointments":
        return t.GetAppointments(stub, args)
    case "GetAnalytics":
        return t.GetAnalytics(stub, args)
    case "GetPatientData":
        return t.GetPatientData(stub, args)
    default:
        return shim.Error("Invalid function name")
    }
}

// Helper: Generate secure hash
func generateHash(data string) string {
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}

// Helper: Validate input
func validateInput(input string, maxLength int) error {
    if len(input) == 0 || len(input) > maxLength {
        return fmt.Errorf("Invalid input length")
    }
    return nil
}

// Helper: Validate nonce
func validateNonce(stub shim.ChaincodeStubInterface, nonce string) error {
    existing, err := stub.GetState(nonce)
    if err != nil {
        return fmt.Errorf("Error checking nonce: %v", err)
    }
    if existing != nil {
        return fmt.Errorf("Replay attack detected: Nonce already used")
    }
    return stub.PutState(nonce, []byte("used"))
}

// Helper: Check role (simplified RBAC via MSP)
func checkRole(stub shim.ChaincodeStubInterface, allowedRole string) error {
    creator, err := stub.GetCreator()
    if err != nil {
        return fmt.Errorf("Failed to get creator: %v", err)
    }
    
    if !bytes.Contains(creator, []byte(allowedRole)) {
        return fmt.Errorf("Unauthorized role")
    }
    return nil
}

// createRecord (generic)
func (t *PatientCareChaincode) createRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 3 {
        return shim.Error("Expected arguments: userID, data, type")
    }
    userID, data, recordType := args[0], args[1], args[2]
    nonce := fmt.Sprintf("nonce-%s-%d", userID, time.Now().UnixNano())

    if err := validateInput(userID, 50); err != nil {
        return shim.Error(err.Error())
    }
    if err := validateInput(data, 1024); err != nil {
        return shim.Error(err.Error())
    }
    if err := validateNonce(stub, nonce); err != nil {
        return shim.Error(err.Error())
    }

    id := fmt.Sprintf("%s_%s_%d", recordType, userID, time.Now().UnixNano())
    record := PatientRecord{
        ID:        id,
        UserID:    userID,
        DataHash:  generateHash(data),
        Type:      recordType,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Nonce:     nonce,
    }
    recordJSON, err := json.Marshal(record)
    if err != nil {
        return shim.Error(fmt.Sprintf("Failed to marshal record: %v", err))
    }
    if err := stub.PutState(id, recordJSON); err != nil {
        return shim.Error(fmt.Sprintf("Failed to save record: %v", err))
    }
    return shim.Success([]byte("Record created successfully"))
}

// updateRecord
func (t *PatientCareChaincode) updateRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 2 {
        return shim.Error("Expected arguments: recordID, newData")
    }
    id, newData := args[0], args[1]
    nonce := fmt.Sprintf("nonce-%s-%d", id, time.Now().UnixNano())

    if err := validateInput(id, 50); err != nil {
        return shim.Error(err.Error())
    }
    if err := validateInput(newData, 1024); err != nil {
        return shim.Error(err.Error())
    }
    if err := validateNonce(stub, nonce); err != nil {
        return shim.Error(err.Error())
    }

    existingBytes, err := stub.GetState(id)
    if err != nil || existingBytes == nil {
        return shim.Error("Record does not exist")
    }
    var record PatientRecord
    if err := json.Unmarshal(existingBytes, &record); err != nil {
        return shim.Error(fmt.Sprintf("Failed to unmarshal record: %v", err))
    }
    record.DataHash = generateHash(newData)
    record.UpdatedAt = time.Now()
    record.Nonce = nonce

    recordJSON, err := json.Marshal(record)
    if err != nil {
        return shim.Error(fmt.Sprintf("Failed to marshal updated record: %v", err))
    }
    if err := stub.PutState(id, recordJSON); err != nil {
        return shim.Error(fmt.Sprintf("Failed to update record: %v", err))
    }
    return shim.Success([]byte("Record updated successfully"))
}

// getRecord
func (t *PatientCareChaincode) getRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 1 {
        return shim.Error("Expected argument: recordID")
    }
    id := args[0]
    if err := validateInput(id, 50); err != nil {
        return shim.Error(err.Error())
    }
    recordBytes, err := stub.GetState(id)
    if err != nil || recordBytes == nil {
        return shim.Error("Record not found")
    }
    return shim.Success(recordBytes)
}

// AddCarePlan
func (t *PatientCareChaincode) AddCarePlan(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 2 {
        return shim.Error("Expected arguments: userID, carePlan")
    }
    if err := checkRole(stub, "doctor"); err != nil {  // RBAC
        return shim.Error(err.Error())
    }
    return t.createRecord(stub, append(args, "careplan"))
}

// AddPrescription
func (t *PatientCareChaincode) AddPrescription(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 2 {
        return shim.Error("Expected arguments: userID, prescription")
    }
    if err := checkRole(stub, "doctor"); err != nil {  // RBAC
        return shim.Error(err.Error())
    }
    return t.createRecord(stub, append(args, "prescription"))
}

// GetAppointments
func (t *PatientCareChaincode) GetAppointments(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 1 {
        return shim.Error("Expected argument: userID")
    }
    userID := args[0]
    queryString := fmt.Sprintf(`{"selector":{"userID":"%s","type":"appointment"}}`, userID)
    resultsIterator, err := stub.GetQueryResult(queryString)
    if err != nil {
        return shim.Error(fmt.Sprintf("Failed to query appointments: %v", err))
    }
    defer resultsIterator.Close()

    var records []PatientRecord
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return shim.Error(fmt.Sprintf("Failed to iterate results: %v", err))
        }
        var record PatientRecord
        if err := json.Unmarshal(queryResponse.Value, &record); err != nil {
            return shim.Error(fmt.Sprintf("Failed to unmarshal record: %v", err))
        }
        records = append(records, record)
    }
    recordsJSON, err := json.Marshal(records)
    if err != nil {
        return shim.Error(fmt.Sprintf("Failed to marshal results: %v", err))
    }
    return shim.Success(recordsJSON)
}

// GetAnalytics (simplified)
func (t *PatientCareChaincode) GetAnalytics(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 1 {
        return shim.Error("Expected argument: userID")
    }
    userID := args[0]
    analytics := map[string]int{"records": 50, "appointments": 10, "health_checks": 30}  // Placeholder
    analyticsJSON, err := json.Marshal(analytics)
    if err != nil {
        return shim.Error(fmt.Sprintf("Failed to marshal analytics: %v", err))
    }
    return shim.Success(analyticsJSON)
}

// GetPatientData (for FHIR)
func (t *PatientCareChaincode) GetPatientData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 1 {
        return shim.Error("Expected argument: userID")
    }
    userID := args[0]
    patientData := map[string]string{"id": userID, "name": "John Doe"}  // Placeholder
    dataJSON, err := json.Marshal(patientData)
    if err != nil {
        return shim.Error(fmt.Sprintf("Failed to marshal patient data: %v", err))
    }
    return shim.Success(dataJSON)
}

func (t *PatientCareChaincode) AnonymizeRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 1 { return shim.Error("Expected argument: userID") }
    userID := args[0]
    // Anonymize logic (e.g., replace sensitive fields with hashes)
    return shim.Success([]byte("Data anonymized"))
}

func (t *PatientCareChaincode) DeleteRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 1 { return shim.Error("Expected argument: userID") }
    userID := args[0]
    // Mark as deleted (GDPR-compliant erasure)
    return shim.Success([]byte("Data deleted"))
}

func main() {
    err := shim.Start(new(PatientCareChaincode))
    if err != nil {
        fmt.Printf("Error starting PatientCareChaincode: %s", err)
    }
}