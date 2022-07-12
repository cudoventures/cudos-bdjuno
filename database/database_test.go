package database_test

import (
	"testing"
	"time"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/forbole/bdjuno/v2/utils"

	juno "github.com/forbole/juno/v2/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/stretchr/testify/suite"
)

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DbTestSuite))
}

type DbTestSuite struct {
	suite.Suite

	database *database.Db
}

func (suite *DbTestSuite) SetupTest() {
	db, _ := utils.NewTestDb(&suite.Suite, "dbTest")
	suite.database = db
}

// TODO TEST IF STILL CAUSING ISSUES WITHOUT IT

// getBlock builds, stores and returns a block for the provided height
func (suite *DbTestSuite) getBlock(height int64) *juno.Block {
	validator := suite.getValidator(
		"cosmosvalcons1qqqqrezrl53hujmpdch6d805ac75n220ku09rl",
		"cosmosvaloper1rcp29q3hpd246n6qak7jluqep4v006cdsc2kkl",
		"cosmosvalconspub1zcjduepq7mft6gfls57a0a42d7uhx656cckhfvtrlmw744jv4q0mvlv0dypskehfk8",
	)

	addr, err := sdk.ConsAddressFromBech32(validator.GetConsAddr())
	suite.Require().NoError(err)

	tmBlock := &tmctypes.ResultBlock{
		BlockID: tmtypes.BlockID{},
		Block: &tmtypes.Block{
			Header: tmtypes.Header{
				Version:            tmversion.Consensus{},
				ChainID:            "",
				Height:             height,
				Time:               time.Now(),
				LastBlockID:        tmtypes.BlockID{},
				LastCommitHash:     nil,
				DataHash:           nil,
				ValidatorsHash:     []byte("hash"),
				NextValidatorsHash: nil,
				ConsensusHash:      nil,
				AppHash:            nil,
				LastResultsHash:    nil,
				EvidenceHash:       nil,
				ProposerAddress:    tmtypes.Address(addr.Bytes()),
			},
			Data:     tmtypes.Data{},
			Evidence: tmtypes.EvidenceData{},
			LastCommit: &tmtypes.Commit{
				Height:     height - 1,
				Round:      1,
				BlockID:    tmtypes.BlockID{},
				Signatures: nil,
			},
		},
	}

	block := juno.NewBlockFromTmBlock(tmBlock, 10000)
	err = suite.database.SaveBlock(block)
	suite.Require().NoError(err)
	return block
}

// getValidator stores inside the database a validator having the given
// consensus address, validator address and validator public key
func (suite *DbTestSuite) getValidator(consAddr, valAddr, pubkey string) types.Validator {
	selfDelegation := suite.getAccount("cosmos1z4hfrxvlgl4s8u4n5ngjcw8kdqrcv43599amxs")

	maxRate := sdk.NewDec(10)
	maxChangeRate := sdk.NewDec(20)

	validator := types.NewValidator(
		consAddr,
		valAddr,
		pubkey,
		selfDelegation.String(),
		&maxChangeRate,
		&maxRate,
		1,
	)
	err := suite.database.SaveValidatorData(validator)
	suite.Require().NoError(err)

	return validator
}

// getAccount saves inside the database an account having the given address
func (suite *DbTestSuite) getAccount(addr string) sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(addr)
	suite.Require().NoError(err)

	_, err = suite.database.Sql.Exec(`INSERT INTO account (address) VALUES ($1) ON CONFLICT DO NOTHING`, delegator.String())
	suite.Require().NoError(err)

	return delegator
}
