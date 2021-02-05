package modulecreate

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"

	"github.com/tendermint/starport/starport/templates/module"

	"github.com/gobuffalo/genny"
	_ "github.com/gobuffalo/plush"
	_ "github.com/gobuffalo/plushgen"
	_ "github.com/tendermint/starport/starport/pkg/cosmosver"
)

// New ...
func NewIBC(opts *CreateOptions) (*genny.Generator, error) {
	g := genny.New()

	g.RunFn(appModifyModule(opts))

	//if err := g.Box(templates[cosmosver.Stargate]); err != nil {
	//	return g, err
	//}
	//ctx := plush.NewContext()
	//ctx.Set("moduleName", opts.ModuleName)
	//ctx.Set("modulePath", opts.ModulePath)
	//ctx.Set("appName", opts.AppName)
	//ctx.Set("ownerName", opts.OwnerName)
	//ctx.Set("title", strings.Title)
	//
	//ctx.Set("nodash", func(s string) string {
	//	return strings.ReplaceAll(s, "-", "")
	//})
	//
	//g.Transformer(plushgen.Transformer(ctx))
	//g.Transformer(genny.Replace("{{moduleName}}", opts.ModuleName))
	return g, nil
}

// app.go modification on Stargate when creating a module
func appModifyModule(opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := module.PathAppGo
		_, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Import
		// PlaceholderIBCModuleImport
		_ = `
capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
porttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/05-port/types"
host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
`

		// Interface to implement
		// PlaceholderIBCModuleInterface
		_  = `
_ porttypes.IBCModule   = AppModule{}`

		// IBC interface implementation
		// PlaceholderIBCModuleMethods
		_ = templateIBCModuleMetohds

		// newFile := genny.NewFileS(path, content)
		return nil // return r.File(newFile)
	}
}

const templateIBCModuleMetohds = `
// OnChanOpenInit implements the IBCModule interface
func (am AppModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	if order != channeltypes.NONE {
		return sdkerrors.Wrapf(channeltypes.ErrInvalidChannelOrdering, "expected %s channel, got %s ", channeltypes.NONE, order)
	}

	// Require portID is the portID module is bound to
	boundPort := am.keeper.GetPort(ctx)
	if boundPort != portID {
		return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	}

	if version != types.Version {
		return sdkerrors.Wrapf(types.ErrInvalidVersion, "got %s, expected %s", version, types.Version)
	}

	// Claim channel capability passed back by IBC module
	if err := am.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return err
	}

	return nil
}

// OnChanOpenTry implements the IBCModule interface
func (am AppModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version,
	counterpartyVersion string,
) error {
	if order != channeltypes.NONE {
		return sdkerrors.Wrapf(channeltypes.ErrInvalidChannelOrdering, "expected %s channel, got %s ", channeltypes.NONE, order)
	}

	// Require portID is the portID module is bound to
	boundPort := am.keeper.GetPort(ctx)
	if boundPort != portID {
		return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	}

	if version != types.Version {
		return sdkerrors.Wrapf(types.ErrInvalidVersion, "got: %s, expected %s", version, types.Version)
	}

	if counterpartyVersion != types.Version {
		return sdkerrors.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: got: %s, expected %s", counterpartyVersion, types.Version)
	}

	// Module may have already claimed capability in OnChanOpenInit in the case of crossing hellos
	// (ie chainA and chainB both call ChanOpenInit before one of them calls ChanOpenTry)
	// If module can already authenticate the capability then module already owns it so we don't need to claim
	// Otherwise, module does not have channel capability and we must claim it from IBC
	if !am.keeper.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
		// Only claim channel capability passed back by IBC module if we do not already own it
		if err := am.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
			return err
		}
	}

	return nil
}

// OnChanOpenAck implements the IBCModule interface
func (am AppModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyVersion string,
) error {
	if counterpartyVersion != types.Version {
		return sdkerrors.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: %s, expected %s", counterpartyVersion, types.Version)
	}
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (am AppModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (am AppModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for channels
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (am AppModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface
func (am AppModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
) (*sdk.Result, []byte, error) {
	var data types.FooPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	}

	acknowledgement := channeltypes.NewResultAcknowledgement([]byte{byte(1)})

	err := am.keeper.OnRecvPacket(ctx, packet, data)
	if err != nil {
		acknowledgement = channeltypes.NewErrorAcknowledgement(err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePacket,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyReceiver, data.Receiver),
			sdk.NewAttribute(types.AttributeKeyAckSuccess, fmt.Sprintf("%t", err != nil)),
		),
	)

	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, acknowledgement.GetBytes(), nil
}

// OnAcknowledgementPacket implements the IBCModule interface
func (am AppModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
) (*sdk.Result, error) {
	var ack channeltypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet acknowledgement: %v", err)
	}
	var data types.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	}

	if err := am.keeper.OnAcknowledgementPacket(ctx, packet, data, ack); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePacket,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyReceiver, data.Receiver),
			sdk.NewAttribute(types.AttributeKeyAck, fmt.Sprintf("%v", ack)),
		),
	)

	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckError, resp.Error),
			),
		)
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// OnTimeoutPacket implements the IBCModule interface
func (am AppModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
) (*sdk.Result, error) {
	var data types.FooPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	}
	// refund tokens
	if err := am.keeper.OnTimeoutPacket(ctx, packet, data); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTimeout,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyRefundAmount, fmt.Sprintf("%d", data.Amount)),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}`