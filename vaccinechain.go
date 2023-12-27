/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Prasenjit43/vaccinechainhelper"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Asset and Token
type SmartContract struct {
	contractapi.Contract
}

type Entity struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	LicenseNo     string `json:"licenseNo"`
	Address       string `json:"address,omitempty"`       //updatable
	OwnerName     string `json:"ownerName,omitempty"`     //updatable
	OwnerIdentity string `json:"ownerIdentity,omitempty"` //updatable
	OwnerAddress  string `json:"ownerAddress,omitempty"`  //updatable
	ContactNo     string `json:"contactNo,omitempty"`     //updatable
	EmailId       string `json:"emailId,omitempty"`       //updatable
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
	Id                string `json:"id"`
	BatchId           string `json:"batchId"`
	CartonId          string `json:"cartonId"`
	Owner             string `json:"owner"`
	Status            string `json:"status"`
	ProductId         string `json:"productId"`
	ManufacturerId    string `json:"manufacturerId"`
	ManufacturingDate string `json:"manufacturingDate"`
	ExpiryDate        string `json:"expiryDate"`
	DocType           string `json:"docType"`
}

type Receipt struct {
	Id              string `json:"id"`
	BundleId        string `json:"bundleId"`
	DocType         string `json:"docType"`
	SupplierId      string `json:"supplierId"`
	CustomerId      string `json:"customerId"`
	ProductId       string `json:"productId"`
	TransactionDate int16  `json:"transactionDate"`
	BillAmount      int16  `json:"billAmount"`
}

