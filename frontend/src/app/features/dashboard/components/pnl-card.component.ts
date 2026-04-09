import { Component, inject, OnInit, signal } from "@angular/core";
import { DecimalPipe, NgClass } from "@angular/common";
import { ApiService, PnlSummary } from "../../../core/api/api.service";

@Component({
  selector: "tk-pnl-card",
  standalone: true,
  imports: [DecimalPipe, NgClass],
  template: `
    <div class="card metric-card">
      <div class="metric-label">Daily P&L</div>
      <div class="metric-subrow">
        <span class="metric-label">Daily Realized P&amp;L</span>
        <span
          class="metric-pnl"
          [ngClass]="
            (summary()?.daily_realized_pnl ?? 0) >= 0 ? 'positive' : 'negative'
          "
        >
          {{ summary()?.daily_realized_pnl | number: "1.2-2" }}
        </span>
      </div>
    </div>
  `,
  styles: [
    `
      .metric-card {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
      }
      .metric-label {
        font-size: 0.8rem;
        color: #6c7086;
        text-transform: uppercase;
        letter-spacing: 0.05em;
      }
      .metric-value {
        font-size: 2rem;
        font-weight: 700;
        color: #cdd6f4;
      }
      .metric-subrow {
        display: flex;
        justify-content: space-between;
        align-items: center;
      }
      .metric-pnl {
        font-weight: 600;
      }
      .positive {
        color: #a6e3a1;
      }
      .negative {
        color: #f38ba8;
      }
    `,
  ],
})
export class PnlCardComponent implements OnInit {
  private readonly api = inject(ApiService);
  readonly summary = signal<PnlSummary | null>(null);

  ngOnInit(): void {
    this.api.getPnL().subscribe({
      next: (s) => this.summary.set(s),
    });
  }
}
