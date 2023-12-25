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

// const SUPER_ADMIN = "admin"

// // const VACCINE_CHAIN_ADMIN = "VACCINE_CHAIN_ADMIN"
// const MANUFACTURER = "MANUFACTURER"
// const ITEM = "ITEM"
// const idDoctypeIndex = "id~doctype"

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
	BatchCount    uint   `json:"batchCount,omitempty"`
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
	CartoonId         string `json:"cartoonId"`
	PacketId          string `json:"packetId"`
	Owner             string `json:"owner"`
	ProductId         string `json:"productId"`
	ManufacturingDate string `json:"manufacturingDate"`
	ExpiryDate        string `json:"expiryDate"`
	DocType           string `json:"docType"`
}

func (s *SmartContract) VaccineChainAdmin(ctx contractapi.TransactionContextInterface, adminInputString string) error {
	var vaccineChainAdmin VaccineChainAdmin
	err := json.Unmarshal([]byte(adminInputString), &vaccineChainAdmin)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal of input string for Admin: %v", err.Error())
	}
	fmt.Println("Input String :", vaccineChainAdmin)

	//validate super admin
	superAdminIdentity, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("superAdminIdentity :", superAdminIdentity)
	if superAdminIdentity != vaccinechainhelper.SUPER_ADMIN {
		return fmt.Errorf("permission denied: only super admin can call this function")
	}

	/*Validate Admin Id*/
	objectBytes, err := vaccinechainhelper.IsExist(ctx, vaccineChainAdmin.Id, vaccineChainAdmin.DocType)
	if err != nil {
		return err
	}

	if objectBytes != nil {
		return fmt.Errorf("Record already exist for %v with Id: %v", vaccineChainAdmin.DocType, vaccineChainAdmin.Id)
	}

	//Inserting Admin record
	vaccineChainAdminJSON, err := json.Marshal(vaccineChainAdmin)
	if err != nil {
		return fmt.Errorf("Failed to marshal of Admin record : %v", err.Error())
	}

	compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{vaccineChainAdmin.Id, vaccineChainAdmin.DocType})
	if err != nil {
		return fmt.Errorf("failed to create composite key for hospital %v and err is :%v", vaccineChainAdmin.Id, err.Error())
	}
	err = ctx.GetStub().PutState(compositeKey, vaccineChainAdminJSON)
	if err != nil {
		return fmt.Errorf("failed to insert hospital details to couchDB : %v", err.Error())
	}

	fmt.Println("********** End of Add Vaccine Chain Admin Function ******************")
	return nil
}

func (s *SmartContract) AddEntity(ctx contractapi.TransactionContextInterface, entityInputString string) error {
	var entityInput Entity
	err := json.Unmarshal([]byte(entityInputString), &entityInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal of input string for User: %v", err.Error())
	}
	fmt.Println("Input String :", entityInput)

	//fetching cerificate attributes
	attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	if err != nil {
		return err
	}
	fmt.Println("userRole :", attributes["userRole"])

	if attributes["userRole"] != vaccinechainhelper.VACCINE_CHAIN_ADMIN {
		return fmt.Errorf("Only Vaccine Chain Admin are allowed to register %v", entityInput.DocType)
	}

	/*Validate User Id*/
	objectBytes, err := vaccinechainhelper.IsExist(ctx, entityInput.Id, entityInput.DocType)
	if err != nil {
		return err
	}

	// var tempData Entity
	// err = json.Unmarshal(objectBytes.([]byte), &tempData)
	// if err != nil {
	// 	return fmt.Errorf("Failed to convert Detailer to User type")
	// }

	// userData, ok := objectBytes.(Entity)
	// if !ok {
	// 	return fmt.Errorf("Failed to convert Detailer to User type")
	// }
	// fmt.Println("tempData :", tempData)
	// fmt.Println("tempData Id:", tempData.Id)

	if objectBytes != nil {
		return fmt.Errorf("Record already exist for %v with Id: %v", entityInput.DocType, entityInput.Id)
	}

	//Inserting User record
	entityJSON, err := json.Marshal(entityInput)
	if err != nil {
		return fmt.Errorf("failed to marshal of user records : %v", err.Error())
	}

	compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{entityInput.Id, entityInput.DocType})
	if err != nil {
		return fmt.Errorf("failed to create composite key for %v and err is :%v", entityInput.Id, err.Error())
	}
	err = ctx.GetStub().PutState(compositeKey, entityJSON)
	if err != nil {
		return fmt.Errorf("failed to insert user details to couchDB : %v", err.Error())
	}
	fmt.Println("********** End of Add User Function ******************")
	return nil
}

