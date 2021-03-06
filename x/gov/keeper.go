package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// Governance Keeper
type Keeper struct {
	// The reference to the CoinKeeper to modify balances
	ck bank.Keeper

	// The ValidatorSet to get information about validators
	vs sdk.ValidatorSet

	// The reference to the DelegationSet to get information about delegators
	ds sdk.DelegationSet

	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey

	// The wire codec for binary encoding/decoding.
	cdc *wire.Codec

	// Reserved codespace
	codespace sdk.CodespaceType
}

// NewGovernanceMapper returns a mapper that uses go-wire to (binary) encode and decode gov types.
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper, ds sdk.DelegationSet, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:  key,
		ck:        ck,
		ds:        ds,
		vs:        ds.GetValidatorSet(),
		cdc:       cdc,
		codespace: codespace,
	}
}

// Returns the go-wire codec.
func (keeper Keeper) WireCodec() *wire.Codec {
	return keeper.cdc
}

// =====================================================
// Proposals

// Creates a NewProposal
func (keeper Keeper) NewTextProposal(ctx sdk.Context, title string, description string, proposalType byte) Proposal {
	proposalID, err := keeper.getNewProposalID(ctx)
	if err != nil {
		return nil
	}
	var proposal Proposal = &TextProposal{
		ProposalID:       proposalID,
		Title:            title,
		Description:      description,
		ProposalType:     proposalType,
		Status:           StatusDepositPeriod,
		TotalDeposit:     sdk.Coins{},
		SubmitBlock:      ctx.BlockHeight(),
		VotingStartBlock: -1, // TODO: Make Time
	}
	keeper.SetProposal(ctx, proposal)
	keeper.InactiveProposalQueuePush(ctx, proposal)
	return proposal
}

// Get Proposal from store by ProposalID
func (keeper Keeper) GetProposal(ctx sdk.Context, proposalID int64) Proposal {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyProposal(proposalID))
	if bz == nil {
		return nil
	}

	var proposal Proposal
	keeper.cdc.MustUnmarshalBinary(bz, &proposal)

	return proposal
}

// Implements sdk.AccountMapper.
func (keeper Keeper) SetProposal(ctx sdk.Context, proposal Proposal) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(proposal)
	store.Set(KeyProposal(proposal.GetProposalID()), bz)
}

// Implements sdk.AccountMapper.
func (keeper Keeper) DeleteProposal(ctx sdk.Context, proposal Proposal) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyProposal(proposal.GetProposalID()))
}

func (keeper Keeper) setInitialProposalID(ctx sdk.Context, proposalID int64) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextProposalID)
	if bz != nil {
		return ErrInvalidGenesis(keeper.codespace, "Initial ProposalID already set")
	}
	bz = keeper.cdc.MustMarshalBinary(proposalID)
	store.Set(KeyNextProposalID, bz)
	return nil
}

func (keeper Keeper) getNewProposalID(ctx sdk.Context) (proposalID int64, err sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextProposalID)
	if bz == nil {
		return -1, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set")
	}
	keeper.cdc.MustUnmarshalBinary(bz, &proposalID)
	bz = keeper.cdc.MustMarshalBinary(proposalID + 1)
	store.Set(KeyNextProposalID, bz)
	return proposalID, nil
}

func (keeper Keeper) activateVotingPeriod(ctx sdk.Context, proposal Proposal) {
	proposal.SetVotingStartBlock(ctx.BlockHeight())
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)
	keeper.ActiveProposalQueuePush(ctx, proposal)
}

// =====================================================
// Procedures

// Gets procedure from store. TODO: move to global param store and allow for updating of this
func (keeper Keeper) GetDepositProcedure() DepositProcedure {
	return DepositProcedure{
		MinDeposit:       sdk.Coins{sdk.NewCoin("steak", 10)},
		MaxDepositPeriod: 200,
	}
}

// Gets procedure from store. TODO: move to global param store and allow for updating of this
func (keeper Keeper) GetVotingProcedure() VotingProcedure {
	return VotingProcedure{
		VotingPeriod: 200,
	}
}

// Gets procedure from store. TODO: move to global param store and allow for updating of this
func (keeper Keeper) GetTallyingProcedure() TallyingProcedure {
	return TallyingProcedure{
		Threshold:         sdk.NewRat(1, 2),
		Veto:              sdk.NewRat(1, 3),
		GovernancePenalty: sdk.NewRat(1, 100),
	}
}

