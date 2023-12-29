/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/Prasenjit43/vaccinechainhelper"
	"github.com/go-playground/validator/v10"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Asset and Token
type SmartContract struct {
	contractapi.Contract
}

type Entity struct {
	Id            string `json:"id" validate:"required"`
	Name          string `json:"name,omitempty" validate:"validName"`
	LicenseNo     string `json:"licenseNo,omitempty" validate:"required"`
	Address       string `json:"address,omitempty"`                          //updatable
	OwnerName     string `json:"ownerName,omitempty"`                        //updatable
	OwnerIdentity string `json:"ownerIdentity,omitempty"`                    //updatable
	OwnerAddress  string `json:"ownerAddress,omitempty"`                     //updatable
	ContactNo     string `json:"contactNo,omitempty" validate:"validNumber"` //updatable
	EmailId       string `json:"emailId,omitempty" validate:"email"`         //updatable
	Suspended     bool   `json:"suspended"`
	BatchCount    int    `json:"batchCount,omitempty"`
	DocType       string `json:"docType" validate:"required,oneof=VACCINE_CHAIN_ADMIN MANUFACTURER DISTRIBUTER CHEMIST"`
}

type Product struct {
	Id             string `json:"id" validate:"required"`
	Name           string `json:"name" validate:"validName"`
	Desc           string `json:"desc"`
	Type           string `json:"type"`
	Price          int16  `json:"price"`
	CartonCapacity int16  `json:"cartonCapacity"`
	PacketCapacity int16  `json:"packetCapacity"`
	DocType        string `json:"docType" validate:"required,eq=ITEM"`
	Suspended      bool   `json:"suspended"`
	Owner          string `json:"owner"`
}

type Batch struct {
	Id                string `json:"id,omitempty"`
	Owner             string `json:"owner"`
	ProductId         string `json:"productId"`
	ManufacturingDate int64  `json:"manufacturingDate" validate:"required"`
	ExpiryDate        int64  `json:"expiryDate" validate:"required,expiryGreaterThanManufacturing"`
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
	ManufacturingDate int64  `json:"manufacturingDate"`
	ExpiryDate        int64  `json:"expiryDate"`
	DocType           string `json:"docType"`
}

type Receipt struct {
	Id              string `json:"id"`
	BundleId        string `json:"bundleId"`
	DocType         string `json:"docType"`
	SupplierId      string `json:"supplierId"`
	CustomerId      string `json:"customerId"`
	ProductId       string `json:"productId"`
	TransactionDate int64  `json:"transactionDate"`
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

/*
VaccineChainAdmin function adds a new admin to the vaccine chain system.
It takes in a JSON string containing admin details and performs several validations
before inserting the admin record into the ledger.

@param ctx: TransactionContextInterface for the smart contract
@param adminInputString: JSON string with Admin details

@returns error: Returns an error if any validation fails or if the process encounters an issue while
interacting with the ledger.
*/
func (s *SmartContract) VaccineChainAdmin(ctx contractapi.TransactionContextInterface, adminInputString string) error {
	var vaccineChainAdmin Entity

	/* Unmarshals the input JSON string into the VaccineChainAdmin struct */
	err := json.Unmarshal([]byte(adminInputString), &vaccineChainAdmin)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal the input string for Admin: %v", err.Error())
	}
	fmt.Println("Input String: ", vaccineChainAdmin)

	/* Validates input parameters */
	err = validateInputParams(vaccineChainAdmin)
	if err != nil {
		return err
	}

	/* Validates the identity of the caller as the super admin */
	superAdminIdentity, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("Super Admin Identity: ", superAdminIdentity)
	if superAdminIdentity != vaccinechainhelper.SUPER_ADMIN {
		return fmt.Errorf("Permission denied: only the super admin can call this function")
	}

	/* Checks if admin is already present or not */
	objectBytes, err := vaccinechainhelper.IsExist(ctx, vaccineChainAdmin.Id, vaccineChainAdmin.DocType)
	if err != nil {
		return err
	}
	fmt.Println("Object Bytes: ", objectBytes)
	if objectBytes != nil {
		return fmt.Errorf("Record already exists for %v with ID: %v", vaccineChainAdmin.DocType, vaccineChainAdmin.Id)
	}

	/* Inserts Admin Details into the ledger */
	err = insertData(ctx, vaccineChainAdmin, vaccineChainAdmin.Id, vaccineChainAdmin.DocType)
	if err != nil {
		return err
	}

	fmt.Println("********** End of Add Vaccine Chain Admin Function ******************")
	return nil
}

