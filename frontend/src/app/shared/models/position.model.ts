export interface Position {
  user_id: string;
  symbol: string;
  qty: number;
  avg_price: number;
  realized_pnl: number;
  updated_at: string;
  // Computed client-side from live prices:
  current_price?: number;
  unrealized_pnl?: number;
}
