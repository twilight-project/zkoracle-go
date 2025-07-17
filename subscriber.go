package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

func getRequest(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return body
}

func nyksSubscriber(blockHeight uint64) {
	nyksUrl := fmt.Sprintf("%v", viper.Get("nyks_url"))
	var curHeight uint64
	var err error
	if blockHeight == 0 {
		curHeight, err = readUint64FromFile("height.txt")
		if err != nil {
			fmt.Println("error reading height from file : ", err)
			curHeight = 1
		}
	} else {
		curHeight = blockHeight
	}

	for {

		url := nyksUrl + "/cosmos/base/tendermint/v1beta1/blocks/latest"

		body := getRequest(url)
		block := RespNyksBlock{}
		err := json.Unmarshal(body, &block)
		if err != nil {
			fmt.Println("error unmarshalling nyks block : ", err)
			continue
		}
		heightStr := block.Block.Header.Height
		latestHeight, err := strconv.ParseUint(heightStr, 10, 64)

		if curHeight > latestHeight {
			time.Sleep(10 * time.Second)
			continue
		}

		url = nyksUrl + "/cosmos/tx/v1beta1/txs/block/" + strconv.FormatUint(curHeight, 10)
		body = getRequest(url)

		result, err := filterBlockTx(body, curHeight)
		if err != nil {
			continue
		}

		jsonBytes, err := json.Marshal(result)
		if err != nil {
			log.Println(err)
		}

		writeUint64ToFile("height.txt", curHeight)

		curHeight = curHeight + 1

		WsHub.broadcast <- jsonBytes

	}

}

func writeUint64ToFile(filename string, num uint64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = binary.Write(file, binary.LittleEndian, num)
	if err != nil {
		return err
	}

	return nil
}

func readUint64FromFile(filename string) (uint64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var num uint64
	err = binary.Read(file, binary.LittleEndian, &num)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func filterBlockTx(data []byte, height uint64) (ResultPubSub, error) {
	a := RespNyksBlockWithTx{}
	err := json.Unmarshal(data, &a)
	if err != nil {
		fmt.Println("error unmarshalling nyks block : ", err)
		return ResultPubSub{}, err
	}

	if a.Code == 3 {
		return ResultPubSub{
			"",
			strconv.FormatUint(height, 10),
			[]Message{},
		}, nil
	}

	filteredTxs := []Message{}
	for i, tx := range a.Txs {
		for _, msg := range tx.Body.Messages {
			if msg.Type == "/twilightproject.nyks.zkos.MsgTransferTx" || msg.Type == "/twilightproject.nyks.zkos.MsgMintBurnTradingBtc" {
				msg.TxId = tx_hash(a.Block.Data.Txs[i])
				confirmed := checkstatus(msg.TxId)
				if confirmed == true {
					filteredTxs = append(filteredTxs, msg)
				}
			}
		}
	}
	result := ResultPubSub{
		a.Block_id.Hash,
		a.Block.Header.Height,
		filteredTxs,
	}
	return result, nil
}

func checkstatus(txid string) bool {
	nyksUrl := fmt.Sprintf("%v", viper.Get("nyks_url"))
	url := nyksUrl + "/cosmos/tx/v1beta1/txs/" + txid

	body := getRequest(url)
	txDetails := TxDetailsResp{}
	_ = json.Unmarshal(body, &txDetails)
	fmt.Println("inside check status : ", txDetails.Tx_response.Code)
	if txDetails.Tx_response.Code == 0 {
		return true
	}
	return false
}

func tx_hash(tx string) string {
	bytes, err := base64.StdEncoding.DecodeString(tx)
	if err != nil {
		log.Println("Decoding failed: ", err)
	}

	hasher := sha256.New()
	hasher.Write(bytes)
	hash := hasher.Sum(nil)

	return hex.EncodeToString(hash)
}
