package queries

const (
	GetTransactionByID = `
		SELECT id, description, transaction_date, amount, status, created_at, updated_at 
		FROM purchase_transactions 
		WHERE id = $1`
)