// =====================================================
// Votes

// Adds a vote on a specific proposal
func (keeper Keeper) AddVote(ctx sdk.Context, proposalID int64, voterAddr sdk.Address, option VoteOption) sdk.Error {
	proposal := keeper.GetProposal(ctx, proposalID)
	if proposal == nil {
		return ErrUnknownProposal(keeper.codespace, proposalID)
	}
	if proposal.GetStatus() != StatusVotingPeriod {
		return ErrInactiveProposal(keeper.codespace, proposalID)
	}

	if option != OptionYes && option != OptionAbstain && option != OptionNo && option != OptionNoWithVeto {
		return ErrInvalidVote(keeper.codespace, VoteOptionToString(option))
	}

	vote := Vote{
		ProposalID: proposalID,
		Voter:      voterAddr,
		Option:     option,
	}
	keeper.setVote(ctx, proposalID, voterAddr, vote)

	return nil
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) GetVote(ctx sdk.Context, proposalID int64, voterAddr sdk.Address) (Vote, bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyVote(proposalID, voterAddr))
	if bz == nil {
		return Vote{}, false
	}
	var vote Vote
	keeper.cdc.MustUnmarshalBinary(bz, &vote)
	return vote, true
}

func (keeper Keeper) setVote(ctx sdk.Context, proposalID int64, voterAddr sdk.Address, vote Vote) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(vote)
	store.Set(KeyVote(proposalID, voterAddr), bz)
}

// Gets all the votes on a specific proposal
func (keeper Keeper) GetVotes(ctx sdk.Context, proposalID int64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, KeyVotesSubspace(proposalID))
}

func (keeper Keeper) deleteVote(ctx sdk.Context, proposalID int64, voterAddr sdk.Address) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyVote(proposalID, voterAddr))
}

// =====================================================
// Deposits

// Gets the deposit of a specific depositer on a specific proposal
func (keeper Keeper) GetDeposit(ctx sdk.Context, proposalID int64, depositerAddr sdk.Address) (Deposit, bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyDeposit(proposalID, depositerAddr))
	if bz == nil {
		return Deposit{}, false
	}
	var deposit Deposit
	keeper.cdc.MustUnmarshalBinary(bz, &deposit)
	return deposit, true
}

func (keeper Keeper) setDeposit(ctx sdk.Context, proposalID int64, depositerAddr sdk.Address, deposit Deposit) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(deposit)
	store.Set(KeyDeposit(proposalID, depositerAddr), bz)
}

// Adds or updates a deposit of a specific depositer on a specific proposal
// Activates voting period when appropriate
func (keeper Keeper) AddDeposit(ctx sdk.Context, proposalID int64, depositerAddr sdk.Address, depositAmount sdk.Coins) (sdk.Error, bool) {
	// Checks to see if proposal exists
	proposal := keeper.GetProposal(ctx, proposalID)
	if proposal == nil {
		return ErrUnknownProposal(keeper.codespace, proposalID), false
	}

	// Check if proposal is still depositable
	if (proposal.GetStatus() != StatusDepositPeriod) && (proposal.GetStatus() != StatusVotingPeriod) {
		return ErrAlreadyFinishedProposal(keeper.codespace, proposalID), false
	}

	// Subtract coins from depositer's account
	_, _, err := keeper.ck.SubtractCoins(ctx, depositerAddr, depositAmount)
	if err != nil {
		return err, false
	}

	// Update Proposal
	proposal.SetTotalDeposit(proposal.GetTotalDeposit().Plus(depositAmount))
	keeper.SetProposal(ctx, proposal)

	// Check if deposit tipped proposal into voting period
	// Active voting period if so
	activatedVotingPeriod := false
	if proposal.GetStatus() == StatusDepositPeriod && proposal.GetTotalDeposit().IsGTE(keeper.GetDepositProcedure().MinDeposit) {
		keeper.activateVotingPeriod(ctx, proposal)
		activatedVotingPeriod = true
	}

	// Add or update deposit object
	currDeposit, found := keeper.GetDeposit(ctx, proposalID, depositerAddr)
	if !found {
		newDeposit := Deposit{depositerAddr, proposalID, depositAmount}
		keeper.setDeposit(ctx, proposalID, depositerAddr, newDeposit)
	} else {
		currDeposit.Amount = currDeposit.Amount.Plus(depositAmount)
		keeper.setDeposit(ctx, proposalID, depositerAddr, currDeposit)
	}

	return nil, activatedVotingPeriod
}

