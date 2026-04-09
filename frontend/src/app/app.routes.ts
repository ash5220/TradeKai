import type { Routes } from '@angular/router';
import { authGuard } from './core/auth/auth.guard';

export const routes: Routes = [
  {
    path: 'auth',
    loadChildren: () =>
      import('./features/auth/auth.routes').then(m => m.AUTH_ROUTES),
  },
  {
    path: 'dashboard',
    canActivate: [authGuard],
    loadChildren: () =>
      import('./features/dashboard/dashboard.routes').then(m => m.DASHBOARD_ROUTES),
  },
  {
    path: 'trading',
    canActivate: [authGuard],
    loadChildren: () =>
      import('./features/trading/trading.routes').then(m => m.TRADING_ROUTES),
  },
  {
    path: 'history',
    canActivate: [authGuard],
    loadChildren: () =>
      import('./features/history/history.routes').then(m => m.HISTORY_ROUTES),
  },
  { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
  { path: '**', redirectTo: 'dashboard' },
];