type History struct {
	Id        string    `json:"id"`
	Owner     string    `json:"owner"`
	Status    string    `json:"status"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// VaccineChainAdmin function adds a new admin to the vaccine chain system.
// It takes in a JSON string containing admin details and performs several validations before inserting the admin record into the ledger.
func (s *SmartContract) VaccineChainAdmin(ctx contractapi.TransactionContextInterface, adminInputString string) error {
	// Define a struct to hold the admin details
	var vaccineChainAdmin Entity

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

	objectBytes, err := vaccinechainhelper.IsExist(ctx, vaccineChainAdmin.Id, vaccineChainAdmin.DocType)
	if err != nil {
		return err
	}
	fmt.Println("Object Bytes 	: ", objectBytes)

	// If the record already exists, return an error
	if objectBytes != nil {
		return fmt.Errorf("Record already exists for %v with Id: %v", vaccineChainAdmin.DocType, vaccineChainAdmin.Id)
	}

	err = insertData(ctx, vaccineChainAdmin, vaccineChainAdmin.Id, vaccineChainAdmin.DocType)
	if err != nil {
		return err
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

	// // Fetch certificate attributes for the logged-in user
	// attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("userRole:", attributes["userRole"])

	//validate loggedin identity
	_, role, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	// Check if the user has required permissions (Vaccine Chain Admin) to register the entity
	if role != vaccinechainhelper.VACCINE_CHAIN_ADMIN {
		return fmt.Errorf("only Vaccine Chain Admin are allowed to register %v", entityInput.DocType)
	}

	objectBytes, err := vaccinechainhelper.IsExist(ctx, entityInput.Id, entityInput.DocType)
	if err != nil {
		return err
	}
	fmt.Println("Object Bytes 	: ", objectBytes)

	// If the record already exists, return an error
	if objectBytes != nil {
		return fmt.Errorf("Record already exists for %v with Id: %v", entityInput.DocType, entityInput.Id)
	}

	err = insertData(ctx, entityInput, entityInput.Id, entityInput.DocType)
	if err != nil {
		return err
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
	// attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("userRole :", attributes["userRole"])

	//validate loggedin identity
	loggedinEntity, role, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	// Check if the user role is that of a manufacturer
	if role != vaccinechainhelper.MANUFACTURER {
		return fmt.Errorf("Only Manufacturers are allowed to register Product details")
	}

	// Fetch the manufacturing ID of the logged-in user
	// manufacturerId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	// fmt.Println("manufacturerId :", manufacturerId)
	// if err != nil {
	// 	return fmt.Errorf("Failed to get Manufacturer ID")
	// }

	// Generate a unique ID for the product using manufacturer ID and product ID
	productId := productInput.Id + loggedinEntity.Id

	// Assign the manufacturer as the owner of the product
	productInput.Owner = loggedinEntity.Id

	objectBytes, err := vaccinechainhelper.IsExist(ctx, productId, productInput.DocType)
	if err != nil {
		return err
	}
	fmt.Println("Object Bytes 	: ", objectBytes)

	// If the record already exists, return an error
	if objectBytes != nil {
		return fmt.Errorf("Record already exists for %v with Id: %v", productInput.DocType, productId)
	}

	err = insertData(ctx, productInput, productId, productInput.DocType)
	if err != nil {
		return err
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

	// // Fetch certificate attributes for the logged-in entity
	// attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("userRole :", attributes["userRole"])

	//validate loggedin identity
	manufacturerDetails, role, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	// Check the role for permission
	if role != vaccinechainhelper.MANUFACTURER {
		return fmt.Errorf("Only Manufacturer is allowed to add a batch")
	}

	// // Fetch manufacturing ID
	// manufacturerId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	// fmt.Println("manufacturerId :", manufacturerId)
	// if err != nil {
	// 	return fmt.Errorf("Failed to get Manufacturer ID")
	// }

	// Fetch Manufacturing details
	// var manufacturerDetails Entity
	// objectBytes, err := vaccinechainhelper.IsActive(ctx, manufacturerId, vaccinechainhelper.MANUFACTURER)
	// if err != nil {
	// 	return err
	// }
	// if objectBytes == nil {
	// 	return fmt.Errorf("Record does not exist with ID: %v", manufacturerId)
	// }
	// err = json.Unmarshal(objectBytes, &manufacturerDetails)
	// fmt.Println("manufacturerDetails :", manufacturerDetails)

	// Fetch Product details
	var productDetails Product
	productId := batchInput.ProductId + manufacturerDetails.Id
	objectBytes, err := vaccinechainhelper.IsActive(ctx, productId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}
	if objectBytes == nil {
		return fmt.Errorf("Record does not exist with ID: %v", productId)
	}

	err = json.Unmarshal(objectBytes, &productDetails)
	fmt.Println("productDetails :", productDetails)

	batchInput.Id = "B" + strconv.Itoa(manufacturerDetails.BatchCount)
	fmt.Println("batchInput.Id :", batchInput.Id)

	// Marshal batch details into JSON
	// batchJSON, err := json.Marshal(batchInput)
	// if err != nil {
	// 	return fmt.Errorf("Failed to marshal batch records: %v", err.Error())
	// }

	// // Create composite key and insert batch details to couchDB
	// compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdIdIndex, []string{manufacturerId, batchInput.Id})
	// if err != nil {
	// 	return fmt.Errorf("Failed to create composite key for %v: %v", batchInput.Id, err.Error())
	// }
	// err = ctx.GetStub().PutState(compositeKey, batchJSON)
	// if err != nil {
	// 	return fmt.Errorf("Failed to insert batch details to couchDB: %v", err.Error())
	// }

	err = insertData(ctx, batchInput, manufacturerDetails.Id, batchInput.Id)
	if err != nil {
		return nil
	}

	// Insert Asset Records for each item in the batch
	var i, j int16
	var assetId, cartonId, packetId string

	for i = 1; i <= batchInput.CartonQnty; i++ {
		for j = 1; j <= productDetails.CartonCapacity; j++ {
			// Generate asset IDs for each item in the batch
			cartonId = batchInput.Id + "_" + "C" + strconv.Itoa(int(i))
			packetId = "P" + strconv.Itoa(int(j))
			assetId = manufacturerDetails.Id + "_" + cartonId + "_" + packetId
			fmt.Println("assetId :", assetId)

			// Create Asset object
			asset := Asset{
				Id:                assetId,
				BatchId:           batchInput.Id,
				CartonId:          cartonId,
				Owner:             manufacturerDetails.Id,
				Status:            vaccinechainhelper.Statuses.ReadyForDistribution,
				ProductId:         productDetails.Id,
				ManufacturerId:    manufacturerDetails.Id,
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
	// manufacturerJSON, err := json.Marshal(manufacturerDetails)
	// if err != nil {
	// 	return fmt.Errorf("Failed to marshal manufacturer records: %v", err.Error())
	// }

	// // Update manufacturer details in couchDB
	// compositeKey, err = ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{manufacturerId, vaccinechainhelper.MANUFACTURER})
	// if err != nil {
	// 	return fmt.Errorf("Failed to create composite key for %v: %v", batchInput.Id, err.Error())
	// }
	// err = ctx.GetStub().PutState(compositeKey, manufacturerJSON)
	// if err != nil {
	// 	return fmt.Errorf("Failed to update batch details in couchDB: %v", err.Error())
	// }

	err = insertData(ctx, manufacturerDetails, manufacturerDetails.Id, manufacturerDetails.DocType)
	if err != nil {
		return nil
	}

	return nil
}

func (s *SmartContract) ShipToDistributer(ctx contractapi.TransactionContextInterface, distributionInputString string) error {
	distributionInput := struct {
		CustomerId          string `json:"customerId"`
		CartonId            string `json:"cartonId"`
		TransactionDate     int16  `json:"transactionDate"`
		PerUnitSellingPrice int16  `json:"perUnitSellingPrice"`
	}{}

	// event := struct {
	// 	SupplierId          string `json:"supplierId"`
	// 	CustomerId          string `json:"customerId"`
	// 	TransactionDate     int16  `json:"transactionDate"`
	// 	PerUnitSellingPrice int16  `json:"perUnitSellingPrice"`
	// 	ManufacturerId      string `json:"manufacturerId"`
	// 	ProductId           string `json:"productId"`
	// 	TotalParcelUnits    int16  `json:"TotalParcelUnits"`
	// 	TotalBill           int16  `json:"totalBill"`
	// }{}

	err := json.Unmarshal([]byte(distributionInputString), &distributionInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for distribution: %v", err.Error())
	}

	// Print input for debugging
	fmt.Println("Input String :", distributionInput)

	// Fetch entity ID for logged-in user
	// entityId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	// fmt.Println("entityId :", entityId)
	// if err != nil {
	// 	return fmt.Errorf("Failed to get Entity ID")
	// }

	//validate loggedin identity
	manufacturerDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	//Validate VendorId
	vendorBytes, err := vaccinechainhelper.IsActive(ctx, distributionInput.CustomerId, vaccinechainhelper.DISTRIBUTER)
	if err != nil {
		return err
	}
	if vendorBytes == nil {
		return fmt.Errorf("Record does not exist with ID: %v", distributionInput.CustomerId)
	}

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","cartonId":"%s"}}`, manufacturerDetails.Id, distributionInput.CartonId)
	fmt.Println("queryString : ", queryString)

	productId, manufacturerId, totalBundle, err := getQueryResultForAssetUpdateQueryString(
		ctx,
		queryString,
		distributionInput.CustomerId,
		vaccinechainhelper.Statuses.ReceivedAtDistributor)
	if err != nil {
		return err
	}

	var productDetails Product
	tempProductId := productId + manufacturerId
	productBytes, err := vaccinechainhelper.IsActive(ctx, tempProductId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}
	if productBytes == nil {
		return fmt.Errorf("Record does not exist with ID: %v", productId)
	}

	err = json.Unmarshal(productBytes, &productDetails)
	fmt.Println("productDetails :", productDetails)

	//Creating Receipt
	billAmount := distributionInput.PerUnitSellingPrice * productDetails.PacketCapacity * totalBundle
	err = createReceipt(
		ctx,
		distributionInput.CartonId,
		manufacturerDetails.Id,
		distributionInput.CustomerId,
		productId,
		distributionInput.TransactionDate,
		billAmount)
	if err != nil {
		return err
	}

	//Emit an event
	event := struct {
		SupplierId          string
		CustomerId          string
		TransactionDate     int16
		PerUnitSellingPrice int16
		ManufacturerId      string
		ProductId           string
		TotalParcelUnits    int16
		TotalBill           int16
	}{
		SupplierId:          manufacturerDetails.Id,
		CustomerId:          distributionInput.CustomerId,
		TransactionDate:     distributionInput.TransactionDate,
		PerUnitSellingPrice: distributionInput.PerUnitSellingPrice,
		ManufacturerId:      manufacturerId,
		ProductId:           productId,
		TotalParcelUnits:    totalBundle,
		TotalBill:           billAmount,
	}

	eventDataJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	eventErr := ctx.GetStub().SetEvent("Distributer Shippment Alert", eventDataJSON)
	if eventErr != nil {
		return fmt.Errorf("failed to setEvent Shippment Alert : %v", err.Error())
	}

	return nil
}

