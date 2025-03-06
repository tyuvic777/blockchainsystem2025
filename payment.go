package main

import (
    "encoding/json"
    "fmt"
    "github.com/hyperledger/fabric-chaincode-go/shim"
    pb "github.com/hyperledger/fabric-protos-go/peer"
    "crypto/sha256"
    "encoding/hex"
    "net/http"
    "io"
    "time"
    "strconv"
)

// PaymentChaincode represents the payment and token reward system
type PaymentChaincode struct {
}

// TokenBalance represents a user's token balance
type TokenBalance struct {
    UserID          string    `json:"user_id"`
    Balance         float64   `json:"balance"`
    BlockchainHash  string    `json:"blockchain_hash"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// Init initializes the chaincode
func (t *PaymentChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
    /**
     * Initialize the payment chaincode, setting up any initial state if needed.
     * 
     * Returns:
     *   pb.Response: Success response with no payload
     */
    return shim.Success(nil)
}

// Invoke handles chaincode invocations
func (t *PaymentChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
    /**
     * Handle chaincode function calls based on the function name and arguments.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     * 
     * Returns:
     *   pb.Response: Response with success or error message
     */
    fn, args := stub.GetFunctionAndParameters()
    switch fn {
    case "initializeToken":
        return t.initializeToken(stub, args)
    case "rewardPatient":
        return t.rewardPatient(stub, args)
    case "rewardDoctor":
        return t.rewardDoctor(stub, args)
    case "getBalance":
        return t.getBalance(stub, args)
    case "transferTokens":
        return t.transferTokens(stub, args)
    default:
        return shim.Error("Invalid function name. Please provide a valid function (initializeToken, rewardPatient, rewardDoctor, getBalance, transferTokens). Thank you!")
    }
}

// Helper function for role-specific messages
func getRoleMessage(role, feature string, success bool) string {
    /**
     * Generate role-specific, user-friendly messages for chaincode responses.
     * 
     * Args:
     *   role (string): User role (admin, doctor, patient)
     *   feature (string): Feature or action being performed
     *   success (bool): Whether the action succeeded
     * 
     * Returns:
     *   string: Formatted message
     */
    messages := map[string]map[bool]string{
        "admin": {
            true:  fmt.Sprintf("Thank you, Admin! Your action on %s has been completed successfully.", feature),
            false: fmt.Sprintf("Sorry, Admin, we couldn’t process your %s request. Please try again or contact support.", feature),
        },
        "doctor": {
            true:  fmt.Sprintf("Great job, Doctor! Your update to %s was successful.", feature),
            false: fmt.Sprintf("Oops, Doctor, we encountered an issue with your %s. Please try again later or reach out to support.", feature),
        },
        "patient": {
            true:  fmt.Sprintf("Thank you, Patient! Your %s has been updated successfully.", feature),
            false: fmt.Sprintf("Sorry, Patient, we couldn’t complete your %s request. Please try again or contact our support team.", feature),
        },
    }
    return messages[role][success]
}

// initializeToken initializes a user's token balance
func (t *PaymentChaincode) initializeToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    /**
     * Initialize a user's token balance with role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     *   args ([]string): Arguments [userID, initialBalance]
     * 
     * Returns:
     *   pb.Response: Success or error response with role-specific message
     */
    if len(args) != 2 {
        return shim.Error("Please provide user ID and initial balance to initialize tokens. Thank you!")
    }

    cid := ClientIdentity(stub)
    role := "admin" // Default role for initialization, adjust based on MSP
    if cid.AssertAttributeValue("role", "doctor") {
        role = "doctor"
    } else if cid.AssertAttributeValue("role", "patient") {
        role = "patient"
    }

    userID, initialBalance := args[0], args[1]
    balance, err := parseFloat(initialBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token initialization", false)))
    }

    tokenBalance := TokenBalance{
        UserID:         userID,
        Balance:        balance,
        BlockchainHash: generateHash(userID + initialBalance),
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }

    tokenJSON, err := json.Marshal(tokenBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token initialization", false)))
    }

    err = stub.PutState(userID, tokenJSON)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token initialization", false)))
    }

    return shim.Success([]byte(fmt.Sprintf("%s", getRoleMessage(role, "token initialization", true))))
}

// rewardPatient rewards a patient with tokens
func (t *PaymentChaincode) rewardPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    /**
     * Reward a patient with tokens with role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     *   args ([]string): Arguments [patientID, amount, reason]
     * 
     * Returns:
     *   pb.Response: Success or error response with role-specific message
     */
    if len(args) != 3 {
        return shim.Error("Please provide the patient ID, reward amount, and reason. Thank you!")
    }

    cid := ClientIdentity(stub)
    role := "admin" 
    if cid.AssertAttributeValue("role", "doctor") {
        role = "doctor"
    } else if cid.AssertAttributeValue("role", "patient") {
        role = "patient"
    }

    patientID, amount, reason := args[0], args[1], args[2]
    reward, err := parseFloat(amount)
    if err != nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the reward amount you entered isn’t valid. Please use a numeric value and try again or contact support: %v", role, err))
    }

    if !cid.AssertAttributeValue("role", "admin") && !cid.AssertAttributeValue("role", "doctor") {
        return shim.Error(fmt.Sprintf("Sorry, %s, only admins or doctors can issue rewards. Please log in with the correct role or contact support.", role))
    }

    patientBytes, err := stub.GetState(patientID)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "patient reward", false)))
    }
    if patientBytes == nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the patient %s doesn’t exist. Please verify the patient ID and try again or contact support.", role, patientID))
    }

    var tokenBalance TokenBalance
    err = json.Unmarshal(patientBytes, &tokenBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "patient reward", false)))
    }

    tokenBalance.Balance += reward
    tokenBalance.BlockchainHash = generateHash(tokenBalance.UserID + fmt.Sprintf("%f", tokenBalance.Balance))
    tokenBalance.UpdatedAt = time.Now()

    tokenJSON, err := json.Marshal(tokenBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "patient reward", false)))
    }

    err = stub.PutState(patientID, tokenJSON)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "patient reward", false)))
    }

    // Sync with Ethereum off-chain (simplified)
    err = syncWithEthereum(patientID, reward)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "Ethereum sync", false)))
    }

    return shim.Success([]byte(fmt.Sprintf("%s", getRoleMessage(role, "patient reward", true))))
}

// rewardDoctor 
func (t *PaymentChaincode) rewardDoctor(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    /**
     * Reward a doctor with tokens with role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     *   args ([]string): Arguments [doctorID, amount, reason]
     * 
     * Returns:
     *   pb.Response: Success or error response with role-specific message
     */
    if len(args) != 3 {
        return shim.Error("Please provide the doctor ID, reward amount, and reason. Thank you!")
    }

    cid := ClientIdentity(stub)
    role := "admin" 
    if cid.AssertAttributeValue("role", "doctor") {
        role = "doctor"
    } else if cid.AssertAttributeValue("role", "patient") {
        role = "patient"
    }

    doctorID, amount, reason := args[0], args[1], args[2]
    reward, err := parseFloat(amount)
    if err != nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the reward amount you entered isn’t valid. Please use a numeric value and try again or contact support: %v", role, err))
    }

    if !cid.AssertAttributeValue("role", "admin") {
        return shim.Error(fmt.Sprintf("Sorry, %s, only admins can issue rewards to doctors. Please log in with the correct role or contact support.", role))
    }

    doctorBytes, err := stub.GetState(doctorID)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "doctor reward", false)))
    }
    if doctorBytes == nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the doctor %s doesn’t exist. Please verify the doctor ID and try again or contact support.", role, doctorID))
    }

    var tokenBalance TokenBalance
    err = json.Unmarshal(doctorBytes, &tokenBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "doctor reward", false)))
    }

    tokenBalance.Balance += reward
    tokenBalance.BlockchainHash = generateHash(tokenBalance.UserID + fmt.Sprintf("%f", tokenBalance.Balance))
    tokenBalance.UpdatedAt = time.Now()

    tokenJSON, err := json.Marshal(tokenBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "doctor reward", false)))
    }

    err = stub.PutState(doctorID, tokenJSON)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "doctor reward", false)))
    }

    // Sync with Ethereum off-chain (simplified)
    err = syncWithEthereum(doctorID, reward)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "Ethereum sync", false)))
    }

    return shim.Success([]byte(fmt.Sprintf("%s", getRoleMessage(role, "doctor reward", true))))
}

// getBalance retrieves a user's token balance
func (t *PaymentChaincode) getBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    /**
     * Retrieve a user's token balance with role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     *   args ([]string): Arguments [userID]
     * 
     * Returns:
     *   pb.Response: Success or error response with role-specific message
     */
    if len(args) != 1 {
        return shim.Error("Please provide a user ID to retrieve balance. Thank you!")
    }

    cid := ClientIdentity(stub)
    role := "patient" // Default role
    if cid.AssertAttributeValue("role", "doctor") {
        role = "doctor"
    } else if cid.AssertAttributeValue("role", "admin") {
        role = "admin"
    }

    userID := args[0]
    balanceBytes, err := stub.GetState(userID)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "balance retrieval", false)))
    }
    if balanceBytes == nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the user %s doesn’t exist. Please verify the user ID and try again or contact support.", role, userID))
    }

    var tokenBalance TokenBalance
    err = json.Unmarshal(balanceBytes, &tokenBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "balance retrieval", false)))
    }

    return shim.Success([]byte(fmt.Sprintf("%s\n{\"user_id\":\"%s\",\"balance\":%f}", getRoleMessage(role, "balance retrieval", true), tokenBalance.UserID, tokenBalance.Balance)))
}

// transferTokens transfers tokens between users
func (t *PaymentChaincode) transferTokens(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    /**
     * Transfer tokens between users with role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     *   args ([]string): Arguments [fromID, toID, amount]
     * 
     * Returns:
     *   pb.Response: Success or error response with role-specific message
     */
    if len(args) != 3 {
        return shim.Error("Please provide from ID, to ID, and amount to transfer tokens. Thank you!")
    }

    cid := ClientIdentity(stub)
    role := "patient" // Default role
    if cid.AssertAttributeValue("role", "doctor") {
        role = "doctor"
    } else if cid.AssertAttributeValue("role", "admin") {
        role = "admin"
    }

    fromID, toID, amount := args[0], args[1], args[2]
    transfer, err := parseFloat(amount)
    if err != nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the transfer amount you entered isn’t valid. Please use a numeric value and try again or contact support: %v", role, err))
    }

    fromBytes, err := stub.GetState(fromID)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token transfer", false)))
    }
    if fromBytes == nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the sender %s doesn’t exist. Please verify the ID and try again or contact support.", role, fromID))
    }

    toBytes, err := stub.GetState(toID)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token transfer", false)))
    }
    if toBytes == nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the recipient %s doesn’t exist. Please verify the ID and try again or contact support.", role, toID))
    }

    var fromBalance, toBalance TokenBalance
    err = json.Unmarshal(fromBytes, &fromBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token transfer", false)))
    }
    err = json.Unmarshal(toBytes, &toBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token transfer", false)))
    }

    if fromBalance.Balance < transfer {
        return shim.Error(fmt.Sprintf("Sorry, %s, insufficient balance for transfer from %s. Please check the balance and try again or contact support.", role, fromID))
    }

    fromBalance.Balance -= transfer
    fromBalance.BlockchainHash = generateHash(fromBalance.UserID + fmt.Sprintf("%f", fromBalance.Balance))
    fromBalance.UpdatedAt = time.Now()

    toBalance.Balance += transfer
    toBalance.BlockchainHash = generateHash(toBalance.UserID + fmt.Sprintf("%f", toBalance.Balance))
    toBalance.UpdatedAt = time.Now()

    fromJSON, err := json.Marshal(fromBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token transfer", false)))
    }
    toJSON, err := json.Marshal(toBalance)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token transfer", false)))
    }

    err = stub.PutState(fromID, fromJSON)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token transfer", false)))
    }
    err = stub.PutState(toID, toJSON)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "token transfer", false)))
    }

    // Sync with Ethereum off-chain (simplified)
    err = syncWithEthereumTransfer(fromID, toID, transfer)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "Ethereum transfer sync", false)))
    }

    return shim.Success([]byte(fmt.Sprintf("%s", getRoleMessage(role, "token transfer", true))))
}

func parseFloat(s string) (float64, error) {
    /**
     * Parse a string to float64, handling errors gracefully.
     * 
     * Args:
     *   s (string): String to parse
     * 
     * Returns:
     *   float64: Parsed float, error if parsing fails
     */
    f, err := strconv.ParseFloat(s, 64)
    return f, err
}

func generateHash(data interface{}) string {
    /**
     * Generate a SHA-256 hash for blockchain data.
     * 
     * Args:
     *   data (interface{}): Input data to hash
     * 
     * Returns:
     *   string: Hex-encoded hash
     */
    dataBytes, _ := json.Marshal(data)
    hash := sha256.Sum256(dataBytes)
    return hex.EncodeToString(hash[:])
}

func syncWithEthereum(userID string, amount float64) error {
    /**
     * Sync token reward with Ethereum off-chain (simplified for demonstration).
     * 
     * Args:
     *   userID (string): User ID
     *   amount (float64): Amount to reward
     * 
     * Returns:
     *   error: Error if sync fails, nil otherwise
     */
    data := map[string]interface{}{
        "from":  "admin_wallet",
        "to":    userID,
        "value": amount,
    }
    dataJSON, _ := json.Marshal(data)
    resp, err := http.Post(fmt.Sprintf("%s/reward", ETH_URL), "application/json", bytes.NewBuffer(dataJSON))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}

func syncWithEthereumTransfer(fromID, toID string, amount float64) error {
    /**
     * Sync token transfer with Ethereum off-chain (simplified for demonstration).
     * 
     * Args:
     *   fromID (string): Sender ID
     *   toID (string): Recipient ID
     *   amount (float64): Amount to transfer
     * 
     * Returns:
     *   error: Error if sync fails, nil otherwise
     */
    data := map[string]interface{}{
        "from":  fromID,
        "to":    toID,
        "value": amount,
    }
    dataJSON, _ := json.Marshal(data)
    resp, err := http.Post(fmt.Sprintf("%s/transfer", ETH_URL), "application/json", bytes.NewBuffer(dataJSON))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}

// ClientIdentity for role checking
type ClientIdentity struct {
    stub shim.ChaincodeStubInterface
}

func ClientIdentity(stub shim.ChaincodeStubInterface) ClientIdentity {
    /**
     * Create a ClientIdentity instance for role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub
     * 
     * Returns:
     *   ClientIdentity: Client identity instance
     */
    return ClientIdentity{stub: stub}
}

func (ci ClientIdentity) AssertAttributeValue(attrName, attrValue string) bool {
    /**
     * Assert an attribute value for role checking.
     * 
     * Args:
     *   attrName (string): Attribute name (e.g., "role")
     *   attrValue (string): Expected attribute value (e.g., "admin")
     * 
     * Returns:
     *   bool: True if attribute matches, false otherwise
     */
    attrs, err := ci.stub.ReadCertAttributes(attrName)
    if err != nil || len(attrs) == 0 {
        return false
    }
    return attrs[0] == attrValue
}

func main() {
    /**
     * Main function to start the PaymentChaincode.
     */
    err := shim.Start(new(PaymentChaincode))
    if err != nil {
        fmt.Printf("Error starting PaymentChaincode: %s", err)
    }
}