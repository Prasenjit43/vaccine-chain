/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Prasenjit43/vaccinechainhelper"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Asset and Token
type SmartContract struct {
	contractapi.Contract
}

type VaccineChainAdmin struct {
	Id            string `json:"id"`
	FirstName     string `json:"firstName"`
	MiddleName    string `json:"middleName,omitempty"`
	LastName      string `json:"lastName,omitempty"`
	AdminIdentity string `json:"adminIdentity"`
	ContactNo     string `json:"contactNo"`
	DocType       string `json:"docType"`
	EmailId       string `json:"emailId"`
	Suspended     bool   `json:"suspended"`
}

type Entity struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	LicenseNo     string `json:"licenseNo"`
	Address       string `json:"address"`       //updatable
	OwnerName     string `json:"ownerName"`     //updatable
	OwnerIdentity string `json:"ownerIdentity"` //updatable
	OwnerAddress  string `json:"ownerAddress"`  //updatable
	ContactNo     string `json:"contactNo"`     //updatable
	EmailId       string `json:"emailId"`       //updatable
	Suspended     bool   `json:"suspended"`
	BatchCount    int    `json:"batchCount,omitempty"`
	DocType       string `json:"docType"`
}

type Product struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Desc           string `json:"desc"`
	Type           string `json:"type"`
	Price          int16  `json:"price"`
	CartonCapacity int16  `json:"cartonCapacity"`
	PacketCapacity int16  `json:"packetCapacity"`
	DocType        string `json:"docType"`
	Suspended      bool   `json:"suspended"`
	Owner          string `json:"owner"`
}

type Batch struct {
	Id                string `json:"id"`
	Owner             string `json:"owner"`
	ProductId         string `json:"productId"`
	ManufacturingDate string `json:"manufacturingDate"`
	ExpiryDate        string `json:"expiryDate"`
	CartonQnty        int16  `json:"cartonQnty"`
}

type Asset struct {
	Id       string `json:"id"`
	BatchId  string `json:"batchId"`
	CartonId string `json:"cartonId"`
	//PacketId          string `json:"packetId"`
	Owner             string `json:"owner"`
	Status            string `json:"status"`
	ProductId         string `json:"productId"`
	ManufacturerId    string `json:"manufacturerId"`
	ManufacturingDate string `json:"manufacturingDate"`
	ExpiryDate        string `json:"expiryDate"`
	DocType           string `json:"docType"`
}

type Receipt struct {
	Id            string `json:"id"`
	CartonId      string `json:"cartonId"`
	DocType       string `json:"docType"`
	SupplierId    string `json:"supplierId"`
	VendorId      string `json:"vendorId"`
	ProductId     string `json:"productId"`
	ShippmentDate string `json:"shippmentDate"`
	BillAmount    int16  `json:"billAmount"`
}

