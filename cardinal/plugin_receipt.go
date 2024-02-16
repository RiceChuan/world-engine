package cardinal

import (
	"pkg.world.dev/world-engine/cardinal/types/engine"
)

type receiptPlugin struct {
}

func newReceiptPlugin() *receiptPlugin {
	return &receiptPlugin{}
}

var _ Plugin = (*receiptPlugin)(nil)

func (p *receiptPlugin) Register(world *World) error {
	err := p.RegisterQueries(world)
	if err != nil {
		return err
	}
	return nil
}

func (p *receiptPlugin) RegisterQueries(world *World) error {
	err := RegisterQuery[listTxReceiptsRequest, listTxReceiptsResponse](world, "list",
		queryReceipts,
		WithCustomQueryGroup[listTxReceiptsRequest, listTxReceiptsResponse]("receipts"))
	if err != nil {
		return err
	}
	return nil
}

type listTxReceiptsRequest struct {
	StartTick uint64 `json:"startTick" mapstructure:"startTick"`
}

// listTxReceiptsResponse returns the transaction receipts for the given range of ticks. The interval is closed on
// StartTick and open on EndTick: i.e. [StartTick, EndTick)
// Meaning StartTick is included and EndTick is not. To iterate over all ticks in the future, use the returned
// EndTick as the StartTick in the next request. If StartTick == EndTick, the receipts list will be empty.
type listTxReceiptsResponse struct {
	StartTick uint64         `json:"startTick"`
	EndTick   uint64         `json:"endTick"`
	Receipts  []receiptEntry `json:"receipts"`
}

// receiptEntry represents a single transaction receipt. It contains an ID, a result, and a list of errors.
type receiptEntry struct {
	TxHash string  `json:"txHash"`
	Tick   uint64  `json:"tick"`
	Result any     `json:"result"`
	Errors []error `json:"errors"`
}

// queryReceipts godoc
//
//	@Summary		Get transaction receipts from Cardinal
//	@Description	Get transaction receipts from Cardinal
//	@Accept			application/json
//	@Produce		application/json
//	@Param			listTxReceiptsRequest	body		listTxReceiptsRequest	true	"List Transaction Receipts Request"
//	@Success		200						{object}	listTxReceiptsResponse
//	@Failure		400						{string}	string	"Invalid transaction request"
//	@Router			/query/receipts/list [post]
func queryReceipts(ctx engine.Context, req *listTxReceiptsRequest) (*listTxReceiptsResponse, error) {
	reply := listTxReceiptsResponse{}
	reply.EndTick = ctx.CurrentTick()
	size := ctx.ReceiptHistorySize()
	if size > reply.EndTick {
		reply.StartTick = 0
	} else {
		reply.StartTick = reply.EndTick - size
	}
	// StartTick and EndTick are now at the largest possible range of ticks.
	// Check to see if we should narrow down the range at all.
	if req.StartTick > reply.EndTick {
		// User is asking for ticks in the future.
		reply.StartTick = reply.EndTick
	} else if req.StartTick > reply.StartTick {
		reply.StartTick = req.StartTick
	}

	for t := reply.StartTick; t < reply.EndTick; t++ {
		currReceipts, err := ctx.GetTransactionReceiptsForTick(t)
		if err != nil || len(currReceipts) == 0 {
			continue
		}
		for _, r := range currReceipts {
			reply.Receipts = append(reply.Receipts, receiptEntry{
				TxHash: string(r.TxHash),
				Tick:   t,
				Result: r.Result,
				Errors: r.Errs,
			})
		}
	}
	return &reply, nil
}