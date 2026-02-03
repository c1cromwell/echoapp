package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thechadcromwell/echoapp/pkg/cardano"
	"github.com/thechadcromwell/echoapp/pkg/utils"
)

// TransactionHandlers handles transaction tracking endpoints
type TransactionHandlers struct {
	client *cardano.Client
}

// NewTransactionHandlers creates new transaction handlers
func NewTransactionHandlers(client *cardano.Client) *TransactionHandlers {
	return &TransactionHandlers{client: client}
}

// GetTransactionStatus retrieves the status of a transaction
// GET /api/v1/transactions/:txHash
func (h *TransactionHandlers) GetTransactionStatus(c *gin.Context) {
	txHash := c.Param("txHash")

	tx, err := h.client.GetTransactionStatus(c.Request.Context(), txHash)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve transaction", err.Error())
		return
	}

	if tx == nil {
		utils.GinErrorResponse(c, http.StatusNotFound, "Transaction not found", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"tx_hash":       tx.TxHash,
		"status":        tx.Status,
		"operation":     tx.OperationType,
		"entity":        tx.RelatedEntity,
		"created_at":    tx.CreatedAt,
		"confirmed_at":  tx.ConfirmedAt,
		"block_height":  tx.BlockHeight,
		"confirmations": tx.Confirmations,
		"request_id":    c.GetString("request_id"),
	})
}

// QueryTransactions queries transactions with filters
// GET /api/v1/transactions
func (h *TransactionHandlers) QueryTransactions(c *gin.Context) {
	filter := &cardano.TransactionFilter{
		TxHash:        c.Query("tx_hash"),
		Status:        cardano.TransactionStatus(c.Query("status")),
		OperationType: c.Query("operation_type"),
		RelatedEntity: c.Query("related_entity"),
		Limit:         100,
		Offset:        0,
	}

	transactions, err := h.client.QueryTransactions(c.Request.Context(), filter)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to query transactions", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"transaction_count": len(transactions),
		"transactions":      transactions,
		"request_id":        c.GetString("request_id"),
	})
}

// GetPendingTransactions retrieves pending transactions
// GET /api/v1/transactions/pending
func (h *TransactionHandlers) GetPendingTransactions(c *gin.Context) {
	transactions, err := h.client.GetPendingTransactions(c.Request.Context(), 100)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve pending transactions", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"pending_count": len(transactions),
		"transactions":  transactions,
		"request_id":    c.GetString("request_id"),
	})
}

// GetTransactionsByEntity retrieves transactions for an entity
// GET /api/v1/transactions/entity/:entityType/:entityId
func (h *TransactionHandlers) GetTransactionsByEntity(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")

	transactions, err := h.client.GetTransactionsByEntity(c.Request.Context(), entityType, entityID)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve transactions", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"entity_type":       entityType,
		"entity_id":         entityID,
		"transaction_count": len(transactions),
		"transactions":      transactions,
		"request_id":        c.GetString("request_id"),
	})
}

// ConfirmTransaction confirms a pending transaction
// POST /api/v1/transactions/:txHash/confirm
func (h *TransactionHandlers) ConfirmTransaction(c *gin.Context) {
	txHash := c.Param("txHash")

	var req struct {
		BlockHeight int `json:"block_height"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err := h.client.ConfirmTransaction(c.Request.Context(), txHash, req.BlockHeight)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to confirm transaction", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"tx_hash":      txHash,
		"status":       "confirmed",
		"block_height": req.BlockHeight,
		"request_id":   c.GetString("request_id"),
	})
}

// FailTransaction marks a transaction as failed
// POST /api/v1/transactions/:txHash/fail
func (h *TransactionHandlers) FailTransaction(c *gin.Context) {
	txHash := c.Param("txHash")

	var req struct {
		ErrorMessage string `json:"error_message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err := h.client.FailTransaction(c.Request.Context(), txHash, req.ErrorMessage)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to mark transaction as failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"tx_hash":       txHash,
		"status":        "failed",
		"error_message": req.ErrorMessage,
		"request_id":    c.GetString("request_id"),
	})
}

// RetryTransaction retries a failed transaction
// POST /api/v1/transactions/:txHash/retry
func (h *TransactionHandlers) RetryTransaction(c *gin.Context) {
	txHash := c.Param("txHash")

	newTxHash, err := h.client.RetryTransaction(c.Request.Context(), txHash)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retry transaction", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"original_tx": txHash,
		"new_tx_hash": newTxHash,
		"status":      "pending",
		"request_id":  c.GetString("request_id"),
	})
}

// GetTransactionStats retrieves transaction statistics
// GET /api/v1/transactions/stats
func (h *TransactionHandlers) GetTransactionStats(c *gin.Context) {
	stats := h.client.GetTransactionStats(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"stats":      stats,
		"request_id": c.GetString("request_id"),
	})
}

// CancelTransaction cancels a pending transaction
// POST /api/v1/transactions/:txHash/cancel
func (h *TransactionHandlers) CancelTransaction(c *gin.Context) {
	txHash := c.Param("txHash")

	err := h.client.CancelTransaction(c.Request.Context(), txHash)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to cancel transaction", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"tx_hash":    txHash,
		"status":     "cancelled",
		"request_id": c.GetString("request_id"),
	})
}

// GetTransactionMetadata retrieves metadata for a transaction
// GET /api/v1/transactions/:txHash/metadata
func (h *TransactionHandlers) GetTransactionMetadata(c *gin.Context) {
	txHash := c.Param("txHash")

	metadata, err := h.client.GetTransactionMetadata(c.Request.Context(), txHash)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve metadata", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"tx_hash":    txHash,
		"metadata":   metadata,
		"request_id": c.GetString("request_id"),
	})
}
