package request

import (
	"bytes"
	"context"
	"derand-cli/backend"
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spf13/cobra"
)

var (
	flagProfileID      uint64
	flagProfileVersion uint32
	flagSeed           uint64
	flagDelayFactor    uint16
	flagMaxDelay       uint16
	flagBy             int
)

var Cmd = &cobra.Command{
	Use:   "request",
	Short: "request a random number",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		if currentChain.DeRand == nil {
			return fmt.Errorf("derand has not been deployed yet, please run `derand deploy` or `derand chain setup` first!")
		}

		var wallet *config.WalletInfo
		if flagBy == -1 {
			wallet, err = cfg.GetCurrentWallet()
			if err != nil {
				return err
			}
		} else {
			wallet, err = cfg.GetWallet(flagBy)
			if err != nil {
				return err
			}
		}

		backend, err := cfg.GetCurrentChainBackend()
		if err != nil {
			return err
		}

		auth, err := wallet.GetDefaultTxOpts(currentChain.ChainID, utils.AskPassword("Enter password: "))
		if err != nil {
			return err
		}

		derand, err := gen.NewDeRand(currentChain.DeRand.Address, backend)
		if err != nil {
			return fmt.Errorf("failed to initialize derand: %w", err)
		}

		lastBlockRlp, err := getRlpLastBlock(cmd.Context(), backend)
		if err != nil {
			return err
		}

		utils.PrintTitle("Requesting")

		auth.GasLimit = 400000
		requestTx, err := derand.RequestRandomNumber(
			auth,
			flagProfileID,
			flagProfileVersion,
			big.NewInt(int64(flagSeed)),
			1,
			flagMaxDelay,
			common.Address{},
			0,
			lastBlockRlp,
		)
		if err != nil {
			name, args, err := utils.DecodeCustomError(gen.DeRandABI, err)
			if err != nil {
				return err
			}

			return fmt.Errorf("%s%v", name, args)
		}

		utils.PrintSubtitle("Transaction:", requestTx.Hash())

		receipt, err := bind.WaitMined(cmd.Context(), backend, requestTx)
		if err != nil {
			return err
		}

		requestID := -1
		for _, vLog := range receipt.Logs {
			ev, err := derand.ParseRequestRandomNumber(*vLog)
			if err == nil {
				requestID = int(ev.RequestId)
				break
			}
		}

		if requestID == -1 {
			return fmt.Errorf("not found the event to catch request id")
		}

		utils.PrintSubtitle("Request ID:", requestID)

		return nil
	},
}

func init() {
	Cmd.Flags().Uint64Var(&flagProfileID, "profile-id", 0, "profile id")
	Cmd.Flags().Uint32Var(&flagProfileVersion, "profile-version", 0, "profile version")
	Cmd.Flags().Uint64Var(&flagSeed, "seed", 0, "seed")
	Cmd.Flags().Uint16Var(&flagDelayFactor, "delay-factor", 1, "delay factor")
	Cmd.Flags().Uint16Var(&flagMaxDelay, "max-delay", 20, "max delay")
	Cmd.Flags().IntVar(&flagBy, "by", -1, "request on behalf of another wallet")

	Cmd.AddCommand(infoCmd)
}

func getRlpLastBlock(ctx context.Context, backend *backend.BackendPool) ([]byte, error) {
	block, err := backend.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, fmt.Errorf("not found any block")
	}

	header := block.Header()
	if header == nil {
		return nil, err
	}

	rlpBytes, err := rlp.EncodeToBytes(header)
	if err != nil {
		return nil, err
	}

	blockhash := block.Hash()
	hash := crypto.Keccak256Hash(rlpBytes)
	if !bytes.Equal(hash[:], blockhash[:]) {
		return nil, fmt.Errorf("invalid rlp hash")
	}

	blocknumber, timestamp, err := extractBlockNumberAndTimestamp(rlpBytes)
	if err != nil {
		return nil, err
	}

	utils.PrintTitle("Checkpoint RLP Block")
	utils.PrintSubtitle("Block number:", blocknumber)
	utils.PrintSubtitle("Timestamp:", timestamp)
	utils.PrintSubtitle("Block hash:", blockhash)

	if blocknumber.Cmp(block.Number()) != 0 {
		return nil, fmt.Errorf("mismatched block number in rlp, expected %s != got %s", blocknumber, block.Number())
	}

	if timestamp != block.Time() {
		return nil, fmt.Errorf("mismatched block timestamp in rlp, expected %d != got %d", timestamp, block.Time())
	}

	return rlpBytes, nil
}

