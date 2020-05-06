package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	accountV1 "github.com/AlpacaLabs/protorepo-account-go/alpacalabs/account/v1"
	"google.golang.org/grpc"

	"github.com/AlpacaLabs/api-password/internal/db"
	clock "github.com/AlpacaLabs/go-timestamp"
	authV1 "github.com/AlpacaLabs/protorepo-auth-go/alpacalabs/auth/v1"
	"github.com/badoux/checkmail"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/ttacon/libphonenumber"
)

var (
	ErrEmptyUserIdentifier = errors.New("user identifier cannot be empty; must be email, phone number, or username")
)

func (s Service) SendCodeOptions(ctx context.Context, request authV1.GetCodeOptionsRequest) (*authV1.GetCodeOptionsResponse, error) {
	var response *authV1.GetCodeOptionsResponse
	if err := s.dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
		accountIdentifier := strings.TrimSpace(request.UserIdentifier)
		if accountIdentifier == "" {
			return ErrEmptyUserIdentifier
		}

		account, err := s.getAccount(ctx, accountIdentifier)
		if err != nil {
			return err
		}

		accountID := account.Id

		if r, err := getSendOptions(ctx, account); err != nil {
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
	}); err != nil {
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

func getSendOptions(ctx context.Context, account *accountV1.Account) (*authV1.GetCodeOptionsResponse, error) {
	res := &authV1.GetCodeOptionsResponse{
		CodeOptions: &authV1.CodeOptions{},
	}

	var phoneNumbers []*authV1.PhoneNumber
	for _, p := range account.PhoneNumbers {
		phoneNumbers = append(phoneNumbers, &authV1.PhoneNumber{
			Id:        p.Id,
			CreatedAt: p.CreatedAt,
			AccountId: p.AccountId,
			// Mask each phone number (only show last two digits)
			PhoneNumber: maskPhoneNumber(p.PhoneNumber),
		})
	}

	var emailAddresses []*authV1.EmailAddress
	for _, e := range account.EmailAddresses {
		emailAddresses = append(emailAddresses, &authV1.EmailAddress{
			Id:             e.Id,
			CreatedAt:      e.CreatedAt,
			LastModifiedAt: e.LastModifiedAt,
			Deleted:        e.Deleted,
			DeletedAt:      e.DeletedAt,
			Confirmed:      e.Confirmed,
			Primary:        e.Primary,
			EmailAddress:   maskEmailAddress(e.EmailAddress),
			AccountId:      e.AccountId,
		})
	}

	return res, nil
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

func isEmailAddress(s string) bool {
	return strings.Contains(s, "@")
}

func isPhoneNumber(s string) bool {
	_, err := libphonenumber.Parse(s, "US")
	return err == nil
}

func (s *Service) getAccount(ctx context.Context, accountIdentifier string) (*accountV1.Account, error) {
	request, err := accountIdentifierToGetAccountRequest(accountIdentifier)
	if err != nil {
		return nil, err
	}

	// Dial Account service
	conn, err := grpc.Dial(s.config.AccountGRPCAddress)
	if err != nil {
		return nil, err
	}

	client := accountV1.NewAccountServiceClient(conn)

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

func (s *Service) getAccountForEmailAddress(ctx context.Context, emailAddress string) (*accountV1.Account, error) {
	// Dial Account service
	conn, err := grpc.Dial(s.config.AccountGRPCAddress)
	if err != nil {
		return nil, err
	}

	client := accountV1.NewAccountServiceClient(conn)

	res, err := client.GetAccount(ctx, &accountV1.GetAccountRequest{
		AccountIdentifier: &accountV1.GetAccountRequest_EmailAddress{
			EmailAddress: emailAddress,
		},
	})
	if err != nil {
		return nil, err
	}

	return res.Account, nil
}
