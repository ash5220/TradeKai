import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { NavbarComponent } from './shared/components/navbar.component';

@Component({
  selector: 'tk-root',
  standalone: true,
  imports: [RouterOutlet, NavbarComponent],
  template: `
    <tk-navbar />
    <main class="main-content">
      <router-outlet />
    </main>
  `,
  styles: [`
    :host {
      display: flex;
      flex-direction: column;
      min-height: 100vh;
    }
    .main-content {
      flex: 1;
      padding: 1.5rem;
      background: #0f1117;
    }
  `],
})
export class AppComponent {}
