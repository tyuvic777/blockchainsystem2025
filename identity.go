package main

import (
    "encoding/json"
    "fmt"
    "github.com/hyperledger/fabric-chaincode-go/shim"
    pb "github.com/hyperledger/fabric-protos-go/peer"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "net/http"
    "io"
    "time"
    "strings"
)

// IdentityChaincode represents the identity management chaincode
type IdentityChaincode struct {
}

// DID represents a decentralized identity
type DID struct {
    ID              string    `json:"id"`
    Owner           string    `json:"owner"`
    PublicKey       string    `json:"public_key"`
    Attributes      string    `json:"attributes"`
    BlockchainHash  string    `json:"blockchain_hash"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    Revoked         bool      `json:"revoked"`
}

// Init initializes the chaincode
func (t *IdentityChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
    /**
     * Initialize the identity chaincode, setting up any initial state if needed.
     * 
     * Returns:
     *   pb.Response: Success response with no payload
     */
    return shim.Success(nil)
}

// Invoke handles chaincode invocations
func (t *IdentityChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
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
    case "createDID":
        return t.createDID(stub, args)
    case "updateDID":
        return t.updateDID(stub, args)
    case "getDID":
        return t.getDID(stub, args)
    case "revokeDID":
        return t.revokeDID(stub, args)
    case "verifySignature":
        return t.verifySignature(stub, args)
    default:
        return shim.Error("Invalid function name. Please provide a valid function (createDID, updateDID, getDID, revokeDID, verifySignature). Thank you!")
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

// createDID creates a new decentralized identity
func (t *IdentityChaincode) createDID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    /**
     * Create a new decentralized identity (DID) with role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     *   args ([]string): Arguments [owner, publicKey, attributes]
     * 
     * Returns:
     *   pb.Response: Success or error response with role-specific message
     */
    if len(args) != 3 {
        return shim.Error("Please provide owner, public key, and attributes to create a DID. Thank you!")
    }

    cid := ClientIdentity(stub)
    role := "patient" // Default role, determine from MSP attribute
    if cid.AssertAttributeValue("role", "doctor") {
        role = "doctor"
    } else if cid.AssertAttributeValue("role", "admin") {
        role = "admin"
    }

    owner, publicKey, attributes := args[0], args[1], args[2]
    didID := fmt.Sprintf("did:mediNet:%s", generateUUID())

    // Generate ECC key pair for signature
    privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID creation", false)))
    }
    publicKeyBytes := elliptic.Marshal(elliptic.P256(), privateKey.PublicKey.X, privateKey.PublicKey.Y)
    publicKeyStr := hex.EncodeToString(publicKeyBytes)

    did := DID{
        ID:              didID,
        Owner:           owner,
        PublicKey:       publicKeyStr,
        Attributes:      attributes,
        BlockchainHash:  generateHash(owner + publicKeyStr + attributes),
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
        Revoked:         false,
    }

    didJSON, err := json.Marshal(did)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID creation", false)))
    }

    err = stub.PutState(didID, didJSON)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID creation", false)))
    }

    return shim.Success([]byte(fmt.Sprintf("%s", getRoleMessage(role, "DID creation", true))))
}

// updateDID updates an existing decentralized identity
func (t *IdentityChaincode) updateDID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    /**
     * Update an existing decentralized identity (DID) with role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     *   args ([]string): Arguments [didID, attributes, owner]
     * 
     * Returns:
     *   pb.Response: Success or error response with role-specific message
     */
    if len(args) != 3 {
        return shim.Error("Please provide DID ID, new attributes, and owner to update a DID. Thank you!")
    }

    cid := ClientIdentity(stub)
    role := "patient" // Default role
    if cid.AssertAttributeValue("role", "doctor") {
        role = "doctor"
    } else if cid.AssertAttributeValue("role", "admin") {
        role = "admin"
    }

    didID, attributes, owner := args[0], args[1], args[2]
    didBytes, err := stub.GetState(didID)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID update", false)))
    }
    if didBytes == nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the DID %s does not exist. Please verify the ID and try again or contact support.", role, didID))
    }

    var did DID
    err = json.Unmarshal(didBytes, &did)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID update", false)))
    }

    if did.Owner != owner && !cid.AssertAttributeValue("role", "admin") {
        return shim.Error(fmt.Sprintf("Sorry, %s, you don’t have permission to update this DID. Please log in as the owner or an admin, or contact support.", role))
    }

    did.Attributes = attributes
    did.BlockchainHash = generateHash(did.Owner + did.PublicKey + did.Attributes)
    did.UpdatedAt = time.Now()

    didJSON, err := json.Marshal(did)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID update", false)))
    }

    err = stub.PutState(didID, didJSON)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID update", false)))
    }

    return shim.Success([]byte(fmt.Sprintf("%s", getRoleMessage(role, "DID update", true))))
}

// getDID retrieves a decentralized identity
func (t *IdentityChaincode) getDID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    /**
     * Retrieve a decentralized identity (DID) with role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     *   args ([]string): Arguments [didID]
     * 
     * Returns:
     *   pb.Response: Success or error response with role-specific message
     */
    if len(args) != 1 {
        return shim.Error("Please provide a DID ID to retrieve. Thank you!")
    }

    cid := ClientIdentity(stub)
    role := "patient" // Default role
    if cid.AssertAttributeValue("role", "doctor") {
        role = "doctor"
    } else if cid.AssertAttributeValue("role", "admin") {
        role = "admin"
    }

    didID := args[0]
    didBytes, err := stub.GetState(didID)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID retrieval", false)))
    }
    if didBytes == nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the DID %s does not exist. Please verify the ID and try again or contact support.", role, didID))
    }

    return shim.Success([]byte(fmt.Sprintf("%s", getRoleMessage(role, "DID retrieval", true) + "\n" + string(didBytes))))
}

// revokeDID revokes a decentralized identity
func (t *IdentityChaincode) revokeDID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    /**
     * Revoke a decentralized identity (DID) with role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     *   args ([]string): Arguments [didID, owner]
     * 
     * Returns:
     *   pb.Response: Success or error response with role-specific message
     */
    if len(args) != 2 {
        return shim.Error("Please provide a DID ID and owner to revoke a DID. Thank you!")
    }

    cid := ClientIdentity(stub)
    role := "patient" // Default role
    if cid.AssertAttributeValue("role", "doctor") {
        role = "doctor"
    } else if cid.AssertAttributeValue("role", "admin") {
        role = "admin"
    }

    didID, owner := args[0], args[1]
    didBytes, err := stub.GetState(didID)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID revocation", false)))
    }
    if didBytes == nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the DID %s does not exist. Please verify the ID and try again or contact support.", role, didID))
    }

    var did DID
    err = json.Unmarshal(didBytes, &did)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID revocation", false)))
    }

    if did.Owner != owner && !cid.AssertAttributeValue("role", "admin") {
        return shim.Error(fmt.Sprintf("Sorry, %s, you don’t have permission to revoke this DID. Please log in as the owner or an admin, or contact support.", role))
    }

    did.Revoked = true
    did.UpdatedAt = time.Now()

    didJSON, err := json.Marshal(did)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID revocation", false)))
    }

    err = stub.PutState(didID, didJSON)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "DID revocation", false)))
    }

    return shim.Success([]byte(fmt.Sprintf("%s", getRoleMessage(role, "DID revocation", true))))
}

// verifySignature verifies a signature for a DID
func (t *IdentityChaincode) verifySignature(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    /**
     * Verify a digital signature for a decentralized identity (DID) with role-based access control.
     * 
     * Args:
     *   stub (shim.ChaincodeStubInterface): Fabric chaincode stub for state operations
     *   args ([]string): Arguments [didID, data, signature]
     * 
     * Returns:
     *   pb.Response: Success or error response with role-specific message
     */
    if len(args) != 3 {
        return shim.Error("Please provide a DID ID, data, and signature to verify. Thank you!")
    }

    cid := ClientIdentity(stub)
    role := "patient" // Default role
    if cid.AssertAttributeValue("role", "doctor") {
        role = "doctor"
    } else if cid.AssertAttributeValue("role", "admin") {
        role = "admin"
    }

    didID, data, signature := args[0], args[1], args[2]
    didBytes, err := stub.GetState(didID)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "signature verification", false)))
    }
    if didBytes == nil {
        return shim.Error(fmt.Sprintf("Sorry, %s, the DID %s does not exist. Please verify the ID and try again or contact support.", role, didID))
    }

    var did DID
    err = json.Unmarshal(didBytes, &did)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "signature verification", false)))
    }

    publicKeyBytes, err := hex.DecodeString(did.PublicKey)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "signature verification", false)))
    }

    x, y := elliptic.Unmarshal(elliptic.P256(), publicKeyBytes)
    publicKey := ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}

    signatureBytes, err := hex.DecodeString(signature)
    if err != nil {
        return shim.Error(fmt.Sprintf("%s", getRoleMessage(role, "signature verification", false)))
    }

    hash := sha256.Sum256([]byte(data))
    valid := ecdsa.Verify(&publicKey, hash[:], signatureBytes[:len(signatureBytes)-8], signatureBytes[len(signatureBytes)-8:])
    if !valid {
        return shim.Error(fmt.Sprintf("Sorry, %s, the signature is invalid for DID %s. Please verify the data and try again or contact support.", role, didID))
    }

    return shim.Success([]byte(fmt.Sprintf("%s", getRoleMessage(role, "signature verification", true))))
}

func generateUUID() string {
    /**
     * Generate a simple UUID for DIDs (simplified for demonstration).
     * 
     * Returns:
     *   string: Hex-encoded UUID
     */
    hash := sha256.Sum256([]byte(time.Now().String()))
    return hex.EncodeToString(hash[:])[:32]
}

func generateHash(data string) string {
    /**
     * Generate a SHA-256 hash for blockchain data.
     * 
     * Args:
     *   data (string): Input data to hash
     * 
     * Returns:
     *   string: Hex-encoded hash
     */
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
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
     * Main function to start the IdentityChaincode.
     */
    err := shim.Start(new(IdentityChaincode))
    if err != nil {
        fmt.Printf("Error starting IdentityChaincode: %s", err)
    }
}