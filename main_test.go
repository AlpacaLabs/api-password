package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AlpacaLabs/api-password/internal/db/entities"

	"github.com/jackc/pgx/v4"

	"github.com/sirupsen/logrus"

	"github.com/rs/xid"

	"github.com/AlpacaLabs/api-password/internal/configuration"
	"github.com/AlpacaLabs/api-password/internal/db"
	. "github.com/smartystreets/goconvey/convey"
)

var dbConn *pgx.Conn

func TestMain(m *testing.M) {
	c := configuration.LoadConfig()
	logrus.Infof("Loaded config: %s", c)

	if dbc, err := db.Connect(c.SQLConfig); err != nil {
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

			resetCode, err := entities.NewPasswordResetCode(xid.New().String(), time.Minute*30)
			if err != nil {
				return err
			}

			return tx.CreatePasswordResetCode(ctx, resetCode)
		})
		So(err, ShouldBeNil)
	})
}