func (s *SmartContract) ShipToChemist(ctx contractapi.TransactionContextInterface, distributionInputString string) error {
	distributionInput := struct {
		CustomerId          string `json:"customerId"`
		PacketId            string `json:"packetId"`
		TransactionDate     int16  `json:"transactionDate"`
		PerUnitSellingPrice int16  `json:"perUnitSellingPrice"`
	}{}

	err := json.Unmarshal([]byte(distributionInputString), &distributionInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for distribution: %v", err.Error())
	}

	// Print input for debugging
	fmt.Println("Input String :", distributionInput)

	// Fetch entity ID for logged-in user
	// entityId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	// fmt.Println("entityId :", entityId)
	// if err != nil {
	// 	return fmt.Errorf("Failed to get Entity ID")
	// }
	distributerDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	//Validate VendorId
	vendorBytes, err := vaccinechainhelper.IsActive(ctx, distributionInput.CustomerId, vaccinechainhelper.CHEMIST)
	if err != nil {
		return err
	}
	if vendorBytes == nil {
		return fmt.Errorf("Record does not exist with ID: %v", distributionInput.CustomerId)
	}

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","id":"%s"}}`, distributerDetails.Id, distributionInput.PacketId)
	fmt.Println("queryString : ", queryString)

	productId, manufacturerId, _, err := getQueryResultForAssetUpdateQueryString(ctx,
		queryString,
		distributionInput.CustomerId,
		vaccinechainhelper.Statuses.ChemistInventoryReceived)
	if err != nil {
		return err
	}

	var productDetails Product
	tempProductId := productId + manufacturerId
	productBytes, err := vaccinechainhelper.IsActive(ctx, tempProductId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}

	err = json.Unmarshal(productBytes, &productDetails)
	fmt.Println("productDetails :", productDetails)

	//Creating Receipt
	billAmount := distributionInput.PerUnitSellingPrice * productDetails.PacketCapacity
	err = createReceipt(
		ctx,
		distributionInput.PacketId,
		distributerDetails.Id,
		distributionInput.CustomerId,
		productId,
		distributionInput.TransactionDate,
		billAmount)
	if err != nil {
		return err
	}

	//Emit an event
	event := struct {
		SupplierId          string
		CustomerId          string
		TransactionDate     int16
		PerUnitSellingPrice int16
		ManufacturerId      string
		ProductId           string
		TotalParcelUnits    int16
		TotalBill           int16
	}{
		SupplierId:          distributerDetails.Id,
		CustomerId:          distributionInput.CustomerId,
		TransactionDate:     distributionInput.TransactionDate,
		PerUnitSellingPrice: distributionInput.PerUnitSellingPrice,
		ManufacturerId:      manufacturerId,
		ProductId:           productId,
		TotalParcelUnits:    1,
		TotalBill:           billAmount,
	}

	eventDataJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	eventErr := ctx.GetStub().SetEvent("Chemist Shippment Alert", eventDataJSON)
	if eventErr != nil {
		return fmt.Errorf("failed to setEvent Shippment Alert : %v", err.Error())
	}

	return nil
}