// VaccineChainAdmin function adds a new admin to the vaccine chain system.
// It takes in a JSON string containing admin details and performs several validations before inserting the admin record into the ledger.
func (s *SmartContract) VaccineChainAdmin(ctx contractapi.TransactionContextInterface, adminInputString string) error {
	// Define a struct to hold the admin details
	var vaccineChainAdmin VaccineChainAdmin

	// Unmarshal the input JSON string into the VaccineChainAdmin struct
	err := json.Unmarshal([]byte(adminInputString), &vaccineChainAdmin)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for Admin: %v", err.Error())
	}
	fmt.Println("Input String :", vaccineChainAdmin)

	// Validate the identity of the caller as the super admin
	superAdminIdentity, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("superAdminIdentity :", superAdminIdentity)
	if superAdminIdentity != vaccinechainhelper.SUPER_ADMIN {
		return fmt.Errorf("permission denied: only super admin can call this function")
	}

	/* Validate Admin Id */

	// Check if the admin record already exists using its ID and document type
	objectBytes, err := vaccinechainhelper.IsExist(ctx, vaccineChainAdmin.Id, vaccineChainAdmin.DocType)
	if err != nil {
		return err
	}

	// If the record already exists, return an error
	if objectBytes != nil {
		return fmt.Errorf("Record already exists for %v with Id: %v", vaccineChainAdmin.DocType, vaccineChainAdmin.Id)
	}

	// Marshal the admin record into JSON format
	vaccineChainAdminJSON, err := json.Marshal(vaccineChainAdmin)
	if err != nil {
		return fmt.Errorf("Failed to marshal Admin record: %v", err.Error())
	}

	// Create a composite key using the admin's ID and document type
	compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{vaccineChainAdmin.Id, vaccineChainAdmin.DocType})
	if err != nil {
		return fmt.Errorf("Failed to create composite key for hospital %v: %v", vaccineChainAdmin.Id, err.Error())
	}

	// Put the admin record into the ledger using the composite key
	err = ctx.GetStub().PutState(compositeKey, vaccineChainAdminJSON)
	if err != nil {
		return fmt.Errorf("Failed to insert hospital details to CouchDB: %v", err.Error())
	}

	fmt.Println("********** End of Add Vaccine Chain Admin Function ******************")
	return nil
}

// AddEntity function adds a new entity to the smart contract ledger.
func (s *SmartContract) AddEntity(ctx contractapi.TransactionContextInterface, entityInputString string) error {
	// Unmarshal input string into Entity struct
	var entityInput Entity
	err := json.Unmarshal([]byte(entityInputString), &entityInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal input string for Entity: %v", err.Error())
	}
	fmt.Println("Input String:", entityInput)

	// Fetch certificate attributes for the logged-in user
	attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	if err != nil {
		return err
	}
	fmt.Println("userRole:", attributes["userRole"])

	// Check if the user has required permissions (Vaccine Chain Admin) to register the entity
	if attributes["userRole"] != vaccinechainhelper.VACCINE_CHAIN_ADMIN {
		return fmt.Errorf("only Vaccine Chain Admin are allowed to register %v", entityInput.DocType)
	}

	// Check if the entity already exists
	objectBytes, err := vaccinechainhelper.IsExist(ctx, entityInput.Id, entityInput.DocType)
	if err != nil {
		return err
	}
	if objectBytes != nil {
		return fmt.Errorf("record already exists for %v with Id: %v", entityInput.DocType, entityInput.Id)
	}

	// Marshal the entity into JSON
	entityJSON, err := json.Marshal(entityInput)
	if err != nil {
		return fmt.Errorf("failed to marshal entity records: %v", err.Error())
	}

	// Create composite key for the entity
	compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{entityInput.Id, entityInput.DocType})
	if err != nil {
		return fmt.Errorf("failed to create composite key for %v, error: %v", entityInput.Id, err.Error())
	}

	// Put the entity details into CouchDB
	err = ctx.GetStub().PutState(compositeKey, entityJSON)
	if err != nil {
		return fmt.Errorf("failed to insert entity details to CouchDB: %v", err.Error())
	}
	fmt.Println("********** End of Add Entity Function ******************")
	return nil
}

