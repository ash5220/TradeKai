import { Component, inject, computed } from '@angular/core';
import { DatePipe, NgClass } from '@angular/common';
import { WebSocketService } from '../../../core/websocket/websocket.service';
import { Order } from '../../../shared/models/order.model';

@Component({
  selector: 'tk-trade-log',
  standalone: true,
  imports: [DatePipe, NgClass],
  template: `
    <div class="card trade-log-card">
      <h3>Live Trade Log</h3>
      @if (orders().length === 0) {
        <p class="empty">Waiting for orders…</p>
      } @else {
        <div class="log-scroll">
          <table class="data-table">
            <thead>
              <tr>
                <th>Time</th>
                <th>Symbol</th>
                <th>Side</th>
                <th>Type</th>
                <th>Qty</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              @for (o of orders(); track o.id) {
                <tr>
                  <td>{{ o.created_at | date:'HH:mm:ss' }}</td>
                  <td class="symbol">{{ o.symbol }}</td>
                  <td [ngClass]="o.side === 'buy' ? 'side-buy' : 'side-sell'">{{ o.side }}</td>
                  <td>{{ o.type }}</td>
                  <td>{{ o.qty }}</td>
                  <td>
                    <span class="badge" [ngClass]="statusClass(o.status)">{{ o.status }}</span>
                  </td>
                </tr>
              }
            </tbody>
          </table>
        </div>
      }
    </div>
  `,
  styles: [`
    .trade-log-card { height: 100%; }
    h3 { margin-bottom: 0.75rem; }
    .empty { color: #6c7086; text-align: center; padding: 2rem; }
    .log-scroll { max-height: 600px; overflow-y: auto; }
    .data-table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
    .data-table th, .data-table td { padding: 0.4rem 0.75rem; text-align: right; border-bottom: 1px solid #313244; }
    .data-table th:first-child, .data-table td:first-child { text-align: left; }
    .symbol { color: #89b4fa; font-weight: 600; }
    .side-buy { color: #a6e3a1; }
    .side-sell { color: #f38ba8; }
    .badge { font-size: 0.7rem; padding: 0.1rem 0.4rem; border-radius: 999px; }
    .badge-pending { background: #f9e2af33; color: #f9e2af; }
    .badge-filled { background: #a6e3a133; color: #a6e3a1; }
    .badge-cancelled { background: #45475a; color: #6c7086; }
    .badge-rejected { background: #f38ba833; color: #f38ba8; }
    .badge-submitted { background: #89b4fa33; color: #89b4fa; }
  `],
})
export class TradeLogComponent {
  private readonly ws = inject(WebSocketService);
  readonly orders = this.ws.latestOrders;

  statusClass(status: string): string {
    const map: Record<string, string> = {
      pending: 'badge-pending',
      submitted: 'badge-submitted',
      filled: 'badge-filled',
      cancelled: 'badge-cancelled',
      rejected: 'badge-rejected',
    };
    return map[status] ?? 'badge-pending';
  }
}
