package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/scrypt"
)

var (
	flagAddName           string
	flagAddSeedPhrase     string
	flagAddPrivateKey     string
	flagAddAutoSeedPhrase bool
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add chain info",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		for i := range cfg.Wallets {
			if flagAddName == cfg.Wallets[i].Name {
				return fmt.Errorf("duplicated wallet name with index %d, please choose another one", i)
			}
		}

		var privateKeyBytes, publicKeyBytes []byte

		if flagAddPrivateKey != "" {
			pk, err := hexutil.Decode(flagAddPrivateKey)
			if err != nil {
				return fmt.Errorf("invalid private key: %w", err)
			}

			privateKey, err := crypto.ToECDSA(pk)
			if err != nil {
				return fmt.Errorf("invalid private key: %w", err)
			}

			privateKeyBytes = crypto.FromECDSA(privateKey)
			publicKeyBytes = crypto.FromECDSAPub(&privateKey.PublicKey)

		} else {
			if flagAddAutoSeedPhrase {
				flagAddSeedPhrase = generateMnemonic()
				fmt.Println("Mnemonic:", flagAddSeedPhrase)
			}
			if !bip39.IsMnemonicValid(flagAddSeedPhrase) {
				return fmt.Errorf("invalid mnemonic")
			}

			if !utils.Confirm("This will generate a new wallet from the mnemonic. Continue? [y/N]: ") {
				return fmt.Errorf("cancelled")
			}

			// === FIX: Dùng HD Wallet derivation chuẩn ===
			wallet, err := hdwallet.NewFromMnemonic(flagAddSeedPhrase)
			if err != nil {
				return err
			}

			path := accounts.DefaultBaseDerivationPath // m/44'/60'/0'/0
			account, err := wallet.Derive(path, false)
			if err != nil {
				return err
			}

			privateKey, err := wallet.PrivateKey(account)
			if err != nil {
				return err
			}

			privateKeyBytes = crypto.FromECDSA(privateKey)
			publicKeyBytes = crypto.FromECDSAPub(&privateKey.PublicKey)
		}

		password := utils.AskPassword("Enter password: ")
		retypedPassword := utils.AskPassword("Re-enter password: ")

		if password != retypedPassword {
			return fmt.Errorf("mismatched password")
		}

		salt, nonce, encryptedPrivateKeyBytes, err := encryptPrivateKey(password, privateKeyBytes)
		if err != nil {
			return err
		}

		cfg.Wallets = append(cfg.Wallets, config.WalletInfo{
			Name:                flagAddName,
			PublicKey:           publicKeyBytes,
			Salt:                salt,
			Nonce:               nonce,
			EncryptedPrivateKey: encryptedPrivateKeyBytes,
		})

		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&flagAddName, "name", "", "chain name")
	addCmd.Flags().StringVar(&flagAddSeedPhrase, "seed-phrase", "", "import from seed phrase")
	addCmd.Flags().StringVar(&flagAddPrivateKey, "private-key", "", "import from private key")
	addCmd.Flags().BoolVar(&flagAddAutoSeedPhrase, "auto", false, "automatically generate seed phrase")

	addCmd.MarkFlagsMutuallyExclusive("seed-phrase", "private-key", "auto")
	addCmd.MarkFlagsOneRequired("seed-phrase", "private-key", "auto")
	addCmd.MarkFlagRequired("name")
}

func encryptPrivateKey(password string, privateKey []byte) ([]byte, []byte, []byte, error) {
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, nil, nil, err
	}

	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, nil, nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, privateKey, nil)
	return salt, nonce, ciphertext, nil
}

func generateMnemonic() string {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		panic(err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		panic(err)
	}

	return mnemonic
}
