import { Component, inject, OnInit, signal } from "@angular/core";
import { NgClass } from "@angular/common";
import { ApiService, StrategyInfo } from "../../../core/api/api.service";

@Component({
  selector: "tk-strategy-list",
  standalone: true,
  imports: [NgClass],
  templateUrl: "./strategy-list.component.html",
  styleUrl: "./strategy-list.component.scss",
})
export class StrategyListComponent implements OnInit {
  private readonly api = inject(ApiService);
  readonly strategies = signal<StrategyInfo[]>([]);

  ngOnInit(): void {
    this.api.getStrategies().subscribe({
      next: (list) => this.strategies.set(list),
    });
  }

  start(s: StrategyInfo): void {
    this.api.startStrategy(s.name, [s.symbol]).subscribe({
      next: () =>
        this.strategies.update((all) =>
          all.map((x) =>
            x.name === s.name && x.symbol === s.symbol
              ? { ...x, active: true }
              : x,
          ),
        ),
    });
  }

  stop(s: StrategyInfo): void {
    this.api.stopStrategy(s.name).subscribe({
      next: () =>
        this.strategies.update((all) =>
          all.map((x) =>
            x.name === s.name && x.symbol === s.symbol
              ? { ...x, active: false }
              : x,
          ),
        ),
    });
  }
}
