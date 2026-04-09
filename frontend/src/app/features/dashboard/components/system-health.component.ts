import { Component, inject, computed } from '@angular/core';
import { NgClass } from '@angular/common';
import { WebSocketService, ConnectionState } from '../../../core/websocket/websocket.service';

@Component({
  selector: 'tk-system-health',
  standalone: true,
  imports: [NgClass],
  template: `
    <div class="card metric-card">
      <div class="metric-label">System Status</div>
      <div class="status-row">
        <span class="status-dot" [ngClass]="dotClass()"></span>
        <span class="status-text">{{ statusLabel() }}</span>
      </div>
      @if (state() === 'error') {
        <p class="status-hint">Attempting to reconnect…</p>
      }
    </div>
  `,
  styles: [`
    .metric-card { display: flex; flex-direction: column; gap: 0.75rem; }
    .metric-label { font-size: 0.8rem; color: #6c7086; text-transform: uppercase; letter-spacing: 0.05em; }
    .status-row { display: flex; align-items: center; gap: 0.5rem; }
    .status-dot { width: 12px; height: 12px; border-radius: 50%; flex-shrink: 0; }
    .dot-connected { background: #a6e3a1; box-shadow: 0 0 6px #a6e3a1; }
    .dot-connecting { background: #f9e2af; }
    .dot-disconnected { background: #6c7086; }
    .dot-error { background: #f38ba8; box-shadow: 0 0 6px #f38ba8; }
    .status-text { font-size: 1.1rem; font-weight: 600; color: #cdd6f4; }
    .status-hint { font-size: 0.8rem; color: #f9e2af; margin: 0; }
  `],
})
export class SystemHealthComponent {
  private readonly ws = inject(WebSocketService);
  readonly state = this.ws.connectionState;

  readonly dotClass = computed(() => {
    const map: Record<ConnectionState, string> = {
      connected: 'dot-connected',
      connecting: 'dot-connecting',
      disconnected: 'dot-disconnected',
      error: 'dot-error',
    };
    return map[this.state()];
  });

  readonly statusLabel = computed(() => {
    const map: Record<ConnectionState, string> = {
      connected: 'Live',
      connecting: 'Connecting…',
      disconnected: 'Disconnected',
      error: 'Error',
    };
    return map[this.state()];
  });
}
