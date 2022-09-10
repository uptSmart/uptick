package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/UptickNetwork/uptick/x/nft"

	nfttransfercli "github.com/cosmos/ibc-go/v3/modules/apps/nft-transfer/client/cli"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	nftTxCmd := &cobra.Command{
		Use:                        nft.ModuleName,
		Short:                      "nft transactions subcommands",
		Long:                       "Provides the most common nft logic for upper-level applications, compatible with Ethereum's erc721 contract",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	nftTxCmd.AddCommand(
		NewCmdIssueClass(),
		NewCmdMintNFT(),
		NewCmdSend(),
		nfttransfercli.NewTransferTxCmd(),
	)

	return nftTxCmd
}

func NewCmdSend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [class-id] [nft-id] [receiver] --from [sender]",
		Args:  cobra.ExactArgs(3),
		Short: "transfer ownership of nft",
		Long: strings.TrimSpace(fmt.Sprintf(`
			$ %s tx %s send <class-id> <nft-id> <receiver> --from <sender> --chain-id <chain-id>`, version.AppName, nft.ModuleName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := nft.MsgSend{
				ClassId:  args[0],
				Id:       args[1],
				Sender:   clientCtx.GetFromAddress().String(),
				Receiver: args[2],
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}


func NewCmdIssueClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue [id] --from [sender]",
		Args:  cobra.ExactArgs(1),
		Short: "Issue a nft class",
		Long: strings.TrimSpace(fmt.Sprintf(`
			$ %s tx %s issue <class-id> --name <name> --symbol <symbol> --description <description> --uri <uri> --uri-hash <uri-hash> --from <sender> --chain-id <chain-id>`,
			version.AppName,
			nft.ModuleName,
		),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString(FlagClassName)
			if err != nil {
				return err
			}

			symbol, err := cmd.Flags().GetString(FlagClassSymbol)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(FlagClassDescription)
			if err != nil {
				return err
			}

			uri, err := cmd.Flags().GetString(FlagURI)
			if err != nil {
				return err
			}

			uriHash, err := cmd.Flags().GetString(FlagURIHash)
			if err != nil {
				return err
			}

			msg := nft.MsgIssueClass{
				Id:          args[0],
				Name:        name,
				Symbol:      symbol,
				Description: description,
				Uri:         uri,
				UriHash:     uriHash,
				Issuer:      clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	cmd.Flags().AddFlagSet(fsIssueClass)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewCmdMintNFT() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [class-id] [nft-id] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Mint a nft",
		Long: strings.TrimSpace(fmt.Sprintf(`
			$ %s tx %s mint [class-id] [id] --uri <uri> --uri-hash <uri-hash> --from <sender> --chain-id <chain-id>`,
			version.AppName,
			nft.ModuleName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			receiver, err := cmd.Flags().GetString(FlagReceiver)
			if err != nil {
				return err
			}

			uri, err := cmd.Flags().GetString(FlagURI)
			if err != nil {
				return err
			}

			uriHash, err := cmd.Flags().GetString(FlagURIHash)
			if err != nil {
				return err
			}

			msg := nft.MsgMintNFT{
				ClassId:  args[0],
				Id:       args[1],
				Uri:      uri,
				UriHash:  uriHash,
				Minter:   clientCtx.GetFromAddress().String(),
				Receiver: receiver,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	cmd.Flags().AddFlagSet(fsMintNFT)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
