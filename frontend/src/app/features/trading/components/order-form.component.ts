import { Component, inject, signal } from "@angular/core";
import { ReactiveFormsModule, FormBuilder, Validators } from "@angular/forms";
import { NgClass } from "@angular/common";
import { ApiService } from "../../../core/api/api.service";

@Component({
  selector: "tk-order-form",
  standalone: true,
  imports: [ReactiveFormsModule, NgClass],
  templateUrl: "./order-form.component.html",
  styleUrl: "./order-form.component.scss",
})
export class OrderFormComponent {
  private readonly api = inject(ApiService);
  private readonly fb = inject(FormBuilder);

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly success = signal(false);

  readonly form = this.fb.group({
    symbol: ["", [Validators.required, Validators.pattern(/^[A-Z]{1,5}$/)]],
    side: ["buy", Validators.required],
    type: ["market", Validators.required],
    quantity: [1, [Validators.required, Validators.min(0.0001)]],
    limitPrice: [null as number | null],
  });

  submit(): void {
    if (this.form.invalid) return;
    this.loading.set(true);
    this.error.set(null);
    this.success.set(false);

    const { symbol, side, type, quantity, limitPrice } =
      this.form.getRawValue();
    this.api
      .placeOrder({
        symbol: symbol!.toUpperCase(),
        side: side as "buy" | "sell",
        type: type as "market" | "limit",
        qty: quantity!,
        limit_price: limitPrice ?? undefined,
      })
      .subscribe({
        next: () => {
          this.loading.set(false);
          this.success.set(true);
          this.form.reset({ side: "buy", type: "market", quantity: 1 });
        },
        error: (err: Error) => {
          this.loading.set(false);
          this.error.set(err.message ?? "Failed to place order.");
        },
      });
  }
}
