package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type User struct {
	ID      string `json:"id"`
	Balance int    `json:"balance"`
}

type Asset struct {
	ID     string `json:"id"`
	Owner  string `json:"owner"`
	Price  int    `json:"price"`
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, assetID string, owner string, price int) error {

	asset := Asset{
		ID:    assetID,
		Owner: owner,
		Price: price,
	}

	assetBytes, _ := json.Marshal(asset)
	return ctx.GetStub().PutState(assetID, assetBytes)
}

func (s *SmartContract) BuyAsset(ctx contractapi.TransactionContextInterface, assetID string, buyerID string) error {

	assetBytes, _ := ctx.GetStub().GetState(assetID)
	if assetBytes == nil {
		return fmt.Errorf("asset not found")
	}

	var asset Asset
	json.Unmarshal(assetBytes, &asset)

	buyerBytes, _ := ctx.GetStub().GetState(buyerID)
	sellerBytes, _ := ctx.GetStub().GetState(asset.Owner)

	var buyer, seller User
	json.Unmarshal(buyerBytes, &buyer)
	json.Unmarshal(sellerBytes, &seller)

	if buyer.Balance < asset.Price {
		return fmt.Errorf("insufficient balance")
	}

	// transfer tokens
	buyer.Balance -= asset.Price
	seller.Balance += asset.Price

	// transfer ownership
	asset.Owner = buyerID

	// save all updates
	buyerUpdated, _ := json.Marshal(buyer)
	sellerUpdated, _ := json.Marshal(seller)
	assetUpdated, _ := json.Marshal(asset)

	ctx.GetStub().PutState(buyerID, buyerUpdated)
	ctx.GetStub().PutState(seller.ID, sellerUpdated)
	ctx.GetStub().PutState(assetID, assetUpdated)

	return nil
}

func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (string, error) {

	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return "", err
	}

	if assetBytes == nil {
		return "Asset not found", nil
	}

	return string(assetBytes), nil
}

func (s *SmartContract) MintToken(ctx contractapi.TransactionContextInterface, userID string, amount int) error {

    fmt.Println("MintToken called with:", userID, amount)

    user := User{
        ID: userID,
        Balance: amount,
    }

    userBytes, err := json.Marshal(user)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(userID, userBytes)
    if err != nil {
        return err
    }

    fmt.Println("State written for:", userID)

    return nil
}

func (s *SmartContract) TransferToken(ctx contractapi.TransactionContextInterface, from string, to string, amount int) error {

	fromBytes, _ := ctx.GetStub().GetState(from)
	toBytes, _ := ctx.GetStub().GetState(to)

	var fromUser, toUser User
	json.Unmarshal(fromBytes, &fromUser)
	json.Unmarshal(toBytes, &toUser)

	if fromUser.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	fromUser.Balance -= amount
	toUser.Balance += amount

	fromUpdated, _ := json.Marshal(fromUser)
	toUpdated, _ := json.Marshal(toUser)

	ctx.GetStub().PutState(from, fromUpdated)
	ctx.GetStub().PutState(to, toUpdated)

	return nil
}

func (s *SmartContract) ReadUser(ctx contractapi.TransactionContextInterface, userID string) (string, error) {

    userBytes, err := ctx.GetStub().GetState(userID)
    if err != nil {
        return "", err
    }

    if userBytes == nil {
        return "No data found", nil
    }

    return string(userBytes), nil
}

func main() {
	chaincode, _ := contractapi.NewChaincode(new(SmartContract))
	chaincode.Start()
}
