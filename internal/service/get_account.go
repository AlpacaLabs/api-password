package service

import (
	"context"
	"fmt"
	"strings"

	accountV1 "github.com/AlpacaLabs/protorepo-account-go/alpacalabs/account/v1"
	"github.com/badoux/checkmail"
	"github.com/ttacon/libphonenumber"
)

func (s *Service) getAccount(ctx context.Context, accountIdentifier string) (*accountV1.Account, error) {
	request, err := accountIdentifierToGetAccountRequest(accountIdentifier)
	if err != nil {
		return nil, err
	}

	client := accountV1.NewAccountServiceClient(s.accountConn)

	res, err := client.GetAccount(ctx, request)
	if err != nil {
		return nil, err
	}

	return res.Account, nil
}

func accountIdentifierToGetAccountRequest(accountIdentifier string) (*accountV1.GetAccountRequest, error) {
	if isEmailAddress(accountIdentifier) {
		if err := checkmail.ValidateFormat(accountIdentifier); err != nil {
			return nil, fmt.Errorf("email address has invalid format: %s", accountIdentifier)
		}
		return &accountV1.GetAccountRequest{
			AccountIdentifier: &accountV1.GetAccountRequest_EmailAddress{
				EmailAddress: accountIdentifier,
			},
		}, nil
	} else if isPhoneNumber(accountIdentifier) {
		return &accountV1.GetAccountRequest{
			AccountIdentifier: &accountV1.GetAccountRequest_PhoneNumber{
				PhoneNumber: accountIdentifier,
			},
		}, nil
	}

	return nil, fmt.Errorf("invalid account identifier: %s", accountIdentifier)
}

func isEmailAddress(s string) bool {
	return strings.Contains(s, "@")
}

func isPhoneNumber(s string) bool {
	_, err := libphonenumber.Parse(s, "US")
	return err == nil
}