func (s *SmartContract) AddProduct(ctx contractapi.TransactionContextInterface, productInputString string) error {
	var productInput Product
	err := json.Unmarshal([]byte(productInputString), &productInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal of input string for Product: %v", err.Error())
	}
	fmt.Println("Input String :", productInput)

	//fetching cerificate attributes
	attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	if err != nil {
		return err
	}
	fmt.Println("userRole :", attributes["userRole"])

	if attributes["userRole"] != vaccinechainhelper.MANUFACTURER {
		return fmt.Errorf("Only Manufacturer are allowed to register Product details")
	}

	//fetching manufacturing Id
	manufacturerId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("manufacturerId :", manufacturerId)
	if err != nil {
		return fmt.Errorf("Failed to get Manufacturer Id")
	}

	/*Validate User Id*/
	//productDocType := []string{productInput.DocType, manufacturerId}
	// productDocType := productInput.DocType + manufacturerId
	productId := productInput.Id + manufacturerId
	objectBytes, err := vaccinechainhelper.IsExist(ctx, productId, productInput.DocType)
	if err != nil {
		return err
	}

	if objectBytes != nil {
		return fmt.Errorf("Record already exist with Id: %v", productInput.Id)
	}

	//adding product owner
	productInput.Owner = manufacturerId

	//Inserting User record
	productJSON, err := json.Marshal(productInput)
	if err != nil {
		return fmt.Errorf("failed to marshal of user records : %v", err.Error())
	}

	compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{productId, productInput.DocType})
	if err != nil {
		return fmt.Errorf("failed to create composite key for %v and err is :%v", productInput.Id, err.Error())
	}
	err = ctx.GetStub().PutState(compositeKey, productJSON)
	if err != nil {
		return fmt.Errorf("failed to insert product details to couchDB : %v", err.Error())
	}
	fmt.Println("********** End of Add Product Function ******************")
	return nil
}

