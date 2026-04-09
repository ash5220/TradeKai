export interface Tick {
  symbol: string;
  price: number;
  volume: number;
  timestamp: string;
}

export interface Candle {
  symbol: string;
  interval: string;
  ts: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface Quote {
  symbol: string;
  bid_price: number;
  ask_price: number;
  bid_size: number;
  ask_size: number;
  timestamp: string;
}
