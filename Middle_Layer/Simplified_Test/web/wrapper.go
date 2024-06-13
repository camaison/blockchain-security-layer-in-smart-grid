package web

import (
    "fmt"
    "time"
    "github.com/hyperledger/fabric-gateway/pkg/client"
)

// WrappedContract extends the original Contract with additional functionality
type WrappedContract struct {
    *client.Contract
}

// NewWrappedContract creates a new instance of WrappedContract
func NewWrappedContract(contract *client.Contract) *WrappedContract {
    return &WrappedContract{Contract: contract}
}

// SubmitTransactionWithTiming wraps the SubmitTransaction method to include timing
func (wc *WrappedContract) SubmitTransactionWithTiming(name string, args ...string) ([]byte, error) {
    metrics := make(map[string]float64)

    proposal_start_time := time.Now()
    proposal, err := wc.Contract.NewProposal(name, client.WithArguments(args...))
    if err != nil {
        return nil, err
    }

    prop_end_time := time.Since(proposal_start_time)
    metrics["proposal_endorsement_time"] = prop_end_time.Seconds()

    transaction_start_time := time.Now()
    transaction, err := proposal.Endorse()


    if err != nil {
        return nil, err
    }
    transaction_end_time := time.Since(transaction_start_time)
    metrics["transaction_endorsement_time"] = transaction_end_time.Seconds()

    submission_start_time := time.Now()
    commit, err := transaction.Submit()
    if err != nil {
        return nil, err
    }
    
    submission_end_time := time.Since(submission_start_time)
    metrics["submission_time"] = submission_end_time.Seconds()

    commitment_start_time := time.Now()
    status, err := commit.Status()
    if err != nil {
        return nil, err
    }
    commitment_end_time := time.Since(commitment_start_time)
    metrics["commitment_time"] = commitment_end_time.Seconds()

    if !status.Successful {
        return nil, fmt.Errorf("transaction %s failed to commit with status code %d", status.TransactionID, status.Code)
    }

	end_time := time.Since(proposal_start_time)
    metrics["end_time"] = end_time.Seconds()

	fmt.Printf("%s\n", commitment_end_time)
	
    broadcastMetrics(metrics) // Send metrics to the WebSocket clients

    return transaction.Result(), nil
}