func (s *SmartContract) ShipToCustomer(ctx contractapi.TransactionContextInterface, distributionInputString string) error {
	distributionInput := struct {
		CustomerId      string `json:"customerId"`
		PacketId        string `json:"packetId"`
		TransactionDate int16  `json:"transactionDate"`
	}{}

	err := json.Unmarshal([]byte(distributionInputString), &distributionInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for distribution: %v", err.Error())
	}

	// Print input for debugging
	fmt.Println("Input String :", distributionInput)

	// Fetch entity ID for logged-in user
	// entityId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	// fmt.Println("entityId :", entityId)
	// if err != nil {
	// 	return fmt.Errorf("Failed to get Entity ID")
	// }
	chemistDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","id":"%s"}}`, chemistDetails.Id, distributionInput.PacketId)
	fmt.Println("queryString : ", queryString)

	productId, manufacturerId, _, err := getQueryResultForAssetUpdateQueryString(ctx,
		queryString,
		distributionInput.CustomerId,
		vaccinechainhelper.Statuses.SoldToCustomer)
	if err != nil {
		return err
	}

	var productDetails Product
	tempProductId := productId + manufacturerId
	productBytes, err := vaccinechainhelper.IsActive(ctx, tempProductId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}

	err = json.Unmarshal(productBytes, &productDetails)
	fmt.Println("productDetails :", productDetails)

	//Creating Receipt
	billAmount := productDetails.Price * productDetails.PacketCapacity
	err = createReceipt(
		ctx,
		distributionInput.PacketId,
		chemistDetails.Id,
		distributionInput.CustomerId,
		productId,
		distributionInput.TransactionDate,
		billAmount)
	if err != nil {
		return err
	}

	//Emit an event
	event := struct {
		SupplierId          string
		CustomerId          string
		TransactionDate     int16
		PerUnitSellingPrice int16
		ManufacturerId      string
		ProductId           string
		TotalParcelUnits    int16
		TotalBill           int16
	}{
		SupplierId:          chemistDetails.Id,
		CustomerId:          distributionInput.CustomerId,
		TransactionDate:     distributionInput.TransactionDate,
		PerUnitSellingPrice: productDetails.Price,
		ManufacturerId:      manufacturerId,
		ProductId:           productId,
		TotalParcelUnits:    1,
		TotalBill:           billAmount,
	}

	eventDataJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	eventErr := ctx.GetStub().SetEvent("Customer Selling Alert", eventDataJSON)
	if eventErr != nil {
		return fmt.Errorf("failed to setEvent Shippment Alert : %v", err.Error())
	}

	return nil
}

