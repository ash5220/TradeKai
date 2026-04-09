import { Component, inject, OnInit, signal } from "@angular/core";
import { NgClass } from "@angular/common";
import { ApiService } from "../../../core/api/api.service";

export interface StrategyInfo {
  name: string;
  symbol: string;
  active: boolean;
}

@Component({
  selector: "tk-strategy-list",
  standalone: true,
  imports: [NgClass],
  template: `
    <div class="card">
      <h3>Strategies</h3>
      @if (strategies().length === 0) {
        <p class="empty">No strategies configured.</p>
      } @else {
        <ul class="strategy-list">
          @for (s of strategies(); track s.name + s.symbol) {
            <li class="strategy-item">
              <div class="strategy-info">
                <span class="strategy-name">{{ s.name }}</span>
                <span class="strategy-symbol">{{ s.symbol }}</span>
              </div>
              <div class="strategy-controls">
                <span
                  class="badge"
                  [ngClass]="s.active ? 'badge-active' : 'badge-idle'"
                >
                  {{ s.active ? "Active" : "Idle" }}
                </span>
                @if (s.active) {
                  <button class="btn-danger-sm" (click)="stop(s)">Stop</button>
                } @else {
                  <button class="btn-success-sm" (click)="start(s)">
                    Start
                  </button>
                }
              </div>
            </li>
          }
        </ul>
      }
    </div>
  `,
  styles: [
    `
      h3 {
        margin-bottom: 0.75rem;
      }
      .empty {
        color: #6c7086;
        font-size: 0.9rem;
      }
      .strategy-list {
        list-style: none;
        padding: 0;
        margin: 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
      }
      .strategy-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.6rem;
        border-radius: 6px;
        background: #181825;
      }
      .strategy-info {
        display: flex;
        flex-direction: column;
        gap: 0.15rem;
      }
      .strategy-name {
        font-weight: 600;
        font-size: 0.9rem;
      }
      .strategy-symbol {
        font-size: 0.75rem;
        color: #89b4fa;
      }
      .strategy-controls {
        display: flex;
        align-items: center;
        gap: 0.5rem;
      }
      .badge {
        font-size: 0.7rem;
        padding: 0.15rem 0.5rem;
        border-radius: 999px;
        font-weight: 600;
      }
      .badge-active {
        background: #a6e3a133;
        color: #a6e3a1;
      }
      .badge-idle {
        background: #45475a;
        color: #6c7086;
      }
      .btn-danger-sm {
        background: transparent;
        border: 1px solid #f38ba8;
        color: #f38ba8;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
        cursor: pointer;
        font-size: 0.8rem;
      }
      .btn-success-sm {
        background: transparent;
        border: 1px solid #a6e3a1;
        color: #a6e3a1;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
        cursor: pointer;
        font-size: 0.8rem;
      }
    `,
  ],
})
export class StrategyListComponent implements OnInit {
  private readonly api = inject(ApiService);
  readonly strategies = signal<StrategyInfo[]>([]);

  ngOnInit(): void {
    this.api.getStrategies().subscribe({
      next: (list) => this.strategies.set(list),
    });
  }

  start(s: StrategyInfo): void {
    this.api.startStrategy(s.name, [s.symbol]).subscribe({
      next: () =>
        this.strategies.update((all) =>
          all.map((x) =>
            x.name === s.name && x.symbol === s.symbol
              ? { ...x, active: true }
              : x,
          ),
        ),
    });
  }

  stop(s: StrategyInfo): void {
    this.api.stopStrategy(s.name).subscribe({
      next: () =>
        this.strategies.update((all) =>
          all.map((x) =>
            x.name === s.name && x.symbol === s.symbol
              ? { ...x, active: false }
              : x,
          ),
        ),
    });
  }
}
