package postgres

import (
	"L0-wb/internal/domain"
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// InsertOrder - Вставка заказа
func (r *Repository) InsertOrder(ctx context.Context, order *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. orders
	ordersQuery := `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`
	_, err = tx.Exec(ctx, ordersQuery,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return err
	}

	// 2. delivery
	deliveryQuery := `
		INSERT INTO delivery (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`
	_, err = tx.Exec(ctx, deliveryQuery,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return err
	}

	// 3. payment
	paymentQuery := `
		INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider, amount,
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`
	_, err = tx.Exec(ctx, paymentQuery,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return err
	}

	// 4. items
	itemsQuery := `
		INSERT INTO items (
			order_uid, chrt_id, track_number, price, rid, name, sale,
			size, total_price, nm_id, brand, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, itemsQuery,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale,
			item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetOrder - Получение заказа
func (r *Repository) GetOrder(ctx context.Context, orderId string) (domain.Order, error) {
	query := `
		SELECT json_build_object(
			'order_uid', o.order_uid,
			'track_number', o.track_number,
			'entry', o.entry,
			'locale', o.locale,
			'internal_signature', o.internal_signature,
			'customer_id', o.customer_id,
			'delivery_service', o.delivery_service,
			'shard_key', o.shardkey,
			'sm_id', o.sm_id,
			'date_created', to_char(o.date_created, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'),
			'oof_shard', o.oof_shard,
			'delivery', json_build_object(
				'name', d.name,
				'phone', d.phone,
				'zip', d.zip,
				'city', d.city,
				'address', d.address,
				'region', d.region,
				'email', d.email
			),
			'payment', json_build_object(
				'transaction', p.transaction,
				'request_id', p.request_id,
				'currency', p.currency,
				'provider', p.provider,
				'amount', p.amount,
				'payment_dt', p.payment_dt,
				'bank', p.bank,
				'delivery_cost', p.delivery_cost,
				'goods_total', p.goods_total,
				'custom_fee', p.custom_fee
			),
			'items', (
				SELECT json_agg(
					json_build_object(
						'chrt_id', i.chrt_id,
						'track_number', i.track_number,
						'price', i.price,
						'rid', i.rid,
						'name', i.name,
						'sale', i.sale,
						'size', i.size,
						'total_price', i.total_price,
						'nm_id', i.nm_id,
						'brand', i.brand,
						'status', i.status
					)
				)
				FROM items i
				WHERE i.order_uid = o.order_uid
			)
		)
		FROM orders o
		JOIN delivery d ON d.order_uid = o.order_uid
		JOIN payment p ON p.order_uid = o.order_uid
		WHERE o.order_uid = $1;
	`

	var row []byte
	err := r.pool.QueryRow(ctx, query, orderId).Scan(&row)
	if err != nil {
		return domain.Order{}, err
	}

	var order domain.Order
	if err := json.Unmarshal(row, &order); err != nil {
		return domain.Order{}, err
	}

	return order, nil
}
