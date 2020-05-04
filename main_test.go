package main

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	clock "github.com/AlpacaLabs/go-timestamp"
	authV1 "github.com/AlpacaLabs/protorepo-auth-go/alpacalabs/auth/v1"
	"github.com/google/uuid"
	"github.com/rs/xid"

	"github.com/AlpacaLabs/password-reset/internal/config"
	"github.com/AlpacaLabs/password-reset/internal/db"
	. "github.com/smartystreets/goconvey/convey"
)

var dbConn *sql.DB

func TestMain(m *testing.M) {
	c := config.LoadConfig()
	dbConn = db.Connect(c.DBUser, c.DBPass, c.DBHost, c.DBName)

	code := m.Run()

	os.Exit(code)
}

func Test_CreatePasswordResetCode(t *testing.T) {
	Convey("Given a user with an email and username", t, func(c C) {
		ctx := context.TODO()

		dbClient := db.NewClient(dbConn)
		err := dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
			return tx.CreatePasswordResetCode(ctx, authV1.PasswordResetCode{
				Code:      uuid.New().String(),
				ExpiresAt: clock.TimeToTimestamp(time.Now().Add(time.Minute * 5)),
				AccountId: xid.New().String(),
			})
		})
		So(err, ShouldBeNil)
	})
}
