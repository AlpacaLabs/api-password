package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/AlpacaLabs/api-password/internal/db/entities"

	accountV1 "github.com/AlpacaLabs/protorepo-account-go/alpacalabs/account/v1"

	"github.com/AlpacaLabs/api-password/internal/db"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
)

var (
	ErrEmptyUserIdentifier = errors.New("user identifier cannot be empty; must be email, phone number, or username")
)

func (s Service) GetDeliveryOptions(ctx context.Context, request passwordV1.GetDeliveryOptionsRequest) (*passwordV1.GetDeliveryOptionsResponse, error) {
	var response *passwordV1.GetDeliveryOptionsResponse
	if err := s.dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
		accountIdentifier := strings.TrimSpace(request.AccountIdentifier)
		if accountIdentifier == "" {
			return ErrEmptyUserIdentifier
		}

		account, err := s.getAccount(ctx, accountIdentifier)
		if err != nil {
			return err
		}

		accountID := account.Id

		response = getDeliveryOptions(account)

		if numOptions(response.CodeOptions) > 0 {
			return errors.New("no emails or phone numbers registered for user")
		}

		if resetCode, err := entities.NewPasswordResetCode(accountID, time.Minute*30); err != nil {
			return err
		} else {
			if err := tx.CreatePasswordResetCode(ctx, resetCode); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return response, nil
}

func numOptions(in *passwordV1.CodeDeliveryOptions) int {
	num := 0
	if in != nil {
		num += len(in.EmailAddresses)
		num += len(in.PhoneNumbers)
	}
	return num
}

func getDeliveryOptions(account *accountV1.Account) *passwordV1.GetDeliveryOptionsResponse {
	res := &passwordV1.GetDeliveryOptionsResponse{
		CodeOptions: &passwordV1.CodeDeliveryOptions{},
	}

	var phoneNumbers []*passwordV1.PhoneNumberOption
	for _, p := range account.PhoneNumbers {
		phoneNumbers = append(phoneNumbers, &passwordV1.PhoneNumberOption{
			Id:        p.Id,
			AccountId: p.AccountId,
			// Mask each phone number (only show last two digits)
			PhoneNumber: maskPhoneNumber(p.PhoneNumber),
		})
	}

	var emailAddresses []*passwordV1.EmailAddressOption
	for _, e := range account.EmailAddresses {
		emailAddresses = append(emailAddresses, &passwordV1.EmailAddressOption{
			Id:           e.Id,
			EmailAddress: maskEmailAddress(e.EmailAddress),
			AccountId:    e.AccountId,
		})
	}

	return res
}

func maskPhoneNumber(phoneNumber string) string {
	return phoneNumber[len(phoneNumber)-2:]
}

func maskEmailAddress(emailAddress string) string {
	return getMaskedEmailUser(emailAddress) + "@" + getMaskedEmailHost(emailAddress)
}

func getMaskedEmailUser(emailAddress string) string {
	splits := strings.Split(emailAddress, "@")
	user := splits[0]
	if len(user) == 1 {
		return user[0:1] + strings.Repeat("*", len(user)-1)
	}
	return user[0:2] + strings.Repeat("*", len(user)-2)
}

func getMaskedEmailHost(emailAddress string) string {
	emailSplits := strings.Split(emailAddress, "@")
	host := emailSplits[1]
	splits := strings.Split(host, ".")
	splits[0] = splits[0][0:1] + strings.Repeat("*", len(splits[0])-1)
	return strings.Join(splits, ".")
}