// AddProduct adds a new product to the blockchain ledger.
func (s *SmartContract) AddProduct(ctx contractapi.TransactionContextInterface, productInputString string) error {
	// Unmarshal the input string into a Product struct
	var productInput Product
	err := json.Unmarshal([]byte(productInputString), &productInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for Product: %v", err.Error())
	}
	fmt.Println("Input String :", productInput)

	// Fetch certificate attributes for the logged-in user
	attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	if err != nil {
		return err
	}
	fmt.Println("userRole :", attributes["userRole"])

	// Check if the user role is that of a manufacturer
	if attributes["userRole"] != vaccinechainhelper.MANUFACTURER {
		return fmt.Errorf("Only Manufacturers are allowed to register Product details")
	}

	// Fetch the manufacturing ID of the logged-in user
	manufacturerId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("manufacturerId :", manufacturerId)
	if err != nil {
		return fmt.Errorf("Failed to get Manufacturer ID")
	}

	// Generate a unique ID for the product using manufacturer ID and product ID
	productId := productInput.Id + manufacturerId

	// Check if the product already exists for the manufacturer
	objectBytes, err := vaccinechainhelper.IsExist(ctx, productId, productInput.DocType)
	if err != nil {
		return err
	}
	if objectBytes != nil {
		return fmt.Errorf("Record already exists with ID: %v", productInput.Id)
	}

	// Assign the manufacturer as the owner of the product
	productInput.Owner = manufacturerId

	// Marshal the product record into JSON
	productJSON, err := json.Marshal(productInput)
	if err != nil {
		return fmt.Errorf("Failed to marshal product records: %v", err.Error())
	}

	// Create a composite key using product ID and document type
	compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{productId, productInput.DocType})
	if err != nil {
		return fmt.Errorf("Failed to create composite key for %v: %v", productInput.Id, err.Error())
	}

	// Put the product details into the ledger's state
	err = ctx.GetStub().PutState(compositeKey, productJSON)
	if err != nil {
		return fmt.Errorf("Failed to insert product details into CouchDB: %v", err.Error())
	}

	fmt.Println("********** End of Add Product Function ******************")
	return nil
}