/*
AddEntity function adds a new entity (Manufacturer, Distributor, Chemist) to the vaccine chain system.
It takes in a JSON string containing entity details and performs several validations
before inserting the entity record into the ledger.

@param ctx: TransactionContextInterface for the smart contract
@param entityInputString: JSON string with Entity details

@returns error: Returns an error if any validation fails or if the process encounters an issue while
interacting with the ledger.
*/
func (s *SmartContract) AddEntity(ctx contractapi.TransactionContextInterface, entityInputString string) error {
	/* Unmarshals the input JSON string into the Entity struct */
	var entityInput Entity
	err := json.Unmarshal([]byte(entityInputString), &entityInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal input string for Entity: %v", err.Error())
	}
	fmt.Println("Input String:", entityInput)

	/* Validates input parameters */
	err = validateInputParams(entityInput)
	if err != nil {
		return err
	}

	/* Validates logged-in entity whether it is active or not */
	_, role, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	/* Check if the user has required permissions (Vaccine Chain Admin) to register the entity */
	if role != vaccinechainhelper.VACCINE_CHAIN_ADMIN {
		return fmt.Errorf("only Vaccine Chain Admin are allowed to register %v", entityInput.DocType)
	}

	/* Checks if the entity is already present or not */
	objectBytes, err := vaccinechainhelper.IsExist(ctx, entityInput.Id, entityInput.DocType)
	if err != nil {
		return err
	}
	fmt.Println("Object Bytes: ", objectBytes)
	if objectBytes != nil {
		return fmt.Errorf("Record already exists for %v with Id: %v", entityInput.DocType, entityInput.Id)
	}

	/* Inserts Entity Details into the ledger */
	err = insertData(ctx, entityInput, entityInput.Id, entityInput.DocType)
	if err != nil {
		return err
	}

	fmt.Println("********** End of Add Entity Function ******************")
	return nil
}

/*
The AddProduct function adds a new Product to the ledger. It is exclusively called by the Manufacturer.
This function receives a JSON string containing Product details, conducts multiple validations, and
then inserts the product record into the ledger.

@param ctx: TransactionContextInterface for the smart contract
@param productInputString: JSON string with Product details

@returns error: Returns an error if any validation fails or if the process encounters an issue while
interacting with the ledger.
*/
func (s *SmartContract) AddProduct(ctx contractapi.TransactionContextInterface, productInputString string) error {
	/* Unmarshals the input JSON string into the Product struct */
	var productInput Product
	err := json.Unmarshal([]byte(productInputString), &productInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal the input string for Product: %v", err.Error())
	}
	fmt.Println("Input String: ", productInput)

	/* Validates input parameters */
	err = validateInputParams(productInput)
	if err != nil {
		return err
	}

	/* Validates the logged-in entity to ensure it is active */
	loggedinEntity, role, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	/* Checks if the user role is that of a manufacturer */
	if role != vaccinechainhelper.MANUFACTURER {
		return fmt.Errorf("Only Manufacturers are allowed to register Product details")
	}

	/* Updates product attributes based on business logic */
	productId := productInput.Id + loggedinEntity.Id
	productInput.Owner = loggedinEntity.Id

	/* Checks if the product, created by the manufacturer, already exists */
	objectBytes, err := vaccinechainhelper.IsExist(ctx, productId, productInput.DocType)
	if err != nil {
		return err
	}
	fmt.Println("Object Bytes: ", objectBytes)
	if objectBytes != nil {
		return fmt.Errorf("Record already exists for %v with Id: %v", productInput.DocType, productId)
	}

	/* Inserts Product Details into the ledger */
	err = insertData(ctx, productInput, productId, productInput.DocType)
	if err != nil {
		return err
	}

	fmt.Println("********** End of Add Product Function ******************")
	return nil
}

