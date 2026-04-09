import { Component, inject, OnInit, signal, computed } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { DatePipe, NgClass, DecimalPipe } from "@angular/common";
import { ApiService } from "../../core/api/api.service";
import { Order } from "../../shared/models/order.model";

@Component({
  selector: "tk-history",
  standalone: true,
  imports: [FormsModule, DatePipe, NgClass, DecimalPipe],
  template: `
    <div class="card">
      <div class="page-header">
        <h2>Trade History</h2>
        <div class="filters">
          <input
            [(ngModel)]="filterSymbol"
            placeholder="Filter by symbol…"
            class="filter-input"
          />
          <select [(ngModel)]="filterSide" class="filter-select">
            <option value="">All sides</option>
            <option value="buy">Buy</option>
            <option value="sell">Sell</option>
          </select>
        </div>
      </div>

      @if (loading()) {
        <p class="state-msg">Loading…</p>
      } @else if (filtered().length === 0) {
        <p class="state-msg">No trades found.</p>
      } @else {
        <table class="data-table">
          <thead>
            <tr>
              <th>Date</th>
              <th>Symbol</th>
              <th>Side</th>
              <th>Type</th>
              <th>Qty</th>
              <th>Avg Fill</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            @for (o of filtered(); track o.id) {
              <tr>
                <td>{{ o.created_at | date: "medium" }}</td>
                <td class="symbol">{{ o.symbol }}</td>
                <td [ngClass]="o.side === 'buy' ? 'side-buy' : 'side-sell'">
                  {{ o.side }}
                </td>
                <td>{{ o.type }}</td>
                <td>{{ o.qty | number: "1.0-4" }}</td>
                <td>
                  {{
                    o.filled_avg_price
                      ? (o.filled_avg_price | number: "1.2-2")
                      : "—"
                  }}
                </td>
                <td>
                  <span class="badge" [ngClass]="statusClass(o.status)">{{
                    o.status
                  }}</span>
                </td>
              </tr>
            }
          </tbody>
        </table>
      }
    </div>
  `,
  styles: [
    `
      .page-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1rem;
      }
      .filters {
        display: flex;
        gap: 0.5rem;
      }
      .filter-input,
      .filter-select {
        padding: 0.4rem 0.6rem;
        border-radius: 4px;
        border: 1px solid #45475a;
        background: #181825;
        color: #cdd6f4;
      }
      .state-msg {
        color: #6c7086;
        text-align: center;
        padding: 2rem;
      }
      .data-table {
        width: 100%;
        border-collapse: collapse;
      }
      .data-table th,
      .data-table td {
        padding: 0.5rem 0.75rem;
        text-align: right;
        border-bottom: 1px solid #313244;
      }
      .data-table th:first-child,
      .data-table td:first-child {
        text-align: left;
      }
      .symbol {
        color: #89b4fa;
        font-weight: 600;
      }
      .side-buy {
        color: #a6e3a1;
      }
      .side-sell {
        color: #f38ba8;
      }
      .badge {
        font-size: 0.7rem;
        padding: 0.1rem 0.4rem;
        border-radius: 999px;
      }
      .badge-pending {
        background: #f9e2af33;
        color: #f9e2af;
      }
      .badge-filled {
        background: #a6e3a133;
        color: #a6e3a1;
      }
      .badge-cancelled {
        background: #45475a;
        color: #6c7086;
      }
      .badge-rejected {
        background: #f38ba833;
        color: #f38ba8;
      }
      .badge-submitted {
        background: #89b4fa33;
        color: #89b4fa;
      }
    `,
  ],
})
export class HistoryComponent implements OnInit {
  private readonly api = inject(ApiService);

  readonly loading = signal(true);
  readonly orders = signal<Order[]>([]);

  filterSymbol = "";
  filterSide = "";

  readonly filtered = computed(() =>
    this.orders().filter((o) => {
      const symMatch =
        !this.filterSymbol ||
        o.symbol.includes(this.filterSymbol.toUpperCase());
      const sideMatch = !this.filterSide || o.side === this.filterSide;
      return symMatch && sideMatch;
    }),
  );

  ngOnInit(): void {
    this.api.getOrders(200).subscribe({
      next: (orders) => {
        this.orders.set(orders);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }

  statusClass(status: string): string {
    const map: Record<string, string> = {
      pending: "badge-pending",
      submitted: "badge-submitted",
      filled: "badge-filled",
      cancelled: "badge-cancelled",
      rejected: "badge-rejected",
    };
    return map[status] ?? "badge-pending";
  }
}
