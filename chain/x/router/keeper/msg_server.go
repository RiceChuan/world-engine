package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/argus-labs/world-engine/chain/x/router/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.MsgServiceServer = &Keeper{}

func (k *Keeper) UpdateNamespace(ctx context.Context, request *types.UpdateNamespaceRequest) (
	*types.UpdateNamespaceResponse, error,
) {
	if k.authority != request.Authority {
		return nil, sdkerrors.ErrUnauthorized.
			Wrapf("%s is not allowed to update namespaces, expected %s", request.Authority, k.authority)
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	k.setNamespace(sdkCtx, request.Namespace)

	return &types.UpdateNamespaceResponse{}, nil
}