import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpParams } from "@angular/common/http";
import type { Observable } from "rxjs";
import type { Order } from "../../shared/models/order.model";
import type { Position } from "../../shared/models/position.model";
import type { Candle } from "../../shared/models/market.model";
import { environment } from "../../../environments/environment";

export interface StrategyInfo {
  name: string;
  running: boolean;
}

export interface PnlSummary {
  daily_realized_pnl: number;
}

export interface PlaceOrderRequest {
  symbol: string;
  side: "buy" | "sell";
  type: "market" | "limit" | "stop";
  qty: number;
  limit_price?: number;
}

@Injectable({ providedIn: "root" })
export class ApiService {
  private readonly http = inject(HttpClient);
  private readonly base = environment.apiUrl;

  // ── Orders ────────────────────────────────────────────────────────────────
  getOrders(limit = 50, offset = 0): Observable<Order[]> {
    const params = new HttpParams().set("limit", limit).set("offset", offset);
    return this.http.get<Order[]>(`${this.base}/orders`, { params });
  }

  placeOrder(req: PlaceOrderRequest): Observable<Order> {
    return this.http.post<Order>(`${this.base}/orders`, req);
  }

  cancelOrder(id: string): Observable<void> {
    return this.http.delete<void>(`${this.base}/orders/${id}`);
  }

  // ── Portfolio ─────────────────────────────────────────────────────────────
  getPositions(): Observable<Position[]> {
    return this.http.get<Position[]>(`${this.base}/portfolio/positions`);
  }

  getPnL(): Observable<PnlSummary> {
    return this.http.get<PnlSummary>(`${this.base}/portfolio/pnl`);
  }

  // ── Market ────────────────────────────────────────────────────────────────
  getCandles(
    symbol: string,
    interval = "1m",
    from?: string,
    to?: string,
    limit = 500,
  ): Observable<Candle[]> {
    let params = new HttpParams().set("interval", interval).set("limit", limit);
    if (from) params = params.set("from", from);
    if (to) params = params.set("to", to);
    return this.http.get<Candle[]>(`${this.base}/market/candles/${symbol}`, {
      params,
    });
  }

  // ── Strategies ────────────────────────────────────────────────────────────
  getStrategies(): Observable<StrategyInfo[]> {
    return this.http.get<StrategyInfo[]>(`${this.base}/strategies`);
  }

  startStrategy(name: string, symbols: string[]): Observable<void> {
    return this.http.post<void>(`${this.base}/strategies/${name}/start`, {
      symbols,
    });
  }

  stopStrategy(name: string): Observable<void> {
    return this.http.post<void>(`${this.base}/strategies/${name}/stop`, {});
  }
}
