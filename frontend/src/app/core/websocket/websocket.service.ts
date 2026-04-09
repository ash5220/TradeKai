import { computed, inject, Injectable, OnDestroy, signal } from "@angular/core";
import type { Tick } from "../../shared/models/market.model";
import type { Order } from "../../shared/models/order.model";
import { AuthService } from "../auth/auth.service";
import { environment } from "../../../environments/environment";

export type ConnectionState =
  | "connecting"
  | "connected"
  | "disconnected"
  | "error";

interface WsMessage<T = unknown> {
  type: string;
  room?: string;
  payload: T;
}

@Injectable({ providedIn: "root" })
export class WebSocketService implements OnDestroy {
  private readonly auth = inject(AuthService);

  private ws: WebSocket | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectDelay = 1000;
  private readonly maxDelay = 30_000;
  private destroyed = false;

  private readonly _connectionState = signal<ConnectionState>("disconnected");
  private readonly _ticks = signal<Map<string, Tick>>(new Map());
  private readonly _orderUpdates = signal<Order[]>([]);

  readonly connectionState = this._connectionState.asReadonly();
  readonly ticks = this._ticks.asReadonly();
  readonly latestOrders = this._orderUpdates.asReadonly();
  readonly isConnected = computed(
    () => this._connectionState() === "connected",
  );

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return;
    this.destroyed = false;
    this.openConnection();
  }

  disconnect(): void {
    this.destroyed = true;
    this.clearReconnectTimer();
    this.ws?.close();
  }

  subscribe(room: string): void {
    this.send({ action: "subscribe", room });
  }

  unsubscribe(room: string): void {
    this.send({ action: "unsubscribe", room });
  }

  private openConnection(): void {
    const token = this.auth.accessToken();
    if (!token) return;

    this._connectionState.set("connecting");
    const url = `${environment.wsUrl}?token=${encodeURIComponent(token)}`;
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      this._connectionState.set("connected");
      this.reconnectDelay = 1000;
    };

    this.ws.onmessage = (event: MessageEvent<string>) => {
      try {
        const msg = JSON.parse(event.data) as WsMessage;
        this.handleMessage(msg);
      } catch {
        // ignore malformed messages
      }
    };

    this.ws.onerror = () => {
      this._connectionState.set("error");
    };

    this.ws.onclose = () => {
      this._connectionState.set("disconnected");
      if (!this.destroyed) {
        this.scheduleReconnect();
      }
    };
  }

  private handleMessage(msg: WsMessage): void {
    switch (msg.type) {
      case "tick": {
        const tick = msg.payload as Tick;
        this._ticks.update((map) => {
          const next = new Map(map);
          next.set(tick.symbol, tick);
          return next;
        });
        break;
      }
      case "order_update": {
        const order = msg.payload as Order;
        this._orderUpdates.update((orders) => [order, ...orders.slice(0, 99)]);
        break;
      }
    }
  }

  private send(payload: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(payload));
    }
  }

  private scheduleReconnect(): void {
    this.clearReconnectTimer();
    this.reconnectTimer = setTimeout(() => {
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxDelay);
      this.openConnection();
    }, this.reconnectDelay);
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  ngOnDestroy(): void {
    this.disconnect();
  }
}
