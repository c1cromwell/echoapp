// Package treasury implements WO-215 treasury multi-sig and PacaSwap seeding state.
package treasury

import (
	"errors"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/tokenomics/genesis"
)

var (
	ErrInsufficientSignatures = errors.New("insufficient founder signatures")
	ErrProposalNotFound       = errors.New("treasury proposal not found")
	ErrProposalExpired        = errors.New("treasury proposal expired")
)

// SubAllocation tracks a treasury bucket balance.
type SubAllocation struct {
	Name      string `json:"name"`
	Allocated int64  `json:"allocated"`
	Remaining int64  `json:"remaining"`
}

// Proposal is a 3-of-5 multi-sig spend request.
type Proposal struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      int64     `json:"amount"`
	Destination string    `json:"destination"`
	Purpose     string    `json:"purpose"`
	Signatures  []string  `json:"signatures"`
	Required    int       `json:"required"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Executed    bool      `json:"executed"`
	TxHash      string    `json:"txHash,omitempty"`
}

// State is the live treasury monitoring view.
type State struct {
	TotalBalance   int64           `json:"totalBalance"`
	SubAllocations []SubAllocation `json:"subAllocations"`
	PacaSwapSeeded bool            `json:"pacaSwapSeeded"`
	PacaSwapSeedTx string          `json:"pacaSwapSeedTx,omitempty"`
	LPBalance      int64           `json:"lpTokenBalance"`
}

// Manager coordinates treasury proposals and balance tracking.
type Manager struct {
	mu          sync.Mutex
	founderDIDs []string
	required    int
	proposals   map[string]*Proposal
	state       State
}

// NewManager creates a treasury manager with 3-of-5 founder multi-sig.
func NewManager(founderDIDs []string) *Manager {
	return &Manager{
		founderDIDs: append([]string(nil), founderDIDs...),
		required:    3,
		proposals:   make(map[string]*Proposal),
		state: State{
			TotalBalance: genesis.TreasuryAmount,
			SubAllocations: []SubAllocation{
				{Name: "paca_swap_liquidity", Allocated: genesis.TreasuryPacaSwapSeed, Remaining: genesis.TreasuryPacaSwapSeed},
				{Name: "operational_reserve", Allocated: genesis.TreasuryOperational, Remaining: genesis.TreasuryOperational},
				{Name: "multisig_reserve", Allocated: genesis.TreasuryMultisigReserve, Remaining: genesis.TreasuryMultisigReserve},
			},
		},
	}
}

// GetState returns a snapshot of treasury balances.
func (m *Manager) GetState() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// SeedPacaSwap records the ECHO/DAG pool seeding at genesis (WO-215).
func (m *Manager) SeedPacaSwap(txHash string, dagAmount int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.state.PacaSwapSeeded {
		return errors.New("paca swap already seeded")
	}
	for i, sub := range m.state.SubAllocations {
		if sub.Name == "paca_swap_liquidity" {
			m.state.SubAllocations[i].Remaining = 0
			break
		}
	}
	m.state.PacaSwapSeeded = true
	m.state.PacaSwapSeedTx = txHash
	m.state.LPBalance = genesis.TreasuryPacaSwapSeed
	_ = dagAmount
	return nil
}

// CreateProposal starts a new multi-sig spend proposal (24h collection window).
func (m *Manager) CreateProposal(id, title, destination, purpose string, amount int64) (*Proposal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	p := &Proposal{
		ID:          id,
		Title:       title,
		Amount:      amount,
		Destination: destination,
		Purpose:     purpose,
		Required:    m.required,
		CreatedAt:   now,
		ExpiresAt:   now.Add(24 * time.Hour),
	}
	m.proposals[id] = p
	return p, nil
}

// SignProposal records a founder signature.
func (m *Manager) SignProposal(proposalID, founderDID string) (*Proposal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.proposals[proposalID]
	if !ok {
		return nil, ErrProposalNotFound
	}
	if time.Now().After(p.ExpiresAt) {
		return nil, ErrProposalExpired
	}
	if !m.isFounder(founderDID) {
		return nil, errors.New("only founder DIDs may sign treasury proposals")
	}
	for _, s := range p.Signatures {
		if s == founderDID {
			return p, nil
		}
	}
	p.Signatures = append(p.Signatures, founderDID)
	return p, nil
}

// ExecuteProposal finalizes when 3-of-5 signatures collected.
func (m *Manager) ExecuteProposal(proposalID, txHash string) (*Proposal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.proposals[proposalID]
	if !ok {
		return nil, ErrProposalNotFound
	}
	if len(p.Signatures) < m.required {
		return nil, ErrInsufficientSignatures
	}
	if p.Executed {
		return p, nil
	}
	p.Executed = true
	p.TxHash = txHash
	m.state.TotalBalance -= p.Amount
	return p, nil
}

func (m *Manager) isFounder(did string) bool {
	for _, f := range m.founderDIDs {
		if f == did {
			return true
		}
	}
	return false
}

// GetProposal returns proposal status.
func (m *Manager) GetProposal(id string) (*Proposal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.proposals[id]
	if !ok {
		return nil, ErrProposalNotFound
	}
	cp := *p
	return &cp, nil
}
