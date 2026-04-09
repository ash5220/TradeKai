import { Routes } from '@angular/router';

export const TRADING_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./trading.component').then(m => m.TradingComponent),
  },
];