func (s *SmartContract) AddBatch(ctx contractapi.TransactionContextInterface, batchInputString string) error {
	var batchInput Batch
	err := json.Unmarshal([]byte(batchInputString), &batchInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal of input string for Batch: %v", err.Error())
	}
	fmt.Println("Input String :", batchInput)

	//fetching cerificate attributes for logged-in entity
	attributes, err := vaccinechainhelper.GetAllCertificateAttributes(ctx, []string{"userRole"})
	if err != nil {
		return err
	}
	fmt.Println("userRole :", attributes["userRole"])

	//checking the role
	if attributes["userRole"] != vaccinechainhelper.MANUFACTURER {
		return fmt.Errorf("Only Manufacturer are allowed to add batch")
	}

	//fetching manufacturing Id
	manufacturerId, err := vaccinechainhelper.GetUserIdentityName(ctx)
	fmt.Println("manufacturerId :", manufacturerId)
	if err != nil {
		return fmt.Errorf("Failed to get Manufacturer Id")
	}

	//Fetching Manufacturing details
	var manufacturerDetails Entity
	objectBytes, err := vaccinechainhelper.IsExist(ctx, manufacturerId, vaccinechainhelper.MANUFACTURER)
	if err != nil {
		return err
	}
	if objectBytes == nil {
		return fmt.Errorf("Record does not exist with Id: %v", manufacturerId)
	}
	err = json.Unmarshal(objectBytes.([]byte), &manufacturerDetails)
	fmt.Println("manufacturerDetails :", manufacturerDetails)

	//Fetching Product details
	var productDetails Product
	productId := batchInput.ProductId + manufacturerId
	objectBytes, err = vaccinechainhelper.IsExist(ctx, productId, vaccinechainhelper.ITEM)
	if err != nil {
		return err
	}
	if objectBytes == nil {
		return fmt.Errorf("Record does not exist with Id: %v", manufacturerId)
	}
	err = json.Unmarshal(objectBytes.([]byte), &productDetails)
	fmt.Println("productDetails :", productDetails)

	fmt.Println("old batch no  :", manufacturerDetails.BatchCount)
	newBatchNo := manufacturerDetails.BatchCount + 1
	fmt.Println("newBatchNo :", newBatchNo)

	//Inserting Batch Details

	batchInput.Id = "B" + strconv.Itoa(int(newBatchNo))
	fmt.Println("batchInput.Id :", batchInput.Id)

	batchJSON, err := json.Marshal(batchInput)
	if err != nil {
		return fmt.Errorf("failed to marshal of batch records : %v", err.Error())
	}

	compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{manufacturerId, batchInput.Id})
	if err != nil {
		return fmt.Errorf("failed to create composite key for %v and err is :%v", batchInput.Id, err.Error())
	}
	err = ctx.GetStub().PutState(compositeKey, batchJSON)
	if err != nil {
		return fmt.Errorf("failed to insert batch details to couchDB : %v", err.Error())
	}

	//Inserting Packet Records
	// asset := struct {
	// 	Id                string `json:"id"`
	// 	BatchId           string `json:"batchId"`
	// 	CartoonId         string `json:"cartoonId"`
	// 	PacketId          string `json:"packetId"`
	// 	Owner             string `json:"owner"`
	// 	ProductId         string `json:"productId"`
	// 	ManufacturingDate string `json:"manufacturingDate"`
	// 	ExpiryDate        string `json:"expiryDate"`
	// }{}

	var i, j int16
	var assetId, cartoonId, packetId string

	for i = 1; i <= batchInput.CartonQnty; i++ {
		for j = 1; j <= productDetails.CartonCapacity; j++ {
			cartoonId = "C" + strconv.Itoa(int(i))
			packetId = "P" + strconv.Itoa(int(j))
			assetId = batchInput.Id + "_" + cartoonId + "_" + packetId
			fmt.Println("assetId :", assetId)

			//txID := ctx.GetStub().GetTxID()
			asset := Asset{
				Id:                assetId,
				BatchId:           batchInput.Id,
				CartoonId:         cartoonId,
				PacketId:          packetId,
				Owner:             manufacturerId,
				ProductId:         productDetails.Id,
				ManufacturingDate: batchInput.ManufacturingDate,
				ExpiryDate:        batchInput.ExpiryDate,
				DocType:           vaccinechainhelper.ASSET,
			}

			assetJSON, err := json.Marshal(asset)
			if err != nil {
				return fmt.Errorf("failed to marshal of asset records : %v", err.Error())
			}

			compositeKey, err := ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{manufacturerId, assetId})
			if err != nil {
				return fmt.Errorf("failed to create composite key for %v and err is :%v", batchInput.Id, err.Error())
			}
			//compositeKey := txID
			err = ctx.GetStub().PutState(compositeKey, assetJSON)
			if err != nil {
				return fmt.Errorf("failed to insert asset details to couchDB : %v", err.Error())
			}

		}
	}

	//strconv.Itoa(int(newBatchNo))

	//Updating manufacturer Details
	manufacturerDetails.BatchCount = newBatchNo
	manufacturerJSON, err := json.Marshal(manufacturerDetails)

	if err != nil {
		return fmt.Errorf("failed to marshal of manufacturer records : %v", err.Error())
	}

	compositeKey, err = ctx.GetStub().CreateCompositeKey(vaccinechainhelper.IdDoctypeIndex, []string{manufacturerId, vaccinechainhelper.MANUFACTURER})
	if err != nil {
		return fmt.Errorf("failed to create composite key for %v and err is :%v", batchInput.Id, err.Error())
	}
	err = ctx.GetStub().PutState(compositeKey, manufacturerJSON)
	if err != nil {
		return fmt.Errorf("failed to insert batch details to couchDB : %v", err.Error())
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

// constructQueryResponseFromIterator constructs a slice of data from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (string, error) {
	//var prescriptions []*Prescription
	// var records []*interface{}
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
		var record interface{}
		err = json.Unmarshal(queryResult.Value, &record)
		if err != nil {
			return "", err
		}
		//fmt.Println("buffer string : ", string(queryResult.Value))
		buffer.WriteString(string(queryResult.Value))
		// records = append(records, &record)
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