/*
AddBatch function incorporates a new Product Batch into the ledger, exclusively called by the Manufacturer.
This function takes a JSON string containing Product Batch details, performs various validations,
and subsequently inserts the batch record into the ledger. Additionally, it generates assets associated with the batch.

@param ctx: TransactionContextInterface for the smart contract
@param batchInputString: JSON string with Batch details

@returns error: Returns an error if any validation fails or if there's an issue interacting with the ledger.
*/
func (s *SmartContract) AddBatch(ctx contractapi.TransactionContextInterface, batchInputString string) error {
	/* Unmarshals the input JSON string into the Batch struct */
	var batchInput Batch
	err := json.Unmarshal([]byte(batchInputString), &batchInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for Batch: %v", err.Error())
	}
	fmt.Println("Input String:", batchInput)

	/* Validates input parameters */
	err = validateInputParams(batchInput)
	if err != nil {
		return err
	}

	/* Validates the logged-in entity to ensure it is active */
	manufacturerDetails, role, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	/* Checks if the user role is that of a manufacturer */
	if role != vaccinechainhelper.MANUFACTURER {
		return fmt.Errorf("Only the Manufacturer is allowed to add a batch")
	}

	/* Checks if the product, created by the manufacturer, exists */
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
	fmt.Println("Product Details:", productDetails)

	/* Generating a unique Batch No */
	batchInput.Id = "B" + strconv.Itoa(manufacturerDetails.BatchCount)
	fmt.Println("Batch ID:", batchInput.Id)

	/* Inserts Batch Details into the ledger */
	err = insertData(ctx, batchInput, manufacturerDetails.Id, batchInput.Id)
	if err != nil {
		return nil
	}

	/* Insert Asset Records for each item in the batch */
	var i, j int16
	var assetId, cartonId, packetId string

	for i = 1; i <= batchInput.CartonQnty; i++ {
		for j = 1; j <= productDetails.CartonCapacity; j++ {
			/* Generate asset IDs for each item in the batch */
			cartonId = batchInput.Id + "_" + "C" + strconv.Itoa(int(i))
			packetId = "P" + strconv.Itoa(int(j))
			assetId = manufacturerDetails.Id + "_" + cartonId + "_" + packetId
			fmt.Println("Asset ID:", assetId)

			/* Create Asset object */
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

			/* Inserts Asset Details into the ledger */
			assetJSON, err := json.Marshal(asset)
			if err != nil {
				return fmt.Errorf("Failed to marshal asset records: %v", err.Error())
			}

			err = ctx.GetStub().PutState(assetId, assetJSON)
			if err != nil {
				return fmt.Errorf("Failed to insert asset details to CouchDB: %v", err.Error())
			}
		}
	}

	/* Update batch count attribute for the manufacturer */
	newBatchNo := manufacturerDetails.BatchCount + 1
	manufacturerDetails.BatchCount = newBatchNo

	/* Updating manufacturer Details into the ledger */
	err = insertData(ctx, manufacturerDetails, manufacturerDetails.Id, manufacturerDetails.DocType)
	if err != nil {
		return nil
	}

	return nil
}