// Gets all the deposits on a specific proposal
func (keeper Keeper) GetDeposits(ctx sdk.Context, proposalID int64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, KeyDepositsSubspace(proposalID))
}

// Returns and deletes all the deposits on a specific proposal
func (keeper Keeper) RefundDeposits(ctx sdk.Context, proposalID int64) {
	store := ctx.KVStore(keeper.storeKey)
	depositsIterator := keeper.GetDeposits(ctx, proposalID)

	for ; depositsIterator.Valid(); depositsIterator.Next() {
		deposit := &Deposit{}
		keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), deposit)

		_, _, err := keeper.ck.AddCoins(ctx, deposit.Depositer, deposit.Amount)
		if err != nil {
			panic("should not happen")
		}

		store.Delete(depositsIterator.Key())
	}

	depositsIterator.Close()
}

// Deletes all the deposits on a specific proposal without refunding them
func (keeper Keeper) DeleteDeposits(ctx sdk.Context, proposalID int64) {
	store := ctx.KVStore(keeper.storeKey)
	depositsIterator := keeper.GetDeposits(ctx, proposalID)

	for ; depositsIterator.Valid(); depositsIterator.Next() {
		store.Delete(depositsIterator.Key())
	}

	depositsIterator.Close()
}

// =====================================================
// ProposalQueues

func (keeper Keeper) getActiveProposalQueue(ctx sdk.Context) ProposalQueue {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyActiveProposalQueue)
	if bz == nil {
		return nil
	}

	var proposalQueue ProposalQueue
	keeper.cdc.MustUnmarshalBinary(bz, &proposalQueue)

	return proposalQueue
}

func (keeper Keeper) setActiveProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(proposalQueue)
	store.Set(KeyActiveProposalQueue, bz)
}

// Return the Proposal at the front of the ProposalQueue
func (keeper Keeper) ActiveProposalQueuePeek(ctx sdk.Context) Proposal {
	proposalQueue := keeper.getActiveProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	return keeper.GetProposal(ctx, proposalQueue[0])
}

// Remove and return a Proposal from the front of the ProposalQueue
func (keeper Keeper) ActiveProposalQueuePop(ctx sdk.Context) Proposal {
	proposalQueue := keeper.getActiveProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	frontElement, proposalQueue := proposalQueue[0], proposalQueue[1:]
	keeper.setActiveProposalQueue(ctx, proposalQueue)
	return keeper.GetProposal(ctx, frontElement)
}

// Add a proposalID to the back of the ProposalQueue
func (keeper Keeper) ActiveProposalQueuePush(ctx sdk.Context, proposal Proposal) {
	proposalQueue := append(keeper.getActiveProposalQueue(ctx), proposal.GetProposalID())
	keeper.setActiveProposalQueue(ctx, proposalQueue)
}

func (keeper Keeper) getInactiveProposalQueue(ctx sdk.Context) ProposalQueue {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyInactiveProposalQueue)
	if bz == nil {
		return nil
	}

	var proposalQueue ProposalQueue

	keeper.cdc.MustUnmarshalBinary(bz, &proposalQueue)

	return proposalQueue
}

func (keeper Keeper) setInactiveProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(proposalQueue)
	store.Set(KeyInactiveProposalQueue, bz)
}

// Return the Proposal at the front of the ProposalQueue
func (keeper Keeper) InactiveProposalQueuePeek(ctx sdk.Context) Proposal {
	proposalQueue := keeper.getInactiveProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	return keeper.GetProposal(ctx, proposalQueue[0])
}

// Remove and return a Proposal from the front of the ProposalQueue
func (keeper Keeper) InactiveProposalQueuePop(ctx sdk.Context) Proposal {
	proposalQueue := keeper.getInactiveProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	frontElement, proposalQueue := proposalQueue[0], proposalQueue[1:]
	keeper.setInactiveProposalQueue(ctx, proposalQueue)
	return keeper.GetProposal(ctx, frontElement)
}

// Add a proposalID to the back of the ProposalQueue
func (keeper Keeper) InactiveProposalQueuePush(ctx sdk.Context, proposal Proposal) {
	proposalQueue := append(keeper.getInactiveProposalQueue(ctx), proposal.GetProposalID())
	keeper.setInactiveProposalQueue(ctx, proposalQueue)
}
