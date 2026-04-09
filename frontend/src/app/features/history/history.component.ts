import { Component, inject, OnInit, signal, computed } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { DatePipe, NgClass, DecimalPipe } from "@angular/common";
import { ApiService } from "../../core/api/api.service";
import { Order } from "../../shared/models/order.model";

@Component({
  selector: "tk-history",
  standalone: true,
  imports: [FormsModule, DatePipe, NgClass, DecimalPipe],
  templateUrl: "./history.component.html",
  styleUrl: "./history.component.scss",
})
export class HistoryComponent implements OnInit {
  private readonly api = inject(ApiService);

  readonly loading = signal(true);
  readonly orders = signal<Order[]>([]);

  readonly filterSymbol = signal("");
  readonly filterSide = signal("");

  readonly filtered = computed(() =>
    this.orders().filter((o) => {
      const symMatch =
        !this.filterSymbol() ||
        o.symbol.includes(this.filterSymbol().toUpperCase());
      const sideMatch = !this.filterSide() || o.side === this.filterSide();
      return symMatch && sideMatch;
    }),
  );

  ngOnInit(): void {
    this.api.getOrders(200).subscribe({
      next: (orders) => {
        this.orders.set(orders);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }

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
