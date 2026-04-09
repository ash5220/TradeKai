import { Component } from "@angular/core";
import { StrategyListComponent } from "./strategy-list/strategy-list.component";
import { OrderFormComponent } from "./order-form/order-form.component";
import { TradeLogComponent } from "./trade-log/trade-log.component";

@Component({
  selector: "tk-trading",
  standalone: true,
  imports: [StrategyListComponent, OrderFormComponent, TradeLogComponent],
  templateUrl: "./trading.component.html",
  styleUrl: "./trading.component.scss",
})
export class TradingComponent {}
