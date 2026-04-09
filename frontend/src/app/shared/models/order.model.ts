export interface Order {
  id: string;
  user_id: string;
  symbol: string;
  side: 'buy' | 'sell';
  type: 'market' | 'limit' | 'stop';
  qty: number;
  limit_price: number | null;
  filled_qty: number;
  filled_avg_price: number | null;
  status: 'pending' | 'submitted' | 'partially_filled' | 'filled' | 'cancelled' | 'rejected';
  exchange_id: string | null;
  idempotency_key: string;
  created_at: string;
  updated_at: string;
  filled_at: string | null;
}
