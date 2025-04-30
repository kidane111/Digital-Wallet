package persistencedb

import (
	"context"
	"digital-wallet/internal/constant/model/dto"

	"github.com/shopspring/decimal"
)

const calculateFee = `-- name: CalculateFee :one
SELECT 
    LEAST(
        GREATEST(
            $2 * (
                base_percentage + 
                (tier_overrides->>$4)::numeric + 
                COALESCE(
                    (SELECT (peak->>'surcharge')::numeric 
                     FROM jsonb_array_elements(peak_hours) AS peak
                     WHERE $3::time BETWEEN 
                           (peak->>'start_hour')::numeric * INTERVAL '1 hour' AND 
                           (peak->>'end_hour')::numeric * INTERVAL '1 hour'
                     LIMIT 1),
                    0
                )
            ) / 100,
            fee_floor
        ),
        fee_cap
    ) AS fee_to_charge
FROM fee_rules
WHERE transaction_type = $1 AND is_active = true
LIMIT 1
`

func (q *PersistenceDB) CalculateFee(ctx context.Context, arg dto.FeeCalculator) (decimal.Decimal, error) {
	var fee decimal.Decimal
	err := q.pool.QueryRow(ctx, calculateFee,
		arg.TransactionType,
		arg.Amount,
		arg.Time.Format("15:04:05"), // Convert time to HH:MM:SS format
		arg.Tire,
	).Scan(&fee)

	return fee, err
}
