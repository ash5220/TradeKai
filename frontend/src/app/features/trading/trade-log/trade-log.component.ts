import { Component, inject, computed } from "@angular/core";
import { DatePipe, NgClass } from "@angular/common";
import { WebSocketService } from "../../../core/websocket/websocket.service";
import { Order } from "../../../shared/models/order.model";

@Component({
  selector: "tk-trade-log",
  standalone: true,
  imports: [DatePipe, NgClass],
  templateUrl: "./trade-log.component.html",
  styleUrl: "./trade-log.component.scss",
})
export class TradeLogComponent {
  private readonly ws = inject(WebSocketService);
  readonly orders = this.ws.latestOrders;

  statusClass(status: string): string {
    const map: Record<string, string> = {
      pending: "badge-pending",
      submitted: "badge-submitted",
      filled: "badge-filled",
      cancelled: "badge-cancelled",
      rejected: "badge-rejected",
    };
    return map[status] ?? "badge-pending";
  }
}
