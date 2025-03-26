package block

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric/protoutil"
)

type SampleBlock struct {
	BlockHeight  uint64    `json:"blockHeight"`  // 区块高度
	PreviousHash string    `json:"previousHash"` // 前一区块哈希
	DataHash     string    `json:"dataHash"`     // 当前区块哈希
	TimeStamp    time.Time `json:"timestamp"`    // 区块时间(由于fabric中没有出块时间则采用最后一比交易作为区块时间)
	TxList       []Tx      `json:"txList"`       // 交易列表
	// Block        *common.Block
}

type Tx struct {
	Id               string                  `json:"id"`               // 交易id
	ChannelId        string                  `json:"channel_id"`       // 通道名称
	TimeStamp        time.Time               `json:"timestamp"`        // 交易时间
	HeaderType       common.HeaderType       `json:"headerType"`       // 交易类型
	ValidationCode   peer.TxValidationCode   `json:"validationCode"`   // 交易状态
	ChaincodeType    peer.ChaincodeSpec_Type `json:"chaincode_type"`   // 合约类型
	ChaincodeName    string                  `json:"chaincode_name"`   // 合约名称
	ChaincodeVersion string                  `json:"chaincodeVersion"` // 合约版本
	ChainCodeInput   []string                `json:"chain_code_input"` // 合约调用入参
	TransientMap     map[string][]byte       `json:"transientMap"`     // 合约调用Transient参数
	IsInit           bool                    `json:"isInit"`           // init
	Response         *peer.Response          `json:"response"`         // 合约调用返回值以及状态
	EventName        string                  `json:"eventName"`        // 事件名称(当没有事件抛出则为空)
	EventInput       []byte                  `json:"eventInput"`       // 事件参数内容(当没有事件抛出则为空)
	MspId            string                  `json:"endorser"`         // mspId
	MspIdBytes       string                  `json:"endorserId"`       // MSP证书
	Payload          []byte                  `json:"payload"`
}

func ParseSampleBlock(block *common.Block) (*SampleBlock, error) {
	if block.Header.Number <= 0 {
		return nil, nil
	}
	var resp = SampleBlock{
		BlockHeight:  block.Header.Number,
		PreviousHash: hex.EncodeToString(block.Header.PreviousHash),
		DataHash:     hex.EncodeToString(block.Header.DataHash),
		TxList:       make([]Tx, 0, len(block.Data.Data)),
	}
	for i, d := range block.Data.Data {
		envelope, err := protoutil.UnmarshalEnvelope(d)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalEnvelope: %w", err)
		}

		payload, err := protoutil.UnmarshalPayload(envelope.Payload)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalPayload: %w", err)
		}

		signHeader, err := protoutil.UnmarshalSignatureHeader(payload.Header.SignatureHeader)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalSignatureHeader: %w", err)
		}

		identity, err := protoutil.UnmarshalSerializedIdentity(signHeader.Creator)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalSerializedIdentity: %w", err)
		}

		channelHeader, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalChannelHeader: %w", err)
		}

		transaction, err := protoutil.UnmarshalTransaction(payload.Data)
		if err != nil {
			return nil, err
		}
		if len(transaction.Actions) == 0 {
			return nil, errors.New("at least one TransactionAction required")
		}

		chaincodePayloadAction, chaincodeAction, err := protoutil.GetPayloads(transaction.Actions[0])
		if err != nil {
			return nil, fmt.Errorf("GetPayloads: %w", err)
		}

		chaincodeEvent, err := protoutil.UnmarshalChaincodeEvents(chaincodeAction.Events)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalChaincodeEvents: %w", err)
		}
		_ = chaincodeEvent.Payload // 解析

		chaincodeProposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(chaincodePayloadAction.ChaincodeProposalPayload)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalChaincodeProposalPayload: %w", err)
		}

		chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(chaincodeProposalPayload.Input)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalChaincodeInvocationSpec: %w", err)
		}
		spec := chaincodeInvocationSpec.ChaincodeSpec

		resp.TxList = append(resp.TxList, Tx{
			Id:               channelHeader.GetTxId(),
			ChannelId:        channelHeader.GetChannelId(),
			TimeStamp:        channelHeader.GetTimestamp().AsTime(),
			HeaderType:       common.HeaderType(channelHeader.GetType()),
			ValidationCode:   peer.TxValidationCode(block.GetMetadata().GetMetadata()[common.BlockMetadataIndex_TRANSACTIONS_FILTER][i]),
			ChaincodeType:    spec.GetType(),
			ChaincodeName:    spec.ChaincodeId.GetName(),
			ChaincodeVersion: spec.ChaincodeId.GetVersion(),
			ChainCodeInput:   byteArray2StrArray(spec.GetInput().GetArgs()),
			TransientMap:     chaincodeProposalPayload.GetTransientMap(),
			IsInit:           spec.GetInput().GetIsInit(),
			Response:         chaincodeAction.GetResponse(),
			EventName:        chaincodeEvent.GetEventName(),
			EventInput:       chaincodeEvent.GetPayload(),
			MspId:            identity.GetMspid(),
			MspIdBytes:       string(identity.GetIdBytes()),
			Payload:          payload.GetData(),
		})
	}
	if len(resp.TxList) > 0 {
		resp.TimeStamp = resp.TxList[len(resp.TxList)-1].TimeStamp
	}
	return &resp, nil
}

func byteArray2StrArray(d [][]byte) []string {
	var args = make([]string, 0, len(d))
	for _, v := range d {
		args = append(args, string(v))
	}
	return args
}
