package main

import (
    "encoding/json"
    "testing"
    "github.com/hyperledger/fabric-chaincode-go/shim"
    "github.com/hyperledger/fabric-chaincode-go/shimtest"
    pb "github.com/hyperledger/fabric-protos-go/peer"
    "time"
)

func setupTest(t *testing.T) (*shimtest.MockStub, *PatientCareChaincode) {
    chaincode := new(PatientCareChaincode)
    stub := shimtest.NewMockStub("patient_care_chaincode", chaincode)
    response := stub.MockInit("1", [][]byte{})
    if response.Status != shim.OK {
        t.Fatalf("Init failed: %s", response.Message)
    }
    return stub, chaincode
}

func TestCreateRecord(t *testing.T) {
    stub, _ := setupTest(t)
    for i := 0; i < 400; i++ {
        args := [][]byte{[]byte("createRecord"), []byte("user1"), []byte("data" + string(i)), []byte("type1")}
        response := stub.MockInvoke("1", args)
        if response.Status != shim.OK {
            t.Fatalf("createRecord failed: %s", response.Message)
        }
        var record PatientRecord
        err := json.Unmarshal(response.Payload, &record)
        if err != nil {
            t.Fatalf("Failed to unmarshal record: %s", err)
        }
        if record.UserID != "user1" || record.DataHash != generateHash("data"+string(i)) || record.Type != "type1" {
            t.Fatalf("Record does not match expected values")
        }
    }
}

func TestUpdateRecord(t *testing.T) {
    stub, _ := setupTest(t)
    args := [][]byte{[]byte("createRecord"), []byte("user1"), []byte("data1"), []byte("type1")}
    response := stub.MockInvoke("1", args)
    if response.Status != shim.OK {
        t.Fatalf("createRecord failed: %s", response.Message)
    }
    args = [][]byte{[]byte("updateRecord"), []byte("type1_user1_1"), []byte("newData")}
    response = stub.MockInvoke("2", args)
    if response.Status != shim.OK {
        t.Fatalf("updateRecord failed: %s", response.Message)
    }
    var record PatientRecord
    err := json.Unmarshal(response.Payload, &record)
    if err != nil {
        t.Fatalf("Failed to unmarshal record: %s", err)
    }
    if record.DataHash != generateHash("newData") {
        t.Fatalf("Record data hash does not match expected values")
    }
}

func TestGetRecord(t *testing.T) {
    stub, _ := setupTest(t)
    args := [][]byte{[]byte("createRecord"), []byte("user1"), []byte("data1"), []byte("type1")}
    response := stub.MockInvoke("1", args)
    if response.Status != shim.OK {
        t.Fatalf("createRecord failed: %s", response.Message)
    }
    args = [][]byte{[]byte("getRecord"), []byte("type1_user1_1")}
    response = stub.MockInvoke("2", args)
    if response.Status != shim.OK {
        t.Fatalf("getRecord failed: %s", response.Message)
    }
    var record PatientRecord
    err := json.Unmarshal(response.Payload, &record)
    if err != nil {
        t.Fatalf("Failed to unmarshal record: %s", err)
    }
    if record.UserID != "user1" || record.DataHash != generateHash("data1") || record.Type != "type1" {
        t.Fatalf("Record does not match expected values")
    }
}

func TestAddCarePlan(t *testing.T) {
    stub, _ := setupTest(t)
    args := [][]byte{[]byte("AddCarePlan"), []byte("user1"), []byte("carePlan1")}
    response := stub.MockInvoke("1", args)
    if response.Status != shim.OK {
        t.Fatalf("AddCarePlan failed: %s", response.Message)
    }
    var record PatientRecord
    err := json.Unmarshal(response.Payload, &record)
    if err != nil {
        t.Fatalf("Failed to unmarshal record: %s", err)
    }
    if record.UserID != "user1" || record.DataHash != generateHash("carePlan1") || record.Type != "careplan" {
        t.Fatalf("Record does not match expected values")
    }
}