func createReceipt(ctx contractapi.TransactionContextInterface, bundleId string, supplierId string, customerId string, productId string, transactionDate int16, billAmount int16) error {
	txID := ctx.GetStub().GetTxID()
	receipt := Receipt{
		Id:              txID,
		BundleId:        bundleId,
		DocType:         vaccinechainhelper.RECEIPT,
		SupplierId:      supplierId,
		CustomerId:      customerId,
		ProductId:       productId,
		TransactionDate: transactionDate,
		BillAmount:      billAmount,
	}

	// receiptJSON, err := json.Marshal(receipt)
	// if err != nil {
	// 	return fmt.Errorf("Failed to marshal receipt records: %v", err.Error())
	// }

	// err = ctx.GetStub().PutState(txID, receiptJSON)
	// if err != nil {
	// 	return fmt.Errorf("Failed to update receipt details in couchDB: %v", err.Error())
	// }
	err := insertData(ctx, receipt, txID, "")
	if err != nil {
		return nil
	}

	return nil
}

func (s *SmartContract) GetProductsByManufacturer(ctx contractapi.TransactionContextInterface) (string, error) {

	//fetching manufacturing Id
	// manufacturerId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	// fmt.Println("manufacturerId :", manufacturerId)
	// if err != nil {
	// 	return "", fmt.Errorf("Failed to get Manufacturer Id")
	// }
	manufacturerDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return "", err
	}

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","docType":"ITEM"}}`, manufacturerDetails.Id)
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
	// entityId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	// fmt.Println("entityId :", entityId)
	// if err != nil {
	// 	return "", fmt.Errorf("Failed to get entity Id")
	// }
	entityDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return "", err
	}

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","docType":"ASSET"}}`, entityDetails.Id)
	fmt.Println("queryString : ", queryString)

	currentAssetsByEntity, err := getQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return "", err
	}
	fmt.Println("currentAssetsByEntity : ", currentAssetsByEntity)

	return currentAssetsByEntity, nil
}

