/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/gin-gonic/gin"
    "net/http"
    "log"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func main() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			fmt.Printf("Failed to populate wallet contents: %s\n", err)
			os.Exit(1)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		os.Exit(1)
	}

	contract := network.GetContract("fabcar")
    router := gin.Default()


	var result []byte
    router.GET("/queryallcars", func(c *gin.Context) {

	result, err := contract.EvaluateTransaction("queryAllCars")
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(result))
    c.String(http.StatusOK, string(result))
})

	router.POST("/createcar", func(c *gin.Context) {
		car := c.PostForm("car")
		make := c.PostForm("make")
		model := c.PostForm("model")
		colour := c.PostForm("colour")
		owner := c.PostForm("owner")
		result, err := contract.SubmitTransaction("createCar", car, make, model, colour, owner)
		if err != nil {
			log.Printf("Failed to submit transaction: %s\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(string(result))
		c.String(http.StatusOK, "Car created successfully")
	})

	router.GET("/querycar", func(c *gin.Context) {
		var car string
		car = c.Query("car")
		result, err = contract.EvaluateTransaction("queryCar", string(car))
		if err != nil {
			fmt.Printf("Failed to evaluate transaction: %s\n", err)
			//os.Exit(1)
		}
	   fmt.Println(string(result))
		c.String(http.StatusOK, string(result))
	})

router.POST("/changecarowner", func(c *gin.Context) {
		car := c.PostForm("car")
		newOwner := c.PostForm("newOwner")
		_, err := contract.SubmitTransaction("changeCarOwner", car, newOwner)
		if err != nil {
			log.Printf("Failed to submit transaction: %s\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "Car owner changed successfully")
	})

	result, err = contract.EvaluateTransaction("queryCar", "CAR10")
		
		if err != nil {
			fmt.Printf("Failed to evaluate transaction: %s\n", err)
			//os.Exit(1)
		}
	   
router.Run(":8000")
}

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	err = wallet.Put("appUser", identity)
	if err != nil {
		return err
	}
	return nil
}
