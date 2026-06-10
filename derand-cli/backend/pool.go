package backend

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

var _ bind.ContractBackend = (*BackendPool)(nil)

type BackendPool struct {
	clients   []*ethclient.Client
	wsClients []*ethclient.Client
	mu        sync.RWMutex
}

func NewPool(rpcs []string, wsRpcs []string) (*BackendPool, error) {
	if len(rpcs) == 0 {
		return nil, fmt.Errorf("no rpc endpoints provided")
	}

	clients := make([]*ethclient.Client, len(rpcs))
	wsClients := make([]*ethclient.Client, len(wsRpcs))

	for i, rpcURL := range rpcs {
		c, err := ethclient.Dial(rpcURL)
		if err != nil {
			for j := range i {
				if clients[j] != nil {
					clients[j].Close()
				}
			}
			return nil, err
		}
		clients[i] = c
	}

	for i, wsRpcURL := range wsRpcs {
		c, err := ethclient.Dial(wsRpcURL)
		if err != nil {
			for j := range rpcs {
				if clients[j] != nil {
					clients[j].Close()
				}
			}
			for j := range i {
				if wsClients[j] != nil {
					wsClients[j].Close()
				}
			}
			return nil, err
		}
		wsClients[i] = c
	}

	return &BackendPool{clients: clients, wsClients: wsClients}, nil
}

func (p *BackendPool) Close() {
	for i := range p.clients {
		if p.clients[i] != nil {
			p.clients[i].Close()
		}
		if p.wsClients[i] != nil {
			p.wsClients[i].Close()
		}
	}
}

func (p *BackendPool) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return call(p, func(c *ethclient.Client) ([]byte, error) {
		return c.CodeAt(ctx, contract, blockNumber)
	})
}

func (p *BackendPool) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return call(p, func(c *ethclient.Client) ([]byte, error) {
		return c.CallContract(ctx, msg, blockNumber)
	})
}

func (p *BackendPool) PendingCodeAt(ctx context.Context, contract common.Address) ([]byte, error) {
	return call(p, func(c *ethclient.Client) ([]byte, error) {
		return c.PendingCodeAt(ctx, contract)
	})
}

func (p *BackendPool) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	return call(p, func(c *ethclient.Client) ([]byte, error) {
		return c.PendingCallContract(ctx, msg)
	})
}

func (p *BackendPool) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return call(p, func(c *ethclient.Client) (uint64, error) {
		return c.PendingNonceAt(ctx, account)
	})
}

func (p *BackendPool) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return call(p, func(c *ethclient.Client) (uint64, error) {
		return c.EstimateGas(ctx, msg)
	})
}

func (p *BackendPool) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	var lastErr error

	for i := range p.clients {
		if err := p.clients[i].SendTransaction(ctx, tx); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	return lastErr
}

func (p *BackendPool) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	return call(p, func(c *ethclient.Client) ([]types.Log, error) {
		return c.FilterLogs(ctx, q)
	})
}

func (p *BackendPool) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	return callWS(p, func(c *ethclient.Client) (ethereum.Subscription, error) {
		return c.SubscribeFilterLogs(ctx, q, ch)
	})
}

func (p *BackendPool) ChainID(ctx context.Context) (*big.Int, error) {
	return call(p, func(c *ethclient.Client) (*big.Int, error) {
		return c.ChainID(ctx)
	})
}

func (p *BackendPool) NetworkID(ctx context.Context) (*big.Int, error) {
	return call(p, func(c *ethclient.Client) (*big.Int, error) {
		return c.NetworkID(ctx)
	})
}

func (p *BackendPool) BlockNumber(ctx context.Context) (uint64, error) {
	return call(p, func(c *ethclient.Client) (uint64, error) {
		return c.BlockNumber(ctx)
	})
}

func (p *BackendPool) PeerCount(ctx context.Context) (uint64, error) {
	return call(p, func(c *ethclient.Client) (uint64, error) {
		return c.PeerCount(ctx)
	})
}

func (p *BackendPool) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return call(p, func(c *ethclient.Client) (*types.Header, error) {
		return c.HeaderByHash(ctx, hash)
	})
}

func (p *BackendPool) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return call(p, func(c *ethclient.Client) (*types.Header, error) {
		return c.HeaderByNumber(ctx, number)
	})
}

func (p *BackendPool) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return call(p, func(c *ethclient.Client) (*types.Block, error) {
		return c.BlockByHash(ctx, hash)
	})
}

func (p *BackendPool) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return call(p, func(c *ethclient.Client) (*types.Block, error) {
		return c.BlockByNumber(ctx, number)
	})
}

func (p *BackendPool) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return call(p, func(c *ethclient.Client) (*big.Int, error) {
		return c.BalanceAt(ctx, account, blockNumber)
	})
}

func (p *BackendPool) BalanceAtHash(ctx context.Context, account common.Address, blockHash common.Hash) (*big.Int, error) {
	return call(p, func(c *ethclient.Client) (*big.Int, error) {
		return c.BalanceAtHash(ctx, account, blockHash)
	})
}

func (p *BackendPool) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	return call(p, func(c *ethclient.Client) ([]byte, error) {
		return c.StorageAt(ctx, account, key, blockNumber)
	})
}

