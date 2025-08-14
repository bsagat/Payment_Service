package bereke

import (
	"fmt"

	bma "github.com/bsagat/bereke-merchant-api"
	"github.com/bsagat/bereke-merchant-api/models/types"
)

var Bereke_Broker string = "BEREKE"

type BerekeClient struct {
	merchant bma.API
}

func NewClient(api_key string, mode types.Mode) (*BerekeClient, error) {
	api, err := bma.NewWithToken(api_key, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create new bma client: %v", err)
	}

	return &BerekeClient{
		merchant: api,
	}, nil
}

func (c *BerekeClient) Ping() error {
	return c.merchant.Ping()
}