func (s *SmartContract) AddBatch(ctx contractapi.TransactionContextInterface, batchInputString string) error {
	// Unmarshal input string into Batch struct
	var batchInput Batch
	err := json.Unmarshal([]byte(batchInputString), &batchInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for Batch: %v", err.Error())
	}

	// Print input for debugging
	fmt.Println("Input String :", batchInput)

	// Fetch certificate attributes for the logged-in entity
	attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	if err != nil {
		return err
	}
	fmt.Println("userRole :", attributes["userRole"])

	// Check the role for permission
	if attributes["userRole"] != vaccinechainhelper.MANUFACTURER {
		return fmt.Errorf("Only Manufacturer is allowed to add a batch")
	}

	// Fetch manufacturing ID
	manufacturerId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("manufacturerId :", manufacturerId)
	if err != nil {
		return fmt.Errorf("Failed to get Manufacturer ID")
	}

	// Fetch Manufacturing details
	var manufacturerDetails Entity
	objectBytes, err := vaccinechainhelper.IsExist(ctx, manufacturerId, vaccinechainhelper.MANUFACTURER)
	if err != nil {
		return err
	}
	if objectBytes == nil {
		return fmt.Errorf("Record does not exist with ID: %v", manufacturerId)
	}
	err = json.Unmarshal(objectBytes.([]byte), &manufacturerDetails)
	fmt.Println("manufacturerDetails :", manufacturerDetails)

	// Fetch Product details
	var productDetails Product
	productId := batchInput.ProductId + manufacturerId
	objectBytes, err = vaccinechainhelper.IsExist(ctx, productId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}
	if objectBytes == nil {
		return fmt.Errorf("Record does not exist with ID: %v", manufacturerId)
	}
	err = json.Unmarshal(objectBytes.([]byte), &productDetails)
	fmt.Println("productDetails :", productDetails)

	// // Update batch number
	// fmt.Println("old batch no  :", manufacturerDetails.BatchCount)
	// newBatchNo := manufacturerDetails.BatchCount + 1
	// fmt.Println("newBatchNo :", newBatchNo)

	// Insert Batch Details
	// batchInput.Id = "B" + strconv.Itoa(int(newBatchNo))
	// fmt.Println("batchInput.Id :", batchInput.Id)

	batchInput.Id = "B" + strconv.Itoa(manufacturerDetails.BatchCount)
	fmt.Println("batchInput.Id :", batchInput.Id)

	// Marshal batch details into JSON
	batchJSON, err := json.Marshal(batchInput)
	if err != nil {
		return fmt.Errorf("Failed to marshal batch records: %v", err.Error())
	}

	// Create composite key and insert batch details to couchDB
	compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{manufacturerId, batchInput.Id})
	if err != nil {
		return fmt.Errorf("Failed to create composite key for %v: %v", batchInput.Id, err.Error())
	}
	err = ctx.GetStub().PutState(compositeKey, batchJSON)
	if err != nil {
		return fmt.Errorf("Failed to insert batch details to couchDB: %v", err.Error())
	}

	// Insert Asset Records for each item in the batch
	var i, j int16
	var assetId, cartonId, packetId string

	for i = 1; i <= batchInput.CartonQnty; i++ {
		for j = 1; j <= productDetails.CartonCapacity; j++ {
			// Generate asset IDs for each item in the batch
			cartonId = batchInput.Id + "_" + "C" + strconv.Itoa(int(i))
			packetId = "P" + strconv.Itoa(int(j))
			assetId = manufacturerId + "_" + cartonId + "_" + packetId
			fmt.Println("assetId :", assetId)

			// Create Asset object
			asset := Asset{
				Id:       assetId,
				BatchId:  batchInput.Id,
				CartonId: cartonId,
				// PacketId:          packetId,
				Owner:             manufacturerId,
				Status:            vaccinechainhelper.Statuses.ReadyForDistribution,
				ProductId:         productDetails.Id,
				ManufacturerId:    manufacturerId,
				ManufacturingDate: batchInput.ManufacturingDate,
				ExpiryDate:        batchInput.ExpiryDate,
				DocType:           vaccinechainhelper.ASSET,
			}

			// Marshal asset details into JSON
			assetJSON, err := json.Marshal(asset)
			if err != nil {
				return fmt.Errorf("Failed to marshal asset records: %v", err.Error())
			}

			// Create composite key and insert asset details to couchDB
			// compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{manufacturerId, assetId})
			// if err != nil {
			// 	return fmt.Errorf("Failed to create composite key for %v: %v", batchInput.Id, err.Error())
			// }
			// err = ctx.GetStub().PutState(compositeKey, assetJSON)
			err = ctx.GetStub().PutState(assetId, assetJSON)
			if err != nil {
				return fmt.Errorf("Failed to insert asset details to couchDB: %v", err.Error())
			}
		}
	}

	// Update manufacturer Details with new batch count
	newBatchNo := manufacturerDetails.BatchCount + 1
	manufacturerDetails.BatchCount = newBatchNo
	manufacturerJSON, err := json.Marshal(manufacturerDetails)
	if err != nil {
		return fmt.Errorf("Failed to marshal manufacturer records: %v", err.Error())
	}

	// Update manufacturer details in couchDB
	compositeKey, err = ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{manufacturerId, vaccinechainhelper.MANUFACTURER})
	if err != nil {
		return fmt.Errorf("Failed to create composite key for %v: %v", batchInput.Id, err.Error())
	}
	err = ctx.GetStub().PutState(compositeKey, manufacturerJSON)
	if err != nil {
		return fmt.Errorf("Failed to update batch details in couchDB: %v", err.Error())
	}

	return nil
}

