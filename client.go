package bundler_client

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stackup-wallet/stackup-bundler/pkg/entrypoint/filter"
	"github.com/stackup-wallet/stackup-bundler/pkg/gas"
	"github.com/stackup-wallet/stackup-bundler/pkg/userop"
)

type EthClient interface {
	SendUserOperation(ctx context.Context, op *userop.UserOperation, entryPoint common.Address) (common.Hash, error)
	EstimateUserOperationGas(ctx context.Context, op *userop.UserOperation, entryPoint common.Address) (*gas.GasEstimates, error)
	// EstimateUserOperationGasWithOverrides is a non-spec method supported by some bundlers (e.g. Stackup)
	EstimateUserOperationGasWithOverrides(ctx context.Context, op *userop.UserOperation, entryPoint common.Address, stateOverrides map[common.Address]OverrideAccount) (*gas.GasEstimates, error)
	GetUserOperationReceipt(ctx context.Context, userOpHash common.Hash) (*filter.UserOperationReceipt, error)
	GetUserOperationByHash(ctx context.Context, userOpHash common.Hash) (*filter.HashLookupResult, error)
	SupportedEntryPoints(ctx context.Context) ([]common.Address, error)
	ChainId(ctx context.Context) (*big.Int, error)
}

type DebugClient interface {
	BundlerClearState(ctx context.Context) error
	BundlerDumpMempool(ctx context.Context, entryPoint common.Address) ([]*userop.UserOperation, error)
	BundlerSendBundleNow(ctx context.Context) (*common.Hash, error)
	BundlerSetBundlingMode(ctx context.Context, mode string) error
}

type Client interface {
	EthClient
	DebugClient
}

type RpcClient struct {
	c *rpc.Client
}

func Dial(rawurl string) (Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

func NewClient(c *rpc.Client) Client {
	return &RpcClient{c}
}

func (c *RpcClient) SendUserOperation(ctx context.Context, op *userop.UserOperation, entryPoint common.Address) (common.Hash, error) {
	var result common.Hash
	err := c.c.CallContext(ctx, &result, "eth_sendUserOperation", op, entryPoint)
	return result, err
}

func (c *RpcClient) EstimateUserOperationGas(ctx context.Context, op *userop.UserOperation, entryPoint common.Address) (*gas.GasEstimates, error) {
	var estimate gas.GasEstimates
	err := c.c.CallContext(ctx, &estimate, "eth_estimateUserOperationGas", op, entryPoint)
	if err != nil {
		return nil, err
	}
	return &estimate, nil
}

func (c *RpcClient) EstimateUserOperationGasWithOverrides(ctx context.Context, op *userop.UserOperation, entryPoint common.Address, stateOverrides map[common.Address]OverrideAccount) (*gas.GasEstimates, error) {
	var estimate gas.GasEstimates
	err := c.c.CallContext(ctx, &estimate, "eth_estimateUserOperationGas", op, entryPoint, stateOverrides)
	if err != nil {
		return nil, err
	}
	return &estimate, nil
}

func (c *RpcClient) GetUserOperationReceipt(ctx context.Context, userOpHash common.Hash) (*filter.UserOperationReceipt, error) {
	var receipt filter.UserOperationReceipt
	err := c.c.CallContext(ctx, &receipt, "eth_getUserOperationReceipt", userOpHash)
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (c *RpcClient) GetUserOperationByHash(ctx context.Context, userOpHash common.Hash) (*filter.HashLookupResult, error) {
	var op filter.HashLookupResult
	err := c.c.CallContext(ctx, &op, "eth_getUserOperationByHash", userOpHash)
	if err != nil {
		return nil, err
	}
	return &op, nil
}

func (c *RpcClient) SupportedEntryPoints(ctx context.Context) ([]common.Address, error) {
	var entryPoints []common.Address
	err := c.c.CallContext(ctx, &entryPoints, "eth_supportedEntryPoints", []interface{}{}...)
	if err != nil {
		return nil, err
	}
	return entryPoints, nil
}

func (c *RpcClient) ChainId(ctx context.Context) (*big.Int, error) {
	var result hexutil.Big
	err := c.c.CallContext(ctx, &result, "eth_chainId", []interface{}{}...)
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), nil
}

func (c *RpcClient) BundlerClearState(ctx context.Context) error {
	return c.c.CallContext(ctx, nil, "debug_bundler_clearState", []interface{}{}...)
}

func (c *RpcClient) BundlerDumpMempool(ctx context.Context, entryPoint common.Address) ([]*userop.UserOperation, error) {
	var ops []*UserOperation
	err := c.c.CallContext(ctx, &ops, "debug_bundler_dumpMempool", entryPoint)
	if err != nil {
		return nil, err
	}
	uops := make([]*userop.UserOperation, len(ops))
	for i, op := range ops {
		uops[i] = op.ToUserOperation()
	}
	return uops, nil
}

func (c *RpcClient) BundlerSendBundleNow(ctx context.Context) (*common.Hash, error) {
	var result string
	err := c.c.CallContext(ctx, &result, "debug_bundler_sendBundleNow", []interface{}{}...)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	hash := common.HexToHash(result)
	return &hash, nil
}

func (c *RpcClient) BundlerSetBundlingMode(ctx context.Context, mode string) error {
	return c.c.CallContext(ctx, nil, "debug_bundler_setBundlingMode", mode)
}

type UserOperation struct {
	Sender               common.Address `json:"sender"`
	Nonce                *hexutil.Big   `json:"nonce"`
	InitCode             hexutil.Bytes  `json:"initCode"`
	CallData             hexutil.Bytes  `json:"callData"`
	CallGasLimit         *hexutil.Big   `json:"callGasLimit"`
	VerificationGasLimit *hexutil.Big   `json:"verificationGasLimit"`
	PreVerificationGas   *hexutil.Big   `json:"preVerificationGas"`
	MaxFeePerGas         *hexutil.Big   `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *hexutil.Big   `json:"maxPriorityFeePerGas"`
	PaymasterAndData     hexutil.Bytes  `json:"paymasterAndData"`
	Signature            hexutil.Bytes  `json:"signature"`
}

func (uo *UserOperation) ToUserOperation() *userop.UserOperation {
	if uo == nil {
		return nil
	}
	return &userop.UserOperation{
		Sender:               uo.Sender,
		Nonce:                uo.Nonce.ToInt(),
		InitCode:             uo.InitCode,
		CallData:             uo.CallData,
		CallGasLimit:         uo.CallGasLimit.ToInt(),
		VerificationGasLimit: uo.VerificationGasLimit.ToInt(),
		PreVerificationGas:   uo.PreVerificationGas.ToInt(),
		MaxFeePerGas:         uo.MaxFeePerGas.ToInt(),
		MaxPriorityFeePerGas: uo.MaxPriorityFeePerGas.ToInt(),
		PaymasterAndData:     uo.PaymasterAndData,
		Signature:            uo.Signature,
	}
}

type OverrideAccount struct {
	Nonce     *hexutil.Uint64              `json:"nonce"`
	Code      *hexutil.Bytes               `json:"code"`
	Balance   *hexutil.Big                 `json:"balance"`
	State     *map[common.Hash]common.Hash `json:"state"`
	StateDiff *map[common.Hash]common.Hash `json:"stateDiff"`
}
