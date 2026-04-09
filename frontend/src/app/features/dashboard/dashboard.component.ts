import { Component, inject, OnInit } from "@angular/core";
import { WebSocketService } from "../../core/websocket/websocket.service";
import { PriceChartComponent } from "./price-chart/price-chart.component";
import { PositionTableComponent } from "./position-table/position-table.component";
import { PnlCardComponent } from "./pnl-card/pnl-card.component";
import { SystemHealthComponent } from "./system-health/system-health.component";

@Component({
  selector: "tk-dashboard",
  standalone: true,
  imports: [
    PriceChartComponent,
    PositionTableComponent,
    PnlCardComponent,
    SystemHealthComponent,
  ],
  templateUrl: "./dashboard.component.html",
  styleUrl: "./dashboard.component.scss",
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
