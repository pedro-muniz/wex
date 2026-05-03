package queries

const (
	InsertTransaction = `
		INSERT INTO purchase_transactions (id, description, transaction_date, amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	GetTransactionByID = `
		SELECT id, description, transaction_date, amount, status, created_at, updated_at 
		FROM purchase_transactions 
		WHERE id = $1`

	UpdateTransaction = `
		UPDATE purchase_transactions 
		SET description = $2, transaction_date = $3, amount = $4, status = $5, updated_at = $6 
		WHERE id = $1`

	DeleteTransaction = `DELETE FROM purchase_transactions WHERE id = $1`
)
