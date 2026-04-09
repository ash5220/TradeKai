import { Component, inject, OnInit, signal } from "@angular/core";
import { DecimalPipe, NgClass } from "@angular/common";
import { ApiService } from "../../../core/api/api.service";
import { WebSocketService } from "../../../core/websocket/websocket.service";
import { Position } from "../../../shared/models/position.model";

@Component({
  selector: "tk-position-table",
  standalone: true,
  imports: [DecimalPipe, NgClass],
  templateUrl: "./position-table.component.html",
  styleUrl: "./position-table.component.scss",
})
export class PositionTableComponent implements OnInit {
  private readonly api = inject(ApiService);
  private readonly ws = inject(WebSocketService);

  readonly positions = signal<Position[]>([]);

  ngOnInit(): void {
    this.refresh();
  }

  currentPrice(symbol: string): number {
    return this.ws.ticks().get(symbol)?.price ?? 0;
  }

  refresh(): void {
    this.api.getPositions().subscribe({
      next: (positions) => this.positions.set(positions),
    });
  }

  closePosition(pos: Position): void {
    this.api
      .placeOrder({
        symbol: pos.symbol,
        side: "sell",
        type: "market",
        qty: pos.qty,
      })
      .subscribe({ next: () => this.refresh() });
  }
}
