import { Component, inject, OnInit, signal } from "@angular/core";
import { DecimalPipe, NgClass } from "@angular/common";
import { ApiService, PnlSummary } from "../../../core/api/api.service";

@Component({
  selector: "tk-pnl-card",
  standalone: true,
  imports: [DecimalPipe, NgClass],
  templateUrl: "./pnl-card.component.html",
  styleUrl: "./pnl-card.component.scss",
})
export class PnlCardComponent implements OnInit {
  private readonly api = inject(ApiService);
  readonly summary = signal<PnlSummary | null>(null);

  ngOnInit(): void {
    this.api.getPnL().subscribe({
      next: (s) => this.summary.set(s),
    });
  }
}
