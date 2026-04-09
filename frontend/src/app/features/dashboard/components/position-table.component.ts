import { Component, inject, OnInit, signal } from '@angular/core';
import { DecimalPipe, NgClass } from '@angular/common';
import { ApiService } from '../../../core/api/api.service';
import { WebSocketService } from '../../../core/websocket/websocket.service';
import { Position } from '../../../shared/models/position.model';

@Component({
  selector: 'tk-position-table',
  standalone: true,
  imports: [DecimalPipe, NgClass],
  template: `
    <div class="card">
      <div class="card-header">
        <h3>Open Positions</h3>
        <button class="btn-ghost" (click)="refresh()">Refresh</button>
      </div>
      @if (positions().length === 0) {
        <p class="empty">No open positions.</p>
      } @else {
        <table class="data-table">
          <thead>
            <tr>
              <th>Symbol</th>
              <th>Qty</th>
              <th>Avg Price</th>
              <th>Current</th>
              <th>Unrealized P&amp;L</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            @for (pos of positions(); track pos.symbol) {
              <tr>
                <td class="symbol">{{ pos.symbol }}</td>
                <td>{{ pos.qty | number:'1.0-4' }}</td>
                <td>{{ pos.avg_price | number:'1.2-2' }}</td>
                <td>{{ currentPrice(pos.symbol) | number:'1.2-2' }}</td>
                <td [ngClass]="(pos.unrealized_pnl ?? 0) >= 0 ? 'pnl-positive' : 'pnl-negative'">
                  {{ pos.unrealized_pnl | number:'1.2-2' }}
                </td>
                <td>
                  <button class="btn-danger-sm" (click)="closePosition(pos)">Close</button>
                </td>
              </tr>
            }
          </tbody>
        </table>
      }
    </div>
  `,
  styles: [`
    .card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
    .empty { color: #6c7086; text-align: center; padding: 2rem; }
    .data-table { width: 100%; border-collapse: collapse; }
    .data-table th, .data-table td { padding: 0.5rem 0.75rem; text-align: right; border-bottom: 1px solid #313244; }
    .data-table th:first-child, .data-table td:first-child { text-align: left; }
    .symbol { font-weight: 600; color: #89b4fa; }
    .pnl-positive { color: #a6e3a1; font-weight: 600; }
    .pnl-negative { color: #f38ba8; font-weight: 600; }
    .btn-ghost { background: transparent; border: 1px solid #45475a; color: #cdd6f4; padding: 0.25rem 0.75rem; border-radius: 4px; cursor: pointer; }
    .btn-danger-sm { background: transparent; border: 1px solid #f38ba8; color: #f38ba8; padding: 0.2rem 0.5rem; border-radius: 4px; cursor: pointer; font-size: 0.8rem; }
  `],
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
      next: positions => this.positions.set(positions),
    });
  }

  closePosition(pos: Position): void {
    this.api.placeOrder({
      symbol: pos.symbol,
      side: 'sell',
      type: 'market',
      qty: pos.qty,
    }).subscribe({ next: () => this.refresh() });
  }
}
