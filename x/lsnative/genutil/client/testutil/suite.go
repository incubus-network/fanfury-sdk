package testutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/incubus-network/fanfury-sdk/v2/furyapp"
	"github.com/incubus-network/fanfury-sdk/v2/testutil/network"
	"github.com/incubus-network/fanfury-sdk/v2/x/lsnative/genutil/client/cli"
	stakingcli "github.com/incubus-network/fanfury-sdk/v2/x/lsnative/staking/client/cli"
	"github.com/incubus-network/fanfury-sdk/v2/x/lsnative/staking/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGenTxCmd() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	amount := sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(12))

	tests := []struct {
		name     string
		args     []string
		expError bool
	}{
		{
			name: "invalid commission rate returns error",
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, s.network.Config.ChainID),
				fmt.Sprintf("--%s=1", stakingcli.FlagCommissionRate),
				val.Moniker,
				amount.String(),
			},
			expError: true,
		},
		{
			name: "valid gentx",
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, s.network.Config.ChainID),
				val.Moniker,
				amount.String(),
			},
			expError: false,
		},
		{
			name: "invalid pubkey",
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, s.network.Config.ChainID),
				fmt.Sprintf("--%s={\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"}", stakingcli.FlagPubKey),
				val.Moniker,
				amount.String(),
			},
			expError: true,
		},
		{
			name: "valid pubkey flag",
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, s.network.Config.ChainID),
				fmt.Sprintf("--%s={\"@type\":\"/cosmos.crypto.ed25519.PubKey\",\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"}", stakingcli.FlagPubKey),
				val.Moniker,
				amount.String(),
			},
			expError: false,
		},
	}

	for _, tc := range tests {
		tc := tc

		dir := s.T().TempDir()
		genTxFile := filepath.Join(dir, "myTx")
		tc.args = append(tc.args, fmt.Sprintf("--%s=%s", flags.FlagOutputDocument, genTxFile))

		s.Run(tc.name, func() {
			cmd := cli.GenTxCmd(
				furyapp.ModuleBasics,
				val.ClientCtx.TxConfig,
				banktypes.GenesisBalancesIterator{},
				val.ClientCtx.HomeDir)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expError {
				s.Require().Error(err)

				_, err = os.Open(genTxFile)
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, "test: %s\noutput: %s", tc.name, out.String())

				// validate generated transaction.
				open, err := os.Open(genTxFile)
				s.Require().NoError(err)

				all, err := io.ReadAll(open)
				s.Require().NoError(err)

				tx, err := val.ClientCtx.TxConfig.TxJSONDecoder()(all)
				s.Require().NoError(err)

				msgs := tx.GetMsgs()
				s.Require().Len(msgs, 1)

				s.Require().Equal(sdk.MsgTypeURL(&types.MsgCreateValidator{}), sdk.MsgTypeURL(msgs[0]))
				s.Require().True(val.Address.Equals(msgs[0].GetSigners()[0]))
				s.Require().Equal(amount, msgs[0].(*types.MsgCreateValidator).Value)
				s.Require().NoError(tx.ValidateBasic())
			}
		})
	}
}