func (s *SmartContract) ViewProfileDetails(ctx contractapi.TransactionContextInterface) (string, error) {

	entityDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return "", err
	}
	fmt.Println("entityDetails :", entityDetails)

	entityDetailsJson, _ := json.Marshal(entityDetails)

	fmt.Println("********** End of ViewProfileDetails Function ******************")
	return string(entityDetailsJson), nil
}

func (s *SmartContract) UpdateProfile(ctx contractapi.TransactionContextInterface, updateProfileInputString string) error {
	fmt.Println("stringInput:", updateProfileInputString)
	var entityUpdateInput Entity
	err := json.Unmarshal([]byte(updateProfileInputString), &entityUpdateInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal of input string for Entity: %v", err.Error())
	}
	fmt.Println("Input String :", entityUpdateInput)

	entityDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return err
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
	// entityJSON, err := json.Marshal(entityDetails)
	// if err != nil {
	// 	return fmt.Errorf("failed to marshal of entity records : %v", err.Error())
	// }

	// compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{entityDetails.Id, entityDetails.DocType})
	// if err != nil {
	// 	return fmt.Errorf("failed to create composite key for %v and err is :%v", entityDetails.Id, err.Error())
	// }
	// err = ctx.GetStub().PutState(compositeKey, entityJSON)
	// if err != nil {
	// 	return fmt.Errorf("failed to insert user details to couchDB : %v", err.Error())
	// }

	err = insertData(ctx, entityDetails, entityDetails.Id, entityDetails.DocType)
	if err != nil {
		return nil
	}

	fmt.Println("********** End of Update Details Function ******************")

	return nil
}

func (s *SmartContract) TrackPacket(ctx contractapi.TransactionContextInterface, key string) ([]History, error) {

	// objectBytes, err := ctx.GetStub().GetState(key)
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to read data from world state %s", err.Error())
	// }
	// if objectBytes == nil {
	// 	return nil, fmt.Errorf("Packet Does not exist for ID:  %s", key)
	// }

	// var asset Asset
	// err = json.Unmarshal(objectBytes, &asset)
	// if err != nil {
	// 	return nil, err
	// }

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get history for key %s: %v", key, err.Error())
	}

	defer resultsIterator.Close()

	var histories []History
	var history History
	var checkStatus = false
	for resultsIterator.HasNext() {
		fmt.Println("Inside Result Iterator")
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("Error on fetching history : %v", err.Error())
		}
		var record Asset
		if !response.IsDelete {
			if err := json.Unmarshal(response.Value, &record); err != nil {
				return nil, fmt.Errorf("Error unmarshaling JSON: %v", err)
			}
		}

		if record.DocType != vaccinechainhelper.ASSET && !checkStatus {
			return nil, fmt.Errorf("This trackng id does not belong to asset is :%s", key)
		}
		checkStatus = true

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		history.Id = record.Id
		history.Owner = record.Owner
		history.Status = record.Status
		history.TxId = response.TxId
		history.Timestamp = timestamp
		history.IsDelete = response.IsDelete

		histories = append(histories, history)
	}
	fmt.Println("*********************")
	return histories, nil
}

func (s *SmartContract) ViewReceipt(ctx contractapi.TransactionContextInterface, receiptId string) (string, error) {
	entityDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return "", err
	}
	fmt.Println("entityDetails:", entityDetails)

	var receipt Receipt
	receiptBytes, err := ctx.GetStub().GetState(receiptId)
	if err != nil {
		return "", fmt.Errorf("failed to get receipt for Id : %s, %v", receiptId, err)
	}
	if receiptBytes == nil {
		return "", fmt.Errorf("Receipt does not exist for Id: %s", receiptId)
	}
	err = json.Unmarshal(receiptBytes, &receipt)
	if err != nil {
		return "", err
	}

	if receipt.SupplierId != entityDetails.Id && receipt.CustomerId != entityDetails.Id {
		return "", fmt.Errorf("You are not authorized to see the receipt")
	}

	return string(receiptBytes), nil
}

