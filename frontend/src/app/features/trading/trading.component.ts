import { Component } from '@angular/core';
import { StrategyListComponent } from './components/strategy-list.component';
import { OrderFormComponent } from './components/order-form.component';
import { TradeLogComponent } from './components/trade-log.component';

@Component({
  selector: 'tk-trading',
  standalone: true,
  imports: [StrategyListComponent, OrderFormComponent, TradeLogComponent],
  template: `
    <div class="trading-layout">
      <div class="trading-sidebar">
        <tk-strategy-list />
        <tk-order-form />
      </div>
      <div class="trading-main">
        <tk-trade-log />
      </div>
    </div>
  `,
  styles: [`
    .trading-layout { display: grid; grid-template-columns: 380px 1fr; gap: 1.5rem; }
    .trading-sidebar { display: flex; flex-direction: column; gap: 1rem; }
  `],
})
export class TradingComponent {}
