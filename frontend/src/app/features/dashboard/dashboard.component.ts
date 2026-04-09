import { Component, inject, OnInit } from "@angular/core";
import { WebSocketService } from "../../core/websocket/websocket.service";
import { PriceChartComponent } from "./components/price-chart.component";
import { PositionTableComponent } from "./components/position-table.component";
import { PnlCardComponent } from "./components/pnl-card.component";
import { SystemHealthComponent } from "./components/system-health.component";

@Component({
  selector: "tk-dashboard",
  standalone: true,
  imports: [
    PriceChartComponent,
    PositionTableComponent,
    PnlCardComponent,
    SystemHealthComponent,
  ],
  template: `
    <div class="dashboard">
      <div class="top-row">
        <tk-pnl-card />
        <tk-system-health />
      </div>
      <tk-price-chart />
      <tk-position-table />
    </div>
  `,
  styles: [
    `
      .dashboard {
        display: flex;
        flex-direction: column;
        gap: 1.5rem;
      }
      .top-row {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 1rem;
      }
    `,
  ],
})
export class DashboardComponent implements OnInit {
  private readonly ws = inject(WebSocketService);

  ngOnInit(): void {
    this.ws.connect();
    // Subscribe to default symbols
    ["AAPL", "TSLA", "MSFT", "GOOGL"].forEach((sym) =>
      this.ws.subscribe(`ticks:${sym}`),
    );
  }
}
