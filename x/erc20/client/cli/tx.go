package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"

	"github.com/UptickNetwork/uptick/x/erc20/types"
)

// NewTxCmd returns a root CLI command handler for certain modules/erc20 transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "erc20 subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewConvertCoinCmd(),
		NewConvertERC20Cmd(),
	)
	return txCmd
}

// NewConvertCoinCmd returns a CLI command handler for converting cosmos coins
func NewConvertCoinCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert-coin [coin] [receiver_hex]",
		Short: "Convert a Cosmos coin to ERC20",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			var receiver string
			sender := cliCtx.GetFromAddress()

			if len(args) == 2 {
				receiver = args[1]
				if err := ethermint.ValidateAddress(receiver); err != nil {
					return fmt.Errorf("invalid receiver hex address %w", err)
				}
			} else {
				receiver = common.BytesToAddress(sender).Hex()
			}

			msg := &types.MsgConvertCoin{
				Coin:     coin,
				Receiver: receiver,
				Sender:   sender.String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewConvertERC20Cmd returns a CLI command handler for converting ERC20s
func NewConvertERC20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert-erc20 [contract-address] [denom] [amount] [receiver]",
		Short: "Convert an ERC20 token to Cosmos coin",
		Args:  cobra.RangeArgs(3, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contract := args[0]
			if err := ethermint.ValidateAddress(contract); err != nil {
				return fmt.Errorf("invalid ERC20 contract address %w", err)
			}
			denom := args[1]
			amount, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid amount %s", args[1])
			}

			from := common.BytesToAddress(cliCtx.GetFromAddress().Bytes())

			receiver := cliCtx.GetFromAddress()
			num := len(args)
			if num == 4 {
				receiver, err = sdk.AccAddressFromBech32(args[3])
				if err != nil {
					return err
				}
			}

			msg := &types.MsgConvertERC20{
				ContractAddress: contract,
				Amount:          amount,
				Receiver:        receiver.String(),
				Sender:          from.Hex(),
				Denom:           denom,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewRegisterCoinProposalCmd implements the command to submit a community-pool-spend proposal
func NewRegisterCoinProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-coin [metadata]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a register coin proposal",
		Long: `Submit a proposal to register a Cosmos coin to the erc20 along with an initial deposit.
Upon passing, the
The proposal details must be supplied via a JSON file.`,
		Example: fmt.Sprintf(`$ %s tx gov submit-proposal register-coin <path/to/metadata.json> --from=<key_or_address>

Where metadata.json contains (example):

{
  "description": "staking, gas and governance token of the Uptick testnets"
  "denom_units": [
		{
			"denom": "auptick",
			"exponent": 0,
			"aliases": ["atto uptick"]
		},
		{
			"denom": "uptick",
			"exponent": 18
		}
	],
	"base": "auptick",
	"display: "uptick",
	"name": "Uptick",
	"symbol": "UPTICK"
}`, version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			metadata, err := ParseMetadata(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			content := types.NewRegisterCoinProposal(title, description, metadata)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "1auptick", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}

// NewAddCoinProposalCmd implements the command to submit a community-pool-spend proposal
func NewAddCoinProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-coin [ERC20Address] [metadata]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit a register coin proposal",
		Long: `Submit a proposal to register a Cosmos coin to the intrarelayer along with an initial deposit.
Upon passing, the
The proposal details must be supplied via a JSON file.`,
		Example: fmt.Sprintf(`$ %s tx gov submit-proposal add-coin <ERC20Address> <path/to/metadata.json> --from=<key_or_address>

Where metadata.json contains (example):

{
  "description": "staking, gas and governance token of the Uptick testnets",
  "denom_units": [
		{
			"denom": "atele",
			"exponent": 0,
			"aliases": ["atto tele"]
		},
		{
			"denom": "tele",
			"exponent": 18
		}
	],
	"base": "atele",
	"display": "tele",
	"name": "Tele",
	"symbol": "TELE"
}`, version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			contract := args[0]

			metadata, err := ParseMetadata(clientCtx.Codec, args[1])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			content := types.NewAddCoinProposal(title, description, metadata, contract)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "1tele", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}

// NewRegisterERC20ProposalCmd implements the command to submit a community-pool-spend proposal
func NewRegisterERC20ProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-erc20 [erc20-address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Submit a proposal to register an ERC20 token",
		Long:    "Submit a proposal to register an ERC20 token to the erc20 along with an initial deposit.",
		Example: fmt.Sprintf("$ %s tx gov submit-proposal register-erc20 <path/to/proposal.json> --from=<key_or_address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			erc20Addr := args[0]
			from := clientCtx.GetFromAddress()
			content := types.NewRegisterERC20Proposal(title, description, erc20Addr)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "1auptick", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}

// NewToggleTokenRelayProposalCmd implements the command to submit a community-pool-spend proposal
func NewToggleTokenRelayProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "toggle-token-relay [token]",
		Args:    cobra.ExactArgs(1),
		Short:   "Submit a toggle token relay proposal",
		Long:    "Submit a proposal to toggle the relaying of a token pair along with an initial deposit.",
		Example: fmt.Sprintf("$ %s tx gov submit-proposal toggle-token-relay <denom_or_contract> --from=<key_or_address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			token := args[0]
			content := types.NewToggleTokenRelayProposal(title, description, token)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "1auptick", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}

// NewUpdateTokenPairERC20ProposalCmd implements the command to submit a community-pool-spend proposal
func NewUpdateTokenPairERC20ProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-token-pair-erc20 [erc20_address] [new_erc20_address]",
		Args:    cobra.ExactArgs(2),
		Short:   "Submit a update token pair ERC20 proposal",
		Long:    `Submit a proposal to update the ERC20 address of a token pair along with an initial deposit.`,
		Example: fmt.Sprintf("$ %s tx gov submit-proposal update-token-pair-erc20 <path/to/proposal.json> --from=<key_or_address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			erc20Addr := args[0]
			newERC20Addr := args[1]

			from := clientCtx.GetFromAddress()
			content := types.NewUpdateTokenPairERC20Proposal(title, description, erc20Addr, newERC20Addr)
			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "1auptick", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}
