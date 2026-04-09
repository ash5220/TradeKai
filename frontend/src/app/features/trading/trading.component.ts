import { Component } from "@angular/core";
import { StrategyListComponent } from "./components/strategy-list.component";
import { OrderFormComponent } from "./components/order-form.component";
import { TradeLogComponent } from "./components/trade-log.component";

@Component({
  selector: "tk-trading",
  standalone: true,
  imports: [StrategyListComponent, OrderFormComponent, TradeLogComponent],
  templateUrl: "./trading.component.html",
  styleUrl: "./trading.component.scss",
})
export class TradingComponent {}