func (p *BackendPool) StorageAtHash(ctx context.Context, account common.Address, key common.Hash, blockHash common.Hash) ([]byte, error) {
	return call(p, func(c *ethclient.Client) ([]byte, error) {
		return c.StorageAtHash(ctx, account, key, blockHash)
	})
}

func (p *BackendPool) CodeAtHash(ctx context.Context, account common.Address, blockHash common.Hash) ([]byte, error) {
	return call(p, func(c *ethclient.Client) ([]byte, error) {
		return c.CodeAtHash(ctx, account, blockHash)
	})
}

func (p *BackendPool) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return call(p, func(c *ethclient.Client) (uint64, error) {
		return c.NonceAt(ctx, account, blockNumber)
	})
}

func (p *BackendPool) NonceAtHash(ctx context.Context, account common.Address, blockHash common.Hash) (uint64, error) {
	return call(p, func(c *ethclient.Client) (uint64, error) {
		return c.NonceAtHash(ctx, account, blockHash)
	})
}

func (p *BackendPool) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	return call(p, func(c *ethclient.Client) (*big.Int, error) {
		return c.PendingBalanceAt(ctx, account)
	})
}

func (p *BackendPool) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error) {
	return call(p, func(c *ethclient.Client) ([]byte, error) {
		return c.PendingStorageAt(ctx, account, key)
	})
}

func (p *BackendPool) PendingTransactionCount(ctx context.Context) (uint, error) {
	return call(p, func(c *ethclient.Client) (uint, error) {
		return c.PendingTransactionCount(ctx)
	})
}

func (p *BackendPool) EstimateGasAtHash(ctx context.Context, msg ethereum.CallMsg, blockHash common.Hash) (uint64, error) {
	return call(p, func(c *ethclient.Client) (uint64, error) {
		return c.EstimateGas(ctx, msg)
	})
}

func (p *BackendPool) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return call(p, func(c *ethclient.Client) (*big.Int, error) {
		return c.SuggestGasPrice(ctx)
	})
}

func (p *BackendPool) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return call(p, func(c *ethclient.Client) (*big.Int, error) {
		return c.SuggestGasTipCap(ctx)
	})
}

func (p *BackendPool) BlobBaseFee(ctx context.Context) (*big.Int, error) {
	return call(p, func(c *ethclient.Client) (*big.Int, error) {
		return c.BlobBaseFee(ctx)
	})
}

func (p *BackendPool) CallContractAtHash(ctx context.Context, msg ethereum.CallMsg, blockHash common.Hash) ([]byte, error) {
	return call(p, func(c *ethclient.Client) ([]byte, error) {
		return c.CallContractAtHash(ctx, msg, blockHash)
	})
}

func (p *BackendPool) BlockReceipts(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) ([]*types.Receipt, error) {
	return call(p, func(c *ethclient.Client) ([]*types.Receipt, error) {
		return c.BlockReceipts(ctx, blockNrOrHash)
	})
}

func (p *BackendPool) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	return call(p, func(c *ethclient.Client) (*ethereum.SyncProgress, error) {
		return c.SyncProgress(ctx)
	})
}

func (p *BackendPool) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	var lastErr error

	for i := range p.clients {
		tx, pending, err := p.clients[i].TransactionByHash(ctx, hash)
		if err == nil {
			return tx, pending, nil
		}
		lastErr = err
	}

	return nil, false, lastErr
}

func (p *BackendPool) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return call(p, func(c *ethclient.Client) (*types.Receipt, error) {
		return c.TransactionReceipt(ctx, txHash)
	})
}

func (p *BackendPool) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	return call(p, func(c *ethclient.Client) (uint, error) {
		return c.TransactionCount(ctx, blockHash)
	})
}

func (p *BackendPool) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error) {
	return call(p, func(c *ethclient.Client) (*types.Transaction, error) {
		return c.TransactionInBlock(ctx, blockHash, index)
	})
}

func (p *BackendPool) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error) {
	return call(p, func(c *ethclient.Client) (common.Address, error) {
		return c.TransactionSender(ctx, tx, block, index)
	})
}

func (p *BackendPool) SubscribeTransactionReceipts(ctx context.Context, q *ethereum.TransactionReceiptsQuery, ch chan<- []*types.Receipt) (ethereum.Subscription, error) {
	return callWS(p, func(c *ethclient.Client) (ethereum.Subscription, error) {
		return c.SubscribeTransactionReceipts(ctx, q, ch)
	})
}

func call[T any](p *BackendPool, fn func(*ethclient.Client) (T, error)) (T, error) {
	var zero T
	var lastErr error

	for i := range p.clients {
		v, err := fn(p.clients[i])
		if err == nil {
			return v, nil
		}
		lastErr = err
	}

	return zero, lastErr
}

func callWS[T any](p *BackendPool, fn func(*ethclient.Client) (T, error)) (T, error) {
	var zero T
	var lastErr error

	for i := range p.wsClients {
		v, err := fn(p.wsClients[i])
		if err == nil {
			return v, nil
		}
		lastErr = err
	}

	return zero, lastErr
}