func (s *SmartContract) ShipToDistributer(ctx contractapi.TransactionContextInterface, distributionInputString string) error {
	distributionInput := struct {
		VendorId            string `json:"vendorId"`
		CartonId            string `json:"cartonId"`
		ShippmentDate       string `json:"shippmentDate"`
		PerUnitSellingPrice int16  `json:"perUnitSellingPrice"`
	}{}

	err := json.Unmarshal([]byte(distributionInputString), &distributionInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for distribution: %v", err.Error())
	}

	// Print input for debugging
	fmt.Println("Input String :", distributionInput)

	// Fetch entity ID for logged-in user
	entityId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("entityId :", entityId)
	if err != nil {
		return fmt.Errorf("Failed to get Entity ID")
	}

	//Validate VendorId
	vendorBytes, err := vaccinechainhelper.IsExist(ctx, distributionInput.VendorId, vaccinechainhelper.DISTRIBUTER)
	if err != nil {
		return err
	}
	if vendorBytes == nil {
		return fmt.Errorf("Record does not exist with ID: %v", distributionInput.VendorId)
	}

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","cartonId":"%s"}}`, entityId, distributionInput.CartonId)
	fmt.Println("queryString : ", queryString)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return err
	}
	defer resultsIterator.Close()

	// Check if there are no records in the iterator
	if !resultsIterator.HasNext() {
		fmt.Println("No records found")
		return fmt.Errorf("No Carton is ready for shipment for carton : %v", distributionInput.CartonId)
	}

	var productId, manufacturerId string
	var totalCarton int16 = 0
	for resultsIterator.HasNext() {
		responseRange, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		fmt.Println("Key : ", responseRange.Key)

		var asset Asset
		assetBytes, err := ctx.GetStub().GetState(responseRange.Key)
		if err != nil {
			return fmt.Errorf("failed to get asset %s: %v", responseRange.Key, err)
		}
		err = json.Unmarshal(assetBytes, &asset)
		if err != nil {
			return err
		}

		productId = asset.ProductId
		manufacturerId = asset.ManufacturerId

		asset.Owner = distributionInput.VendorId
		asset.Status = vaccinechainhelper.Statuses.ReceivedAtDistributor
		assetBytes, err = json.Marshal(asset)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(responseRange.Key, assetBytes)
		if err != nil {
			return fmt.Errorf("Shippment failed for asset %s: %v", asset.Id, err)
		}
		totalCarton++
	}

	var productDetails Product
	tempProductId := productId + manufacturerId
	productBytes, err := vaccinechainhelper.IsExist(ctx, tempProductId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}

	err = json.Unmarshal(productBytes.([]byte), &productDetails)
	fmt.Println("productDetails :", productDetails)

	//Creating Receipt
	txID := ctx.GetStub().GetTxID()
	billAmount := distributionInput.PerUnitSellingPrice * productDetails.PacketCapacity * totalCarton
	receipt := Receipt{
		Id:            txID,
		CartonId:      distributionInput.CartonId,
		DocType:       vaccinechainhelper.RECEIPT,
		SupplierId:    entityId,
		VendorId:      distributionInput.VendorId,
		ProductId:     productId,
		ShippmentDate: distributionInput.ShippmentDate,
		BillAmount:    billAmount,
	}

	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("Failed to marshal receipt records: %v", err.Error())
	}

	err = ctx.GetStub().PutState(txID, receiptJSON)
	if err != nil {
		return fmt.Errorf("Failed to update receipt details in couchDB: %v", err.Error())
	}

	return nil
}

