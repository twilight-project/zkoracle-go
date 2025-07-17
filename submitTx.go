package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	zktypes "github.com/twilight-project/nyks/x/zkos/types"
)

func sendTransactionTransferTx(accountName string, cosmos cosmosclient.Client, data *zktypes.MsgTransferTx) (cosmosclient.Response, error) {
	// Specify the broadcast mode (e.g., 'async' for asynchronous)
	// broadcastMode := "async"
	// contxt:= cosmos.Context()
	// cotxt = contxt.WithBroadcastMode(broadcastMode)
	// resp, err := cosmos.BroadcastTx(cotxt, accountName, data)
	// Specify the broadcast mode (e.g., 'async' for asynchronous)
	//txFactory := cosmos.TxFactory().WithBroadcastMode(sdktypes.BroadcastAsync)
	fmt.Println("Broadcasting Transfer Tx")
	for i := 0; i < 3; i++ {
		resp, err := cosmos.BroadcastTx(accountName, data)
		if err != nil {
			fmt.Println("Error broadcasting transaction, retrying...", err)
			time.Sleep(time.Second * 2) // wait for 2 seconds before retrying
			continue
		}
		return resp, nil
	}
	return cosmosclient.Response{}, fmt.Errorf("failed to broadcast transaction after 3 attempts : ")

}

func sendTransactionBurnMessage(accountName string, cosmos cosmosclient.Client, data *zktypes.MsgMintBurnTradingBtc) (cosmosclient.Response, error) {
	resp, err := cosmos.BroadcastTx(accountName, data)
	return resp, err
}
func getCosmosClient() cosmosclient.Client {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
	}

	homePath := filepath.Join(home, ".nyks")

	cosmosOptions := []cosmosclient.Option{
		cosmosclient.WithHome(homePath),
	}

	config := sdktypes.GetConfig()
	config.SetBech32PrefixForAccount("twilight", "twilight"+"pub")

	// create an instance of cosmosclient
	cosmos, err := cosmosclient.New(context.Background(), cosmosOptions...)
	if err != nil {
		log.Println(err)
	}

	return cosmos
}

func handleTransferTx(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Transfer Tx handler")
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var p PayloadHttpReq
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&p)
	defer req.Body.Close()

	if err != nil && err != io.EOF {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Transfer Tx: ", p.Tx)
	msg := &zktypes.MsgTransferTx{
		TxId:            p.Txid,
		TxByteCode:      p.Tx,
		ZkOracleAddress: oracleAddr,
		TxFee:           p.Fee,
	}

	cosmosClient := getCosmosClient()

	// Record the current time before executing the code
	// startTime := time.Now()
	resp, err := sendTransactionTransferTx(accountName, cosmosClient, msg)
	// // Record the current time after executing the code
	// endTime := time.Now()

	// // Calculate the elapsed time
	// elapsedTime := endTime.Sub(startTime)

	// Print the elapsed time
	// fmt.Printf("Elapsed time for Tx Broadcast: %v\n", elapsedTime)

	if err != nil {
		fmt.Println("Error in sending transfer tx :", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error" : "` + err.Error() + `"}`))
	} else {
		fmt.Println("Transfer Tx Hash: ", resp.TxHash)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"txHash" : "` + resp.TxHash + `"}`))
	}

	txCounter.Inc()
}

func handleBurnMessageTx(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var p PayloadBurnReq
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&p)
	defer req.Body.Close()

	if err != nil && err != io.EOF {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	msg := &zktypes.MsgMintBurnTradingBtc{
		MintOrBurn:      false,
		BtcValue:        p.BtcValue,
		QqAccount:       p.QqAccount,
		EncryptScalar:   p.EncryptScalar,
		TwilightAddress: p.TwilightAddress,
	}

	cosmosClient := getCosmosClient()

	resp, err := sendTransactionBurnMessage(accountName, cosmosClient, msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error" : "` + err.Error() + `"}`))
	} else {

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"txHash" : "` + resp.TxHash + `"}`))
	}

	txCounter.Inc()
}

func server() {
	fmt.Println("Server is running")
	http.HandleFunc("/transaction", handleTransferTx)
	http.HandleFunc("/burnmessage", handleBurnMessageTx)
	err := http.ListenAndServe(":7000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
