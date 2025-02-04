package gov

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/forbole/bdjuno/v2/database"

	govsource "github.com/forbole/bdjuno/v2/modules/gov/source"

	"github.com/forbole/juno/v2/modules"
)

var (
	_ modules.Module        = &Module{}
	_ modules.GenesisModule = &Module{}
	_ modules.BlockModule   = &Module{}
	_ modules.MessageModule = &Module{}
)

// Module represent x/gov module
type Module struct {
	cdc                        codec.Codec
	db                         *database.Db
	source                     govsource.Source
	authModule                 AuthModule
	distrModule                DistrModule
	slashingModule             SlashingModule
	stakingModule              StakingModule
	proposalNotFoundCount      map[uint64]int
	proposalNotFoundCountMutex sync.Mutex
}

// NewModule returns a new Module instance
func NewModule(
	source govsource.Source,
	authModule AuthModule,
	distrModule DistrModule,
	slashingModule SlashingModule,
	stakingModule StakingModule,
	cdc codec.Codec,
	db *database.Db,
) *Module {
	return &Module{
		cdc:                   cdc,
		source:                source,
		authModule:            authModule,
		distrModule:           distrModule,
		slashingModule:        slashingModule,
		stakingModule:         stakingModule,
		db:                    db,
		proposalNotFoundCount: make(map[uint64]int),
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "gov"
}
