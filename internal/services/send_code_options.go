package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	clock "github.com/AlpacaLabs/go-timestamp"
	"github.com/AlpacaLabs/password-reset/internal/db"
	authV1 "github.com/AlpacaLabs/protorepo-auth-go/alpacalabs/auth/v1"
	"github.com/badoux/checkmail"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/ttacon/libphonenumber"
)

var (
	ErrEmptyUserIdentifier = errors.New("user identifier cannot be empty; must be email, phone number, or username")
)

func (s Service) SendCodeOptions(ctx context.Context, request authV1.GetCodeOptionsRequest) (response *authV1.GetCodeOptionsResponse, err error) {
	err = s.dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
		accountIdentifier := strings.TrimSpace(request.UserIdentifier)
		if accountIdentifier == "" {
			return ErrEmptyUserIdentifier
		}

		var accountID string
		if accountID, err = getAccountIdForAccount(ctx, tx, accountIdentifier); err != nil || accountID == "" {
			// We deliberately do not leak if email is not found
			return nil
		}

		if r, err := getSendOptions(ctx, accountID, tx); err != nil {
			// We deliberately do not leak if email is not found
			return nil
		} else {
			response = r
		}

		if numOptions(response.CodeOptions) == 1 {
			// TODO actually send an email
			log.Println("Fake sending an email")

			expiration := time.Now().Add(time.Minute * 30)
			if resetCode, err := newPasswordResetCode(accountID, expiration); err != nil {
				return err
			} else {
				if err := tx.CreatePasswordResetCode(ctx, resetCode); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func newPasswordResetCode(accountID string, expiration time.Time) (authV1.PasswordResetCode, error) {
	var empty authV1.PasswordResetCode
	var code string
	if u, err := uuid.NewRandom(); err != nil {
		return empty, err
	} else {
		code = u.String()
		log.Debugf("Generated password reset code: %s", code)
	}

	return authV1.PasswordResetCode{
		Code:      code,
		ExpiresAt: clock.TimeToTimestamp(expiration),
		AccountId: accountID,
	}, nil
}

func numOptions(in *authV1.CodeOptions) int {
	num := 0
	if in != nil {
		num += len(in.EmailAddresses)
		num += len(in.PhoneNumbers)
	}
	return num
}

func getSendOptions(ctx context.Context, accountID string, tx db.Transaction) (*authV1.GetCodeOptionsResponse, error) {
	res := &authV1.GetCodeOptionsResponse{
		CodeOptions: &authV1.CodeOptions{},
	}

	if phoneNumbers, err := tx.GetPhoneNumbersForAccount(ctx, accountID); err != nil {
		return nil, err
	} else {
		res.CodeOptions.PhoneNumbers = phoneNumbers
		// Mask each phone number (only show last two digits)
		for _, p := range res.CodeOptions.PhoneNumbers {
			p.PhoneNumber = p.PhoneNumber[len(p.PhoneNumber)-2:]
		}
	}

	if emailAddresses, err := tx.GetConfirmedEmailAddressesForAccountID(ctx, accountID); err != nil {
		return nil, err
	} else {
		res.CodeOptions.EmailAddresses = emailAddresses
		// Mask each email address
		for _, e := range res.CodeOptions.EmailAddresses {
			e.EmailAddress = maskEmailAddress(e.EmailAddress)
		}
	}

	return res, nil
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

func getAccountIdForAccount(ctx context.Context, tx db.Transaction, accountIdentifier string) (string, error) {
	if isEmailAddress(accountIdentifier) {
		if err := checkmail.ValidateFormat(accountIdentifier); err != nil {
			return "", fmt.Errorf("email address has invalid format: %s", accountIdentifier)
		}
		return tx.GetAccountIDForEmailAddress(ctx, accountIdentifier)
	} else if isPhoneNumber(accountIdentifier) {
		phoneNumber, err := tx.GetPhoneNumber(ctx, accountIdentifier)
		if err != nil {
			return "", err
		}
		return phoneNumber.AccountId, nil
	} else {
		return tx.GetAccountIDForUsername(ctx, accountIdentifier)
	}
}

func isEmailAddress(s string) bool {
	return strings.Contains(s, "@")
}

func isPhoneNumber(s string) bool {
	_, err := libphonenumber.Parse(s, "US")
	return err == nil
}