func extractBlockNumberAndTimestamp(headerRlp []byte) (*big.Int, uint64, error) {
	offset :=
		3 + // list prefix
			33 + // parentHash
			33 + // uncleHash
			21 + // coinbase
			33 + // stateRoot
			33 + // txRoot
			33 + // receiptRoot
			259 // bloom

	// difficulty
	sz, err := skipRLP(headerRlp[offset:])
	if err != nil {
		return nil, 0, err
	}
	offset += sz

	// blockNumber
	blockNumber, err := parseBigIntRlp(headerRlp[offset:])
	if err != nil {
		return nil, 0, err
	}

	sz, err = skipRLP(headerRlp[offset:])
	if err != nil {
		return nil, 0, err
	}
	offset += sz

	// gasLimit
	sz, err = skipRLP(headerRlp[offset:])
	if err != nil {
		return nil, 0, err
	}
	offset += sz

	// gasUsed
	sz, err = skipRLP(headerRlp[offset:])
	if err != nil {
		return nil, 0, err
	}
	offset += sz

	// timestamp
	timestamp, err := parseUint64Rlp(headerRlp[offset:])
	if err != nil {
		return nil, 0, err
	}

	return blockNumber, timestamp, nil
}

func skipRLP(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, fmt.Errorf("empty rlp")
	}

	p := b[0]

	// single byte
	if p <= 0x7f {
		return 1, nil
	}

	// short string
	if p <= 0xb7 {
		l := int(p - 0x80)

		if len(b) < 1+l {
			return 0, fmt.Errorf("truncated short string")
		}

		return 1 + l, nil
	}

	// long string
	if p <= 0xbf {
		lenOfLen := int(p - 0xb7)

		if len(b) < 1+lenOfLen {
			return 0, fmt.Errorf("truncated long string prefix")
		}

		l := 0
		for i := range lenOfLen {
			l = (l << 8) | int(b[1+i])
		}

		if len(b) < 1+lenOfLen+l {
			return 0, fmt.Errorf("truncated long string")
		}

		return 1 + lenOfLen + l, nil
	}

	// short list
	if p <= 0xf7 {
		l := int(p - 0xc0)

		if len(b) < 1+l {
			return 0, fmt.Errorf("truncated short list")
		}

		return 1 + l, nil
	}

	// long list
	lenOfLen := int(p - 0xf7)

	if len(b) < 1+lenOfLen {
		return 0, fmt.Errorf("truncated long list prefix")
	}

	l := 0
	for i := range lenOfLen {
		l = (l << 8) | int(b[1+i])
	}

	if len(b) < 1+lenOfLen+l {
		return 0, fmt.Errorf("truncated long list")
	}

	return 1 + lenOfLen + l, nil
}

func parseUint64Rlp(b []byte) (uint64, error) {
	if len(b) == 0 {
		return 0, fmt.Errorf("empty rlp")
	}

	size, err := skipRLP(b)
	if err != nil {
		return 0, err
	}

	p := b[0]

	var start int

	switch {
	case p <= 0x7f:
		start = 0

	case p <= 0xb7:
		start = 1

	case p <= 0xbf:
		start = 1 + int(p-0xb7)

	case p <= 0xf7:
		start = 1

	default:
		start = 1 + int(p-0xf7)
	}

	payloadLen := size - start

	if payloadLen > 8 {
		return 0, fmt.Errorf("rlp payload is too big for uint64")
	}

	var out uint64
	for _, v := range b[start:size] {
		out = (out << 8) | uint64(v)
	}

	return out, nil
}

func parseBigIntRlp(b []byte) (*big.Int, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("empty rlp")
	}

	size, err := skipRLP(b)
	if err != nil {
		return nil, err
	}

	p := b[0]

	var start int

	switch {
	case p <= 0x7f:
		start = 0

	case p <= 0xb7:
		start = 1

	case p <= 0xbf:
		start = 1 + int(p-0xb7)

	case p <= 0xf7:
		start = 1

	default:
		start = 1 + int(p-0xf7)
	}

	return new(big.Int).SetBytes(b[start:size]), nil
}