func (s *SmartContract) ShipToChemist(ctx contractapi.TransactionContextInterface, distributionInputString string) error {
	distributionInput := struct {
		VendorId            string `json:"vendorId"`
		PacketId            string `json:"packetId"`
		ShippmentDate       string `json:"shippmentDate"`
		PerUnitSellingPrice int16  `json:"perUnitSellingPrice"`
	}{}

	err := json.Unmarshal([]byte(distributionInputString), &distributionInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for distribution: %v", err.Error())
	}

	// Print input for debugging
	fmt.Println("Input String :", distributionInput)

	// Fetch entity ID for logged-in user
	entityId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("entityId :", entityId)
	if err != nil {
		return fmt.Errorf("Failed to get Entity ID")
	}

	//Validate VendorId
	vendorBytes, err := vaccinechainhelper.IsExist(ctx, distributionInput.VendorId, vaccinechainhelper.CHEMIST)
	if err != nil {
		return err
	}
	if vendorBytes == nil {
		return fmt.Errorf("Record does not exist with ID: %v", distributionInput.VendorId)
	}

	// queryString := fmt.Sprintf(`{"selector":{"owner":"%s","cartonId":"%s"}}`, entityId, distributionInput.CartonId)
	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","id":"%s"}}`, entityId, distributionInput.PacketId)
	fmt.Println("queryString : ", queryString)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return err
	}
	defer resultsIterator.Close()

	// Check if there are no records in the iterator
	if !resultsIterator.HasNext() {
		fmt.Println("No records found")
		return fmt.Errorf("No Packet is ready for shipment for carton : %v", distributionInput.PacketId)
	}

	var productId, manufacturerId string
	// var totalCarton int16 = 0
	for resultsIterator.HasNext() {
		responseRange, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		fmt.Println("Key : ", responseRange.Key)

		var asset Asset
		assetBytes, err := ctx.GetStub().GetState(responseRange.Key)
		if err != nil {
			return fmt.Errorf("failed to get asset %s: %v", responseRange.Key, err)
		}
		err = json.Unmarshal(assetBytes, &asset)
		if err != nil {
			return err
		}

		productId = asset.ProductId
		manufacturerId = asset.ManufacturerId

		asset.Owner = distributionInput.VendorId
		asset.Status = vaccinechainhelper.Statuses.ChemistInventoryReceived
		assetBytes, err = json.Marshal(asset)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(responseRange.Key, assetBytes)
		if err != nil {
			return fmt.Errorf("Shippment failed for asset %s: %v", asset.Id, err)
		}
		// totalCarton++
	}

	var productDetails Product
	tempProductId := productId + manufacturerId
	productBytes, err := vaccinechainhelper.IsExist(ctx, tempProductId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}

	err = json.Unmarshal(productBytes.([]byte), &productDetails)
	fmt.Println("productDetails :", productDetails)

	//Creating Receipt
	txID := ctx.GetStub().GetTxID()
	billAmount := distributionInput.PerUnitSellingPrice * productDetails.PacketCapacity
	receipt := Receipt{
		Id:            txID,
		CartonId:      distributionInput.PacketId,
		DocType:       vaccinechainhelper.RECEIPT,
		SupplierId:    entityId,
		VendorId:      distributionInput.VendorId,
		ProductId:     productId,
		ShippmentDate: distributionInput.ShippmentDate,
		BillAmount:    billAmount,
	}

	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("Failed to marshal receipt records: %v", err.Error())
	}

	err = ctx.GetStub().PutState(txID, receiptJSON)
	if err != nil {
		return fmt.Errorf("Failed to update receipt details in couchDB: %v", err.Error())
	}

	return nil
}

