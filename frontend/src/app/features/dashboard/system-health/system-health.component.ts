import { Component, inject, computed } from "@angular/core";
import { NgClass } from "@angular/common";
import {
  WebSocketService,
  ConnectionState,
} from "../../../core/websocket/websocket.service";

@Component({
  selector: "tk-system-health",
  standalone: true,
  imports: [NgClass],
  templateUrl: "./system-health.component.html",
  styleUrl: "./system-health.component.scss",
})
export class SystemHealthComponent {
  private readonly ws = inject(WebSocketService);
  readonly state = this.ws.connectionState;

  readonly dotClass = computed(() => {
    const map: Record<ConnectionState, string> = {
      connected: "dot-connected",
      connecting: "dot-connecting",
      disconnected: "dot-disconnected",
      error: "dot-error",
    };
    return map[this.state()];
  });

  readonly statusLabel = computed(() => {
    const map: Record<ConnectionState, string> = {
      connected: "Live",
      connecting: "Connecting…",
      disconnected: "Disconnected",
      error: "Error",
    };
    return map[this.state()];
  });
}
