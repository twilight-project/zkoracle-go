// submitTx.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	cosmosaccount "github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	zktypes "github.com/twilight-project/nyks/x/zkos/types"
)

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func getCosmosClient() cosmosclient.Client {
	homePath := filepath.Join("/home/ubuntu/.nyks")

	cfg := sdktypes.GetConfig()
	cfg.SetBech32PrefixForAccount("twilight", "twilight"+"pub")

	cosmos, err := cosmosclient.New(
		context.Background(),
		cosmosclient.WithHome(homePath),
	)
	if err != nil {
		log.Fatalf("failed to create cosmos client: %v", err)
	}
	return cosmos
}

func broadcastWithRetry(
	ctx context.Context,
	client cosmosclient.Client,
	acc cosmosaccount.Account,
	msg sdktypes.Msg,
) (cosmosclient.Response, error) {
	fmt.Println("Broadcasting Transfer Tx")
	for i := 0; i < 3; i++ {
		resp, err := client.BroadcastTx(ctx, acc, msg)
		if err == nil {
			return resp, nil
		}
		fmt.Println("broadcast failed, retrying…", err)
		time.Sleep(2 * time.Second)
	}
	return cosmosclient.Response{}, fmt.Errorf("failed to broadcast tx after 3 attempts")
}

// -----------------------------------------------------------------------------
// Transfer‑Tx
// -----------------------------------------------------------------------------

func handleTransferTx(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Transfer Tx handler")
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var p PayloadHttpReq
	if err := json.NewDecoder(req.Body).Decode(&p); err != nil && err != io.EOF {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	_ = req.Body.Close()
	fmt.Println("Transfer Tx: ", p.Tx)
	msg := &zktypes.MsgTransferTx{
		TxId:            p.Txid,
		TxByteCode:      p.Tx,
		ZkOracleAddress: oracleAddr,
		TxFee:           p.Fee,
	}

	cosmos := getCosmosClient()
	acc, err := cosmos.Account(accountName)
	if err != nil {
		http.Error(w, "account not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := broadcastWithRetry(req.Context(), cosmos, acc, msg)
	if err != nil {
		fmt.Println("Error in sending transfer tx :", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error" : "` + err.Error() + `"}`))
	} else {
		fmt.Println("Transfer Tx Hash: ", resp.TxHash)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"txHash" : "` + resp.TxHash + `"}`))
		txCounter.Inc()
	}

}

// -----------------------------------------------------------------------------
// Burn‑Message Tx
// -----------------------------------------------------------------------------

func handleBurnMessageTx(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var p PayloadBurnReq
	if err := json.NewDecoder(req.Body).Decode(&p); err != nil && err != io.EOF {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	_ = req.Body.Close()

	msg := &zktypes.MsgMintBurnTradingBtc{
		MintOrBurn:      false,
		BtcValue:        p.BtcValue,
		QqAccount:       p.QqAccount,
		EncryptScalar:   p.EncryptScalar,
		TwilightAddress: p.TwilightAddress,
	}

	cosmos := getCosmosClient()
	acc, err := cosmos.Account(accountName)
	if err != nil {
		http.Error(w, "account not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := cosmos.BroadcastTx(req.Context(), acc, msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error" : "` + err.Error() + `"}`))
	} else {

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"txHash" : "` + resp.TxHash + `"}`))
		txCounter.Inc()
	}

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