func (s *SmartContract) ChangeAdminStatus(ctx contractapi.TransactionContextInterface, changeStatusInputString string) error {
	changeStatusInput := struct {
		Id      string `json:"id"`
		DocType string `json:"docType"`
		Status  bool   `json:"status"`
	}{}

	err := json.Unmarshal([]byte(changeStatusInputString), &changeStatusInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for change status: %v", err.Error())
	}

	// Print input for debugging
	fmt.Println("Input String :", changeStatusInput)

	superAdminIdentity, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("superAdminIdentity :", superAdminIdentity)
	if !(superAdminIdentity == vaccinechainhelper.SUPER_ADMIN && changeStatusInput.DocType == vaccinechainhelper.VACCINE_CHAIN_ADMIN) {
		return fmt.Errorf("permission denied: only super admin can call this function")
	}

	//validating Id
	objectBytes, err := vaccinechainhelper.IsExist(ctx, changeStatusInput.Id, changeStatusInput.DocType)
	if err != nil {
		return err
	}
	if objectBytes == nil {
		return fmt.Errorf("Record for %v user does not exist", changeStatusInput.Id)
	}

	var entity Entity
	err = json.Unmarshal(objectBytes, &entity)
	fmt.Println("entity :", entity)

	status, err := vaccinechainhelper.SuspendStatus(ctx, objectBytes)
	if err != nil {
		return err
	}

	if status == changeStatusInput.Status {
		return fmt.Errorf("Status is already %v", status)
	}

	//validate loggedin identity
	// _, role, err := getProfileDetails(ctx)
	// if err != nil {
	// 	return err
	// }

	// if changeStatusInput.DocType==vaccinechainhelper.VACCINE_CHAIN_ADMIN && role != vaccinechainhelper.SUPER_ADMIN{
	// 	return fmt.Errorf("permission denied: you are not authorized to do changes")
	// }

	// if changeStatusInput.DocType==vaccinechainhelper.VACCINE_CHAIN_ADMIN

	entity.Suspended = changeStatusInput.Status
	err = insertData(ctx, entity, changeStatusInput.Id, changeStatusInput.DocType)
	if err != nil {
		return nil
	}

	return nil

}

func (s *SmartContract) ChangeEntityStatus(ctx contractapi.TransactionContextInterface, changeStatusInputString string) error {
	changeStatusInput := struct {
		Id      string `json:"id"`
		DocType string `json:"docType"`
		Status  bool   `json:"status"`
	}{}

	err := json.Unmarshal([]byte(changeStatusInputString), &changeStatusInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for change status: %v", err.Error())
	}

	// Print input for debugging
	fmt.Println("Input String :", changeStatusInput)

	//validate loggedin identity
	_, role, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	if !(role == vaccinechainhelper.VACCINE_CHAIN_ADMIN &&
		((changeStatusInput.DocType == vaccinechainhelper.MANUFACTURER) ||
			(changeStatusInput.DocType == vaccinechainhelper.DISTRIBUTER) ||
			(changeStatusInput.DocType == vaccinechainhelper.CHEMIST))) {
		return fmt.Errorf("permission denied: only admin can call this function")
	}

	//validating Id
	objectBytes, err := vaccinechainhelper.IsExist(ctx, changeStatusInput.Id, changeStatusInput.DocType)
	if err != nil {
		return err
	}
	if objectBytes == nil {
		return fmt.Errorf("Record for %v user does not exist", changeStatusInput.Id)
	}

	var entity Entity
	err = json.Unmarshal(objectBytes, &entity)
	fmt.Println("entity :", entity)

	status, err := vaccinechainhelper.SuspendStatus(ctx, objectBytes)
	if err != nil {
		return err
	}

	if status == changeStatusInput.Status {
		return fmt.Errorf("Status is already %v", status)
	}

	entity.Suspended = changeStatusInput.Status
	err = insertData(ctx, entity, changeStatusInput.Id, changeStatusInput.DocType)
	if err != nil {
		return nil
	}

	return nil

}