/*
ShipToDistributor function is exclusively called by the Manufacturer.
This function takes a JSON string containing Shipment details and performs the following operations:

1. Transfers the Assets corresponding to BatchId that belong to the manufacturer, to the Distributor.
2. Creates a receipt for the shipment details.
3. Emits an event for the transaction.

@param ctx: TransactionContextInterface for the smart contract.
@param distributionInputString: JSON string with Distributor Shipment details.

@returns error: Returns an error if any validation fails or if there's an issue interacting with the ledger.
*/
func (s *SmartContract) ShipToDistributor(ctx contractapi.TransactionContextInterface, distributionInputString string) error {
	distributionInput := struct {
		CustomerId          string `json:"customerId"`
		CartonId            string `json:"cartonId"`
		TransactionDate     int64  `json:"transactionDate"`
		PerUnitSellingPrice int16  `json:"perUnitSellingPrice"`
	}{}

	/* Unmarshals the input JSON string into the unnamed struct */
	err := json.Unmarshal([]byte(distributionInputString), &distributionInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for distribution: %v", err.Error())
	}
	fmt.Println("Input String:", distributionInput)

	/* Validates the logged-in entity to ensure it is active */
	manufacturerDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	/* Checks if the vendor exists */
	vendorBytes, err := vaccinechainhelper.IsActive(ctx, distributionInput.CustomerId, vaccinechainhelper.DISTRIBUTER)
	if err != nil {
		return err
	}
	if vendorBytes == nil {
		return fmt.Errorf("Record does not exist with ID: %v", distributionInput.CustomerId)
	}

	/* Updating Owner from Manufacturer to distributor for all assets corresponding to batchid */
	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","cartonId":"%s"}}`, manufacturerDetails.Id, distributionInput.CartonId)
	fmt.Println("queryString:", queryString)

	productId, manufacturerId, totalBundle, err := getQueryResultForAssetUpdateQueryString(
		ctx,
		queryString,
		distributionInput.CustomerId,
		vaccinechainhelper.Statuses.ReceivedAtDistributor)
	if err != nil {
		return err
	}

	fmt.Println("productId:", productId)
	fmt.Println("manufacturerId:", manufacturerId)
	fmt.Println("totalBundle:", totalBundle)

	/* Checks if the product, created by the manufacturer, exists */
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
	fmt.Println("productDetails:", productDetails)

	/* Creating Receipt for the shipment transaction */
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

	/* Emitting an event for the shipment transaction */
	event := struct {
		SupplierId          string
		CustomerId          string
		TransactionDate     int64
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

	eventErr := ctx.GetStub().SetEvent("Distributor Shipment Alert", eventDataJSON)
	if eventErr != nil {
		return fmt.Errorf("failed to setEvent Shipment Alert: %v", err.Error())
	}

	return nil
}

/*
The ShipToChemist function is specifically invoked by the Distributor.
It processes a JSON string holding Shipment details and executes the following actions:

1. Transfers Assets from the Distributor to the Chemist.
2. Generates a receipt for the shipment specifics.
3. Emits an event to mark the transaction.

@param ctx: TransactionContextInterface for the smart contract.
@param distributionInputString: JSON string containing Chemist Shipment details.

@returns error: Returns an error if any validation fails or if there's an issue interacting with the ledger.
*/
func (s *SmartContract) ShipToChemist(ctx contractapi.TransactionContextInterface, distributionInputString string) error {
	distributionInput := struct {
		CustomerId          string `json:"customerId"`
		PacketId            string `json:"packetId"`
		TransactionDate     int64  `json:"transactionDate"`
		PerUnitSellingPrice int16  `json:"perUnitSellingPrice"`
	}{}

	/* Unmarshals the input JSON string into the unnamed struct */
	err := json.Unmarshal([]byte(distributionInputString), &distributionInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for distribution: %v", err.Error())
	}
	fmt.Println("Input String :", distributionInput)

	/* Validates the logged-in entity to ensure it is active */
	distributerDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	/* Checks if the vendor exists */
	vendorBytes, err := vaccinechainhelper.IsActive(ctx, distributionInput.CustomerId, vaccinechainhelper.CHEMIST)
	if err != nil {
		return err
	}
	if vendorBytes == nil {
		return fmt.Errorf("Record does not exist with ID: %v", distributionInput.CustomerId)
	}

	/* Updates Owner from Distributor to Chemist for an asset */
	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","id":"%s"}}`, distributerDetails.Id, distributionInput.PacketId)
	fmt.Println("queryString : ", queryString)

	productId, manufacturerId, _, err := getQueryResultForAssetUpdateQueryString(ctx,
		queryString,
		distributionInput.CustomerId,
		vaccinechainhelper.Statuses.ChemistInventoryReceived)
	if err != nil {
		return err
	}

	/* Checks if the product, created by the manufacturer, exists */
	var productDetails Product
	tempProductId := productId + manufacturerId
	productBytes, err := vaccinechainhelper.IsActive(ctx, tempProductId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}

	err = json.Unmarshal(productBytes, &productDetails)
	fmt.Println("productDetails :", productDetails)

	/* Creates a Receipt for the shipment transaction */
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

	/* Emits an event for the shipment transaction */
	event := struct {
		SupplierId          string
		CustomerId          string
		TransactionDate     int64
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

	eventErr := ctx.GetStub().SetEvent("Chemist Shipment Alert", eventDataJSON)
	if eventErr != nil {
		return fmt.Errorf("failed to setEvent Shipment Alert : %v", err.Error())
	}

	return nil
}

/*
The ShipToCustomer function is specifically invoked by the Chemist.
It processes a JSON string holding Shipment details and executes the following actions:

1. Transfers Assets from the Chemist to the Customer.
2. Generates a receipt for the selling specifics.
3. Emits an event to mark the transaction.

@param ctx: TransactionContextInterface for the smart contract.
@param distributionInputString: JSON string containing Customer Selling details.

@returns error: Returns an error if any validation fails or if there's an issue interacting with the ledger.
*/
func (s *SmartContract) ShipToCustomer(ctx contractapi.TransactionContextInterface, distributionInputString string) error {
	distributionInput := struct {
		CustomerId      string `json:"customerId"`
		PacketId        string `json:"packetId"`
		TransactionDate int64  `json:"transactionDate"`
	}{}

	/* Unmarshals the input JSON string into the unnamed struct */
	err := json.Unmarshal([]byte(distributionInputString), &distributionInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal the input string for distribution: %v", err.Error())
	}
	fmt.Println("Input String:", distributionInput)

	/* Validates the logged-in entity to ensure it is active */
	chemistDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	/* Updates Owner from Chemist to customer for an asset */
	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","id":"%s"}}`, chemistDetails.Id, distributionInput.PacketId)
	fmt.Println("queryString:", queryString)

	productId, manufacturerId, _, err := getQueryResultForAssetUpdateQueryString(ctx,
		queryString,
		distributionInput.CustomerId,
		vaccinechainhelper.Statuses.SoldToCustomer)
	if err != nil {
		return err
	}

	/* Checks if the product, created by the manufacturer, exists */
	var productDetails Product
	tempProductId := productId + manufacturerId
	productBytes, err := vaccinechainhelper.IsActive(ctx, tempProductId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}

	err = json.Unmarshal(productBytes, &productDetails)
	fmt.Println("productDetails:", productDetails)

	/* Creates a Receipt for the sell transaction */
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

	/* Emits an event for the sell transaction */
	event := struct {
		SupplierId          string
		CustomerId          string
		TransactionDate     int64
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
		return fmt.Errorf("failed to setEvent Shipment Alert: %v", err.Error())
	}

	return nil
}

/*
createReceipt function creates a receipt for a transaction.
@returns error: Returns an error if any validation fails or if there's an issue interacting with the ledger.
*/
func createReceipt(ctx contractapi.TransactionContextInterface, bundleId string, supplierId string, customerId string, productId string, transactionDate int64, billAmount int16) error {
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

	/* Inserts receipt details into the ledger */
	err := insertData(ctx, receipt, txID, "")
	if err != nil {
		return nil
	}
	return nil
}

/*
GetProductsByManufacturer retrieves all products listed by the manufacturer and is exclusively called by the manufacturer.

@param ctx: TransactionContextInterface for the smart contract

@returns error: Returns an error if any validation fails or if there's an issue while interacting with the ledger.
@returns string: List of products created by the manufacturer
*/
func (s *SmartContract) GetProductsByManufacturer(ctx contractapi.TransactionContextInterface) (string, error) {

	/* Validates the logged-in entity to ensure it is active */
	manufacturerDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return "", err
	}

	/* Retrieves the list of products created by the manufacturer */
	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","docType":"ITEM"}}`, manufacturerDetails.Id)
	fmt.Println("queryString: ", queryString)

	productsByManufacturer, err := getQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return "", err
	}
	fmt.Println("productsByManufacturer: ", productsByManufacturer)

	return productsByManufacturer, nil
}