func TestAddPrescription(t *testing.T) {
    stub, _ := setupTest(t)
    args := [][]byte{[]byte("AddPrescription"), []byte("user1"), []byte("prescription1")}
    response := stub.MockInvoke("1", args)
    if response.Status != shim.OK {
        t.Fatalf("AddPrescription failed: %s", response.Message)
    }
    var record PatientRecord
    err := json.Unmarshal(response.Payload, &record)
    if err != nil {
        t.Fatalf("Failed to unmarshal record: %s", err)
    }
    if record.UserID != "user1" || record.DataHash != generateHash("prescription1") || record.Type != "prescription" {
        t.Fatalf("Record does not match expected values")
    }
}

func TestGetAppointments(t *testing.T) {
    stub, _ := setupTest(t)
    args := [][]byte{[]byte("createRecord"), []byte("user1"), []byte("appointment1"), []byte("appointment")}
    response := stub.MockInvoke("1", args)
    if response.Status != shim.OK {
        t.Fatalf("createRecord failed: %s", response.Message)
    }
    args = [][]byte{[]byte("GetAppointments"), []byte("user1")}
    response = stub.MockInvoke("2", args)
    if response.Status != shim.OK {
        t.Fatalf("GetAppointments failed: %s", response.Message)
    }
    var records []PatientRecord
    err := json.Unmarshal(response.Payload, &records)
    if err != nil {
        t.Fatalf("Failed to unmarshal records: %s", err)
    }
    if len(records) == 0 || records[0].UserID != "user1" || records[0].DataHash != generateHash("appointment1") || records[0].Type != "appointment" {
        t.Fatalf("Records do not match expected values")
    }
}

func TestGetAnalytics(t *testing.T) {
    stub, _ := setupTest(t)
    args := [][]byte{[]byte("GetAnalytics"), []byte("user1")}
    response := stub.MockInvoke("1", args)
    if response.Status != shim.OK {
        t.Fatalf("GetAnalytics failed: %s", response.Message)
    }
    var analytics map[string]int
    err := json.Unmarshal(response.Payload, &analytics)
    if err != nil {
        t.Fatalf("Failed to unmarshal analytics: %s", err)
    }
    if analytics["records"] != 50 || analytics["appointments"] != 10 || analytics["health_checks"] != 30 {
        t.Fatalf("Analytics do not match expected values")
    }
}

func TestGetPatientData(t *testing.T) {
    stub, _ := setupTest(t)
    args := [][]byte{[]byte("GetPatientData"), []byte("user1")}
    response := stub.MockInvoke("1", args)
    if response.Status != shim.OK {
        t.Fatalf("GetPatientData failed: %s", response.Message)
    }
    var patientData map[string]string
    err := json.Unmarshal(response.Payload, &patientData)
    if err != nil {
        t.Fatalf("Failed to unmarshal patient data: %s", err)
    }
    if patientData["id"] != "user1" || patientData["name"] != "John Doe" {
        t.Fatalf("Patient data does not match expected values")
    }
}

func TestAnonymizeRecord(t *testing.T) {
    stub, _ := setupTest(t)
    args := [][]byte{[]byte("AnonymizeRecord"), []byte("user1")}
    response := stub.MockInvoke("1", args)
    if response.Status != shim.OK {
        t.Fatalf("AnonymizeRecord failed: %s", response.Message)
    }
    if string(response.Payload) != "Data anonymized" {
        t.Fatalf("AnonymizeRecord response does not match expected value")
    }
}

func TestDeleteRecord(t *testing.T) {
    stub, _ := setupTest(t)
    args := [][]byte{[]byte("DeleteRecord"), []byte("user1")}
    response := stub.MockInvoke("1", args)
    if response.Status != shim.OK {
        t.Fatalf("DeleteRecord failed: %s", response.Message)
    }
    if string(response.Payload) != "Data deleted" {
        t.Fatalf("DeleteRecord response does not match expected value")
    }
}
