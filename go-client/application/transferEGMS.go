package main

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "log"
    "math/big"

    "golang.org/x/crypto/sha3"
    // "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/common/hexutil"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "os"
)

func main() {
    // User input for reciever address 
    var toAddressString string
    fmt.Println("Please input the address you want to transfer to:")
    fmt.Scanln(&toAddressString)

    // Connect to RPC
    client, err := ethclient.Dial("https://ethereum.rpc.evmos.dev")
    if err != nil {
        log.Fatal(err)
    }

    // Get the private key via OS Environment
    privateKey, err := crypto.HexToECDSA(os.Getenv("PRIVATE_KEY"))
    fmt.Println(privateKey)
    if err != nil {
        log.Fatal(err)
    }

    // Get the public key via private key
    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
    }

    // Get the address via public key
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        log.Fatal(err)
    }

    // Get suggested gas price
    value := big.NewInt(100) // in wei (0 eth)
    gasPrice, err := client.SuggestGasPrice(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    // EGMS address and reciever address from HEX
    toAddress := common.HexToAddress(toAddressString)
    tokenAddress := common.HexToAddress("0xe3fFAA89E058E916182e3A0e986DEE45cED77A6e")

    // Create the transaction   
    transferFnSignature := []byte("transfer(address,uint256)")
    hash := sha3.NewLegacyKeccak256()
    hash.Write(transferFnSignature)
    methodID := hash.Sum(nil)[:4]
    fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

    paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
    fmt.Println(hexutil.Encode(paddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d

    amount := new(big.Int)
    amount.SetString("1000000000000000000000", 10) // sets the value to 1000 tokens, in the token denomination

    paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
    fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

    var data []byte
    data = append(data, methodID...)
    data = append(data, paddedAddress...)
    data = append(data, paddedAmount...)

    gasLimit := uint64(240000000)

    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(gasLimit)

    // Create the transaction
    tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

    chainID, err := client.NetworkID(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
    if err != nil {
        log.Fatal(err)
    }

    err = client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("tx sent: %s", signedTx.Hash().Hex()) // tx sent: 0xa56316b637a94c4cc0331c73ef26389d6c097506d581073f927275e7a6ece0bc
}
