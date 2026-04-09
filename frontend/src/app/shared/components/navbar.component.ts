import { Component, inject } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { AuthService } from '../core/auth/auth.service';

@Component({
  selector: 'tk-navbar',
  standalone: true,
  imports: [RouterLink, RouterLinkActive],
  template: `
    <nav class="navbar">
      <a class="brand" routerLink="/dashboard">TradeKai</a>
      @if (auth.isAuthenticated()) {
        <ul class="nav-links">
          <li><a routerLink="/dashboard" routerLinkActive="active">Dashboard</a></li>
          <li><a routerLink="/trading" routerLinkActive="active">Trading</a></li>
          <li><a routerLink="/history" routerLinkActive="active">History</a></li>
        </ul>
        <button class="btn-logout" (click)="auth.logout()">Logout</button>
      }
    </nav>
  `,
  styles: [`
    .navbar {
      display: flex;
      align-items: center;
      gap: 1.5rem;
      padding: 0.75rem 1.5rem;
      background: #1a1d27;
      border-bottom: 1px solid #2a2d3a;
    }
    .brand {
      font-size: 1.25rem;
      font-weight: 700;
      color: #00d1b2;
      text-decoration: none;
    }
    .nav-links {
      display: flex;
      gap: 1rem;
      list-style: none;
      margin: 0;
      padding: 0;
    }
    .nav-links a {
      color: #aaa;
      text-decoration: none;
      &.active { color: #fff; }
    }
    .btn-logout {
      margin-left: auto;
      background: transparent;
      border: 1px solid #555;
      color: #aaa;
      padding: 0.4rem 1rem;
      border-radius: 4px;
      cursor: pointer;
      &:hover { background: #2a2d3a; }
    }
  `],
})
export class NavbarComponent {
  protected readonly auth = inject(AuthService);
}
