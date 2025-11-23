package fabricclient

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type Client struct {
	gw       *gateway.Gateway
	network  *gateway.Network
	contract *gateway.Contract
}

func NewClient(configPath, channelName, contractName, mspID, certPath, keyPath string) (*Client, error) {
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet, mspID, certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to populate wallet: %v", err)
		}
	}

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(configPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %v", err)
	}

	network, err := gw.GetNetwork(channelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %v", err)
	}

	contract := network.GetContract(contractName)

	return &Client{
		gw:       gw,
		network:  network,
		contract: contract,
	}, nil
}

func (c *Client) SubmitTransaction(name string, args ...string) ([]byte, error) {
	return c.contract.SubmitTransaction(name, args...)
}

func (c *Client) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	return c.contract.EvaluateTransaction(name, args...)
}

func (c *Client) RegisterChaincodeEventListener(eventName string) (<-chan *gateway.ChaincodeEvent, error) {
	reg, notifier, err := c.contract.RegisterEvent(eventName)
	if err != nil {
		return nil, err
	}
	// Note: In a real app we'd need to manage unregistration.
	// For now we return the channel.
	_ = reg
	return notifier, nil
}

func (c *Client) Close() {
	c.gw.Close()
}

func populateWallet(wallet *gateway.Wallet, mspID, certPath, keyPath string) error {
	cert, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	key, err := os.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity(mspID, string(cert), string(key))

	return wallet.Put("appUser", identity)
}