func getProfileDetails(ctx contractapi.TransactionContextInterface) (Entity, string, error) {

	//getting logged-in entity username
	entityIdentity, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("entityIdentity :", entityIdentity)
	if err != nil {
		return Entity{}, "", err
	}

	//getting logged-in entity attributes
	attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	if err != nil {
		return Entity{}, "", err
	}
	fmt.Println("userRole for entityIdentity :", attributes["userRole"])

	//validating User Id
	entityDetailer, err := vaccinechainhelper.IsActive(ctx, entityIdentity, attributes["userRole"])
	if err != nil {
		return Entity{}, "", err
	}
	if entityDetailer == nil {
		return Entity{}, "", fmt.Errorf("Record for %v user does not exist", entityIdentity)
	}

	var entityDetails Entity
	err = json.Unmarshal(entityDetailer, &entityDetails)
	if err != nil {
		return Entity{}, "", fmt.Errorf("Failed to convert Detailer to Entity type")
	}
	fmt.Println("entityDetails:", entityDetails)

	fmt.Println("********** End of getProfileDetails Function ******************")
	return entityDetails, attributes["userRole"], nil
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
		buffer.WriteString(string(queryResult.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Println("buffer string : ", buffer.String())
	return buffer.String(), nil

}

func getQueryResultForAssetUpdateQueryString(ctx contractapi.TransactionContextInterface, queryString string, newOwner string, newStatus string) (string, string, int16, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return "", "", 0, err
	}
	defer resultsIterator.Close()

	// Check if there are no records in the iterator
	if !resultsIterator.HasNext() {
		fmt.Println("No Records found for Transaction")
		return "", "", 0, fmt.Errorf("No Records found for Transaction")
	}

	var productId, manufacturerId string
	var totalBundle int16 = 0
	for resultsIterator.HasNext() {
		responseRange, err := resultsIterator.Next()
		if err != nil {
			return "", "", 0, err
		}

		var asset Asset
		assetBytes, err := ctx.GetStub().GetState(responseRange.Key)
		if err != nil {
			return "", "", 0, fmt.Errorf("failed to get asset %s: %v", responseRange.Key, err)
		}
		err = json.Unmarshal(assetBytes, &asset)
		if err != nil {
			return "", "", 0, err
		}

		asset.Owner = newOwner
		asset.Status = newStatus
		assetBytes, err = json.Marshal(asset)
		if err != nil {
			return "", "", 0, err
		}
		err = ctx.GetStub().PutState(responseRange.Key, assetBytes)
		if err != nil {
			return "", "", 0, fmt.Errorf("Shipment failed for asset %s: %v", asset.Id, err)
		}
		productId = asset.ProductId
		manufacturerId = asset.ManufacturerId
		totalBundle++
	}

	return productId, manufacturerId, totalBundle, nil
}

func insertData(ctx contractapi.TransactionContextInterface, entity interface{}, id string, docType string) error {
	// // Check if the admin record already exists using its ID and document type
	// objectBytes, err := vaccinechainhelper.IsActive(ctx, id, docType)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("Object Bytes 	: ", objectBytes)

	// // If the record already exists, return an error
	// if objectBytes != nil {
	// 	return fmt.Errorf("Record already exists for %v with Id: %v", docType, id)
	// }

	// Marshal the admin record into JSON format
	entityJSON, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("Failed to marshal Admin record: %v", err.Error())
	}

	// Create a composite key using the admin's ID and document type
	var compositeKey string
	if docType != "" {
		compositeKey, err = ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{id, docType})
		if err != nil {
			return fmt.Errorf("Failed to create composite key for %v: %v", id, err.Error())
		}
	} else {
		compositeKey = id
	}

	// Put the admin record into the ledger using the composite key
	err = ctx.GetStub().PutState(compositeKey, entityJSON)
	if err != nil {
		return fmt.Errorf("Failed to insert details to CouchDB: %v", err.Error())
	}

	return nil
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
