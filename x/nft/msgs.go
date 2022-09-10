package nft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"strings"
)

var (
	_ sdk.Msg = &MsgSend{}
	_ sdk.Msg = &MsgIssueClass{}
	_ sdk.Msg = &MsgMintNFT{}
)

// GetSigners implements the Msg.ValidateBasic method.
func (m MsgSend) ValidateBasic() error {
	if err := ValidateClassID(m.ClassId); err != nil {
		return sdkerrors.Wrapf(ErrInvalidID, "Invalid class id (%s)", m.ClassId)
	}

	if err := ValidateNFTID(m.Id); err != nil {
		return sdkerrors.Wrapf(ErrInvalidID, "Invalid nft id (%s)", m.Id)
	}

	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", m.Sender)
	}

	if _, err := sdk.AccAddressFromBech32(m.Receiver); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid receiver address (%s)", m.Receiver)
	}

	return nil
}

// GetSigners implements Msg
func (m MsgSend) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{signer}
}



// ValidateBasic implements sdk.Msg
func (msg MsgIssueClass) ValidateBasic() error {
	if strings.TrimSpace(msg.Issuer) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Issuer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.Issuer)
	}

	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgIssueClass) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{accAddr}
}

// GetSigners implements sdk.Msg
func (msg MsgMintNFT) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Minter)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{accAddr}
}

// ValidateBasic implements sdk.Msg
func (msg MsgMintNFT) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Minter)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid minter address")
	}

	return nil
}