func (s *SmartContract) ShipToCustomer(ctx contractapi.TransactionContextInterface, distributionInputString string) error {
	distributionInput := struct {
		CustomerId      string `json:"customerId"`
		PacketId        string `json:"packetId"`
		TransactionDate string `json:"transactionDate"`
	}{}

	err := json.Unmarshal([]byte(distributionInputString), &distributionInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for distribution: %v", err.Error())
	}

	// Print input for debugging
	fmt.Println("Input String :", distributionInput)

	// Fetch entity ID for logged-in user
	entityId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("entityId :", entityId)
	if err != nil {
		return fmt.Errorf("Failed to get Entity ID")
	}

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","id":"%s"}}`, entityId, distributionInput.PacketId)
	fmt.Println("queryString : ", queryString)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return err
	}
	defer resultsIterator.Close()

	// Check if there are no records in the iterator
	if !resultsIterator.HasNext() {
		fmt.Println("No records found")
		return fmt.Errorf("No Packet is ready for shipment for carton : %v", distributionInput.PacketId)
	}

	var productId, manufacturerId string
	// var totalCarton int16 = 0
	for resultsIterator.HasNext() {
		responseRange, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		fmt.Println("Key : ", responseRange.Key)

		var asset Asset
		assetBytes, err := ctx.GetStub().GetState(responseRange.Key)
		if err != nil {
			return fmt.Errorf("failed to get asset %s: %v", responseRange.Key, err)
		}
		err = json.Unmarshal(assetBytes, &asset)
		if err != nil {
			return err
		}

		productId = asset.ProductId
		manufacturerId = asset.ManufacturerId

		asset.Owner = distributionInput.CustomerId
		asset.Status = vaccinechainhelper.Statuses.SoldToCustomer
		assetBytes, err = json.Marshal(asset)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(responseRange.Key, assetBytes)
		if err != nil {
			return fmt.Errorf("Shippment failed for asset %s: %v", asset.Id, err)
		}
	}

	var productDetails Product
	tempProductId := productId + manufacturerId
	productBytes, err := vaccinechainhelper.IsExist(ctx, tempProductId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}

	err = json.Unmarshal(productBytes.([]byte), &productDetails)
	fmt.Println("productDetails :", productDetails)

	//Creating Receipt
	txID := ctx.GetStub().GetTxID()
	billAmount := productDetails.Price * productDetails.PacketCapacity
	receipt := Receipt{
		Id:            txID,
		CartonId:      distributionInput.PacketId,
		DocType:       vaccinechainhelper.RECEIPT,
		SupplierId:    entityId,
		VendorId:      distributionInput.CustomerId,
		ProductId:     productId,
		ShippmentDate: distributionInput.TransactionDate,
		BillAmount:    billAmount,
	}

	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("Failed to marshal receipt records: %v", err.Error())
	}

	err = ctx.GetStub().PutState(txID, receiptJSON)
	if err != nil {
		return fmt.Errorf("Failed to update receipt details in couchDB: %v", err.Error())
	}

	return nil
}

func (s *SmartContract) GetProductsByManufacturer(ctx contractapi.TransactionContextInterface) (string, error) {

	//fetching manufacturing Id
	manufacturerId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("manufacturerId :", manufacturerId)
	if err != nil {
		return "", fmt.Errorf("Failed to get Manufacturer Id")
	}

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","docType":"ITEM"}}`, manufacturerId)
	fmt.Println("queryString : ", queryString)

	productsByManufacturer, err := getQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return "", err
	}
	fmt.Println("productsByManufacturer : ", productsByManufacturer)

	return productsByManufacturer, nil
}

func (s *SmartContract) GetAssetByEntity(ctx contractapi.TransactionContextInterface) (string, error) {

	//fetching entity Id
	entityId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("entityId :", entityId)
	if err != nil {
		return "", fmt.Errorf("Failed to get entity Id")
	}

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","docType":"ASSET"}}`, entityId)
	fmt.Println("queryString : ", queryString)

	currectAssetsByEntity, err := getQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return "", err
	}
	fmt.Println("currectAssetsByEntity : ", currectAssetsByEntity)

	return currectAssetsByEntity, nil
}

