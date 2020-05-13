package main

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"

	clock "github.com/AlpacaLabs/go-timestamp"
	"github.com/google/uuid"
	"github.com/rs/xid"

	"github.com/AlpacaLabs/api-password/internal/configuration"
	"github.com/AlpacaLabs/api-password/internal/db"
	. "github.com/smartystreets/goconvey/convey"
)

var dbConn *sql.DB

func TestMain(m *testing.M) {
	c := configuration.LoadConfig()
	if dbc, err := c.SQLConfig.Connect(); err != nil {
		logrus.Fatalf("failed to dial account service: %v", err)
	} else {
		dbConn = dbc
	}

	code := m.Run()

	os.Exit(code)
}

func Test_CreatePasswordResetCode(t *testing.T) {
	Convey("Given a user with an email and username", t, func(c C) {
		ctx := context.TODO()

		dbClient := db.NewClient(dbConn)
		err := dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
			return tx.CreatePasswordResetCode(ctx, passwordV1.PasswordResetCode{
				Code:      uuid.New().String(),
				ExpiresAt: clock.TimeToTimestamp(time.Now().Add(time.Minute * 5)),
				AccountId: xid.New().String(),
			})
		})
		So(err, ShouldBeNil)
	})
}