/*
GetAssetByEntity retrieves all assets held by the entity.

@param ctx: TransactionContextInterface for the smart contract

@returns error: Returns an error if any validation fails or if there is an issue while interacting with the ledger.
@returns string: List of assets held by the entity
*/
func (s *SmartContract) GetAssetByEntity(ctx contractapi.TransactionContextInterface) (string, error) {

	/* Validates the logged-in entity to ensure it is active */
	entityDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return "", err
	}

	/* Retrieves the list of assets held by the entity */
	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","docType":"ASSET"}}`, entityDetails.Id)
	fmt.Println("queryString: ", queryString)

	currentAssetsByEntity, err := getQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return "", err
	}
	fmt.Println("currentAssetsByEntity: ", currentAssetsByEntity)

	return currentAssetsByEntity, nil
}

/*
ViewProfileDetails retrieves the profile details of an entity.

@param ctx: TransactionContextInterface for the smart contract

@returns error: Returns an error if any validation fails or if there is an issue while interacting with the ledger.
@returns string: Returns the JSON-encoded entity profile details
*/
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

/*
TrackPacket retrieves the entire history of an asset from Manufacturer to Customer.

@param ctx: TransactionContextInterface for the smart contract
@param key: Key for the asset

@returns []History : Returns an array of JSON objects representing the asset's history
@returns error: Returns an error if any validation fails or if there is an issue while interacting with the ledger.
*/
func (s *SmartContract) TrackPacket(ctx contractapi.TransactionContextInterface, key string) ([]History, error) {

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
			return nil, fmt.Errorf("error fetching history: %v", err.Error())
		}
		var record Asset
		if !response.IsDelete {
			if err := json.Unmarshal(response.Value, &record); err != nil {
				return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
			}
		}

		if record.DocType != vaccinechainhelper.ASSET && !checkStatus {
			return nil, fmt.Errorf("this tracking ID does not belong to asset: %s", key)
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

/*
ViewReceipt retrieves the entire details of a specific receipt transaction.
This function can only be called by the Supplier and Vendor involved in this transaction.

@param ctx: TransactionContextInterface for the smart contract
@param receiptId: ReceiptID for the transaction

@returns string: Returns the complete receipt details for the transaction
@returns error: Returns an error if any validation fails or if there's an issue while interacting with the ledger.
*/

func (s *SmartContract) ViewReceipt(ctx contractapi.TransactionContextInterface, receiptId string) (string, error) {

	/* Validates the logged-in entity to ensure it is active */
	entityDetails, _, err := getProfileDetails(ctx)
	if err != nil {
		return "", err
	}
	fmt.Println("entityDetails:", entityDetails)

	/* Retrieving receipt details using the receipt ID */
	var receipt Receipt
	receiptBytes, err := ctx.GetStub().GetState(receiptId)
	if err != nil {
		return "", fmt.Errorf("failed to get receipt for ID: %s, %v", receiptId, err)
	}
	if receiptBytes == nil {
		return "", fmt.Errorf("Receipt does not exist for ID: %s", receiptId)
	}
	err = json.Unmarshal(receiptBytes, &receipt)
	if err != nil {
		return "", err
	}

	/* Verifying if the logged-in entity is authorized to view the receipt */
	if receipt.SupplierId != entityDetails.Id && receipt.CustomerId != entityDetails.Id {
		return "", fmt.Errorf("You are not authorized to view the receipt")
	}

	return string(receiptBytes), nil
}

/*
ChangeAdminStatus modifies the status of VaccineChainAdmin (Suspend/Unsuspend).
This function requires Super Admin privileges for execution.

@param ctx: TransactionContextInterface for the smart contract
@param changeStatusInputString: JSON string containing Admin details

@returns error: Returns an error if the transaction encounters issues
*/
func (s *SmartContract) ChangeAdminStatus(ctx contractapi.TransactionContextInterface, changeStatusInputString string) error {
	changeStatusInput := struct {
		Id      string `json:"id"`
		DocType string `json:"docType" validate:"required,eq=VACCINE_CHAIN_ADMIN"`
		Status  bool   `json:"status"`
	}{}

	/* Unmarshals the input JSON string into the unnamed struct */
	err := json.Unmarshal([]byte(changeStatusInputString), &changeStatusInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal the input string for status change: %v", err.Error())
	}
	fmt.Println("Input String:", changeStatusInput)

	/* Validates input parameters */
	err = validateInputParams(changeStatusInput)
	if err != nil {
		return err
	}

	/* Verifies if the logged-in entity is authorized to change the status */
	superAdminIdentity, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("superAdminIdentity:", superAdminIdentity)
	if !(superAdminIdentity == vaccinechainhelper.SUPER_ADMIN && changeStatusInput.DocType == vaccinechainhelper.VACCINE_CHAIN_ADMIN) {
		return fmt.Errorf("Permission denied: Only the super admin can call this function")
	}

	/* Verifies if the ID for which the status needs to be updated exists or not */
	objectBytes, err := vaccinechainhelper.IsExist(ctx, changeStatusInput.Id, changeStatusInput.DocType)
	if err != nil {
		return err
	}
	if objectBytes == nil {
		return fmt.Errorf("Record for user %v does not exist", changeStatusInput.Id)
	}

	var entity Entity
	err = json.Unmarshal(objectBytes, &entity)
	fmt.Println("entity:", entity)

	/* Verifies the status of the ID for which the status needs to be updated */
	status, err := vaccinechainhelper.SuspendStatus(ctx, objectBytes)
	if err != nil {
		return err
	}

	if status == changeStatusInput.Status {
		return fmt.Errorf("Status is already %v", status)
	}

	/* Updates the new status to the ledger */
	entity.Suspended = changeStatusInput.Status
	err = insertData(ctx, entity, changeStatusInput.Id, changeStatusInput.DocType)
	if err != nil {
		return nil
	}

	return nil
}

/*
ChangeEntityStatus modifies the status of Entity (Suspend/Unsuspend).
This function requires Admin privileges for execution.

@param ctx: TransactionContextInterface for the smart contract
@param changeStatusInputString: JSON string containing Admin details

@returns error: Returns an error if the transaction encounters issues
*/
func (s *SmartContract) ChangeEntityStatus(ctx contractapi.TransactionContextInterface, changeStatusInputString string) error {
	changeStatusInput := struct {
		Id      string `json:"id"`
		DocType string `json:"docType" validate:"required,oneof=MANUFACTURER DISTRIBUTER CHEMIST"`
		Status  bool   `json:"status"`
	}{}

	/* Unmarshals the input JSON string into the unnamed struct */
	err := json.Unmarshal([]byte(changeStatusInputString), &changeStatusInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for change status: %v", err.Error())
	}
	fmt.Println("Input String :", changeStatusInput)

	/* Validates input parameters */
	err = validateInputParams(changeStatusInput)
	if err != nil {
		return err
	}

	/* Validates logged-in entity whether it is active or not */
	_, role, err := getProfileDetails(ctx)
	if err != nil {
		return err
	}

	/* Verifies if the logged-in entity is authorized to change the status */
	if !(role == vaccinechainhelper.VACCINE_CHAIN_ADMIN &&
		((changeStatusInput.DocType == vaccinechainhelper.MANUFACTURER) ||
			(changeStatusInput.DocType == vaccinechainhelper.DISTRIBUTER) ||
			(changeStatusInput.DocType == vaccinechainhelper.CHEMIST))) {
		return fmt.Errorf("permission denied: only admin can call this function")
	}

	/* Verifies if the ID for which the status needs to be updated exists or not */
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

	/* Verifies the status of the ID for which the status needs to be updated */
	status, err := vaccinechainhelper.SuspendStatus(ctx, objectBytes)
	if err != nil {
		return err
	}

	if status == changeStatusInput.Status {
		return fmt.Errorf("Status is already %v", status)
	}

	/* Updates the new status to the ledger */
	entity.Suspended = changeStatusInput.Status
	err = insertData(ctx, entity, changeStatusInput.Id, changeStatusInput.DocType)
	if err != nil {
		return nil
	}

	return nil

}

/*
ChangeStatus modifies the status of Entity (Suspend/Unsuspend).
This function requires Admin privileges for execution.

@param ctx: TransactionContextInterface for the smart contract
@param changeStatusInputString: JSON string containing Entity details

@returns error: Returns an error if the transaction encounters issues
*/
func (s *SmartContract) ChangeStatus(ctx contractapi.TransactionContextInterface, changeStatusInputString string) error {
	changeStatusInput := struct {
		Id      string `json:"id"`
		DocType string `json:"docType" validate:"required,oneof=VACCINE_CHAIN_ADMIN MANUFACTURER DISTRIBUTER CHEMIST"`
		Status  bool   `json:"status"`
	}{}

	/* Unmarshals the input JSON string into the unnamed struct */
	err := json.Unmarshal([]byte(changeStatusInputString), &changeStatusInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal input string for change status: %v", err.Error())
	}
	fmt.Println("Input String :", changeStatusInput)

	/* Validates input parameters */
	err = validateInputParams(changeStatusInput)
	if err != nil {
		return err
	}

	if changeStatusInput.DocType == vaccinechainhelper.VACCINE_CHAIN_ADMIN {
		/* Verifies if the logged-in entity is authorized to change the status */
		superAdminIdentity, err := vaccinechainhelper.GetUserIdentityName(ctx)
		if err != nil {
			return err
		}
		fmt.Println("superAdminIdentity:", superAdminIdentity)
		if !(superAdminIdentity == vaccinechainhelper.SUPER_ADMIN && changeStatusInput.DocType == vaccinechainhelper.VACCINE_CHAIN_ADMIN) {
			return fmt.Errorf("Permission denied: Only the super admin can call this function")
		}
	} else {
		/* Validates logged-in entity whether it is active or not */
		_, role, err := getProfileDetails(ctx)
		if err != nil {
			return err
		}

		/* Verifies if the logged-in entity is authorized to change the status */
		if !(role == vaccinechainhelper.VACCINE_CHAIN_ADMIN &&
			((changeStatusInput.DocType == vaccinechainhelper.MANUFACTURER) ||
				(changeStatusInput.DocType == vaccinechainhelper.DISTRIBUTER) ||
				(changeStatusInput.DocType == vaccinechainhelper.CHEMIST))) {
			return fmt.Errorf("permission denied: only admin can call this function")
		}
	}

	/* Verifies if the ID for which the status needs to be updated exists or not */
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

	/* Verifies the status of the ID for which the status needs to be updated */
	status, err := vaccinechainhelper.SuspendStatus(ctx, objectBytes)
	if err != nil {
		return err
	}

	if status == changeStatusInput.Status {
		return fmt.Errorf("Status is already %v", status)
	}

	/* Updates the new status to the ledger */
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

func validateInputParams(object interface{}) error {
	validate := validator.New()
	validate.RegisterValidation("validName", validateName)
	validate.RegisterValidation("validNumber", validateNumber)
	validate.RegisterValidation("expiryGreaterThanManufacturing", expiryGreaterThanManufacturing)
	err := validate.Struct(object)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		var errMsg string
		for _, e := range validationErrors {
			fmt.Printf("Validation Error: Field %s failed validation$$$$", e.Field())
			errMsg += fmt.Sprintf("Validation Error: Field %s failed validation$$$$", e.Field())
		}
		fmt.Println("errMsg :", errMsg)
		return fmt.Errorf("validation error(s): %s", errMsg)
	}
	return nil
}

func validateName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	// Define a regular expression to allow only characters (no special characters or numbers)
	regex := regexp.MustCompile("^[A-Za-z ]+$")

	return regex.MatchString(name)
}

func validateNumber(fl validator.FieldLevel) bool {
	number := fl.Field().String()
	// Define a regular expression to allow only numbers
	regex := regexp.MustCompile("^[0-9]+$")

	return regex.MatchString(number)
}

func expiryGreaterThanManufacturing(fl validator.FieldLevel) bool {
	expiry := fl.Field().Int()
	manufacturing := fl.Parent().FieldByName("ManufacturingDate").Int()
	fmt.Println("expiry :", expiry)
	fmt.Println("manufacturing :", manufacturing)

	return expiry > manufacturing
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