func (s *SmartContract) ViewProfileDetails(ctx contractapi.TransactionContextInterface) (interface{}, error) {

	// entityIdentity, err := vaccinechainhelper.GetUserIdentityName(ctx)
	// fmt.Println("entityIdentity :", entityIdentity)
	// if err != nil {
	// 	return nil, err
	// }

	// attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("userRole for entityIdentity :", attributes["userRole"])

	// //validating User Id
	// entityDetailer, err := vaccinechainhelper.IsExist(ctx, entityIdentity, attributes["userRole"])
	// if err != nil {
	// 	return nil, err
	// }
	// if entityDetailer == nil {
	// 	return nil, fmt.Errorf("Record for %v user does not exist", entityIdentity)
	// }

	// // userData, ok := userDetailer.(User)
	// // if !ok {
	// // 	return nil, fmt.Errorf("Failed to convert Detailer to User type")
	// // }

	profileDetailsAsObject, err := getProfileDetails(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println("profileDetailsAsObject :", profileDetailsAsObject)
	// fmt.Println("profileDetailsAsObject entity :", profileDetailsAsObject.(Entity))

	fmt.Println("********** End of ViewProfileDetails Function ******************")
	return string(profileDetailsAsObject.([]byte)), nil
}

func (s *SmartContract) UpdateProfile(ctx contractapi.TransactionContextInterface, updateProfileInputString string) error {
	fmt.Println("stringInput:", updateProfileInputString)
	var entityUpdateInput Entity
	err := json.Unmarshal([]byte(updateProfileInputString), &entityUpdateInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal of input string for Entity: %v", err.Error())
	}
	fmt.Println("Input String :", entityUpdateInput)

	entityBytes, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}
	fmt.Println("entityBytes:", entityBytes)

	// entityDetails, ok := (entityBytes.(interface{})).(Entity)
	// if !ok {
	// 	return fmt.Errorf("Failed to convert Detailer to Entity type")
	// }

	var entityDetails Entity
	err = json.Unmarshal(entityBytes.([]byte), &entityDetails)
	if err != nil {
		return fmt.Errorf("Failed to convert Detailer to User type")
	}
	fmt.Println("entityDetails:", entityDetails)

	if entityUpdateInput.Address != "" {
		entityDetails.Address = entityUpdateInput.Address
	}
	if entityUpdateInput.OwnerName != "" {
		entityDetails.OwnerName = entityUpdateInput.OwnerName
	}
	if entityUpdateInput.OwnerIdentity != "" {
		entityDetails.OwnerIdentity = entityUpdateInput.OwnerIdentity
	}
	if entityUpdateInput.OwnerAddress != "" {
		entityDetails.OwnerAddress = entityUpdateInput.OwnerAddress
	}
	if entityUpdateInput.ContactNo != "" {
		entityDetails.ContactNo = entityUpdateInput.ContactNo
	}
	if entityUpdateInput.EmailId != "" {
		entityDetails.EmailId = entityUpdateInput.EmailId
	}

	//Updating Entity record
	entityJSON, err := json.Marshal(entityDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal of entity records : %v", err.Error())
	}

	compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{entityDetails.Id, entityDetails.DocType})
	if err != nil {
		return fmt.Errorf("failed to create composite key for %v and err is :%v", entityDetails.Id, err.Error())
	}
	err = ctx.GetStub().PutState(compositeKey, entityJSON)
	if err != nil {
		return fmt.Errorf("failed to insert user details to couchDB : %v", err.Error())
	}

	fmt.Println("********** End of Update Details Function ******************")

	return nil
}

func getProfileDetails(ctx contractapi.TransactionContextInterface) (interface{}, error) {

	//getting logged-in entity username
	entityIdentity, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("entityIdentity :", entityIdentity)
	if err != nil {
		return nil, err
	}

	//getting logged-in entity attributes
	attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	if err != nil {
		return nil, err
	}
	fmt.Println("userRole for entityIdentity :", attributes["userRole"])

	//validating User Id
	entityDetailer, err := vaccinechainhelper.IsExist(ctx, entityIdentity, attributes["userRole"])
	if err != nil {
		return nil, err
	}
	if entityDetailer == nil {
		return nil, fmt.Errorf("Record for %v user does not exist", entityIdentity)
	}

	// userData, ok := userDetailer.(User)
	// if !ok {
	// 	return nil, fmt.Errorf("Failed to convert Detailer to User type")
	// }

	fmt.Println("********** End of getProfileDetails Function ******************")
	//return string(userDetailer.([]byte)), nil
	return entityDetailer, nil
}

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) (string, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()

	prescriptions, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return "", err
	}

	return prescriptions, nil
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (string, error) {

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}

		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		// var record interface{}
		// err = json.Unmarshal(queryResult.Value, &record)
		// if err != nil {
		// 	return "", err
		// }
		buffer.WriteString(string(queryResult.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Println("buffer string : ", buffer.String())
	return buffer.String(), nil

}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
