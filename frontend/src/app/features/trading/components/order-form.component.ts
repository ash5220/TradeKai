import { Component, inject, signal } from "@angular/core";
import { ReactiveFormsModule, FormBuilder, Validators } from "@angular/forms";
import { NgClass } from "@angular/common";
import { ApiService } from "../../../core/api/api.service";

@Component({
  selector: "tk-order-form",
  standalone: true,
  imports: [ReactiveFormsModule, NgClass],
  template: `
    <div class="card">
      <h3>Place Order</h3>
      <form [formGroup]="form" (ngSubmit)="submit()" class="order-form">
        <div class="field">
          <label>Symbol</label>
          <input formControlName="symbol" placeholder="e.g. AAPL" />
        </div>
        <div class="field-row">
          <div class="field">
            <label>Side</label>
            <select formControlName="side">
              <option value="buy">Buy</option>
              <option value="sell">Sell</option>
            </select>
          </div>
          <div class="field">
            <label>Type</label>
            <select formControlName="type">
              <option value="market">Market</option>
              <option value="limit">Limit</option>
            </select>
          </div>
        </div>
        <div class="field-row">
          <div class="field">
            <label>Quantity</label>
            <input
              type="number"
              formControlName="quantity"
              min="0.0001"
              step="0.0001"
            />
          </div>
          @if (form.value.type === "limit") {
            <div class="field">
              <label>Limit Price</label>
              <input
                type="number"
                formControlName="limitPrice"
                min="0.01"
                step="0.01"
              />
            </div>
          }
        </div>
        @if (error()) {
          <p class="form-error">{{ error() }}</p>
        }
        @if (success()) {
          <p class="form-success">Order submitted.</p>
        }
        <button
          type="submit"
          class="btn-primary"
          [disabled]="form.invalid || loading()"
        >
          {{ loading() ? "Submitting…" : "Submit Order" }}
        </button>
      </form>
    </div>
  `,
  styles: [
    `
      h3 {
        margin-bottom: 0.75rem;
      }
      .order-form {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
      }
      .field {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        flex: 1;
      }
      .field-row {
        display: flex;
        gap: 0.75rem;
      }
      label {
        font-size: 0.8rem;
        color: #6c7086;
        text-transform: uppercase;
        letter-spacing: 0.04em;
      }
      input,
      select {
        padding: 0.4rem 0.6rem;
        border-radius: 4px;
        border: 1px solid #45475a;
        background: #181825;
        color: #cdd6f4;
        font-size: 0.95rem;
      }
      input:focus,
      select:focus {
        outline: none;
        border-color: #89b4fa;
      }
      .btn-primary {
        padding: 0.5rem;
        border-radius: 4px;
        border: none;
        background: #89b4fa;
        color: #1e1e2e;
        font-weight: 700;
        cursor: pointer;
      }
      .btn-primary:disabled {
        opacity: 0.5;
        cursor: not-allowed;
      }
      .form-error {
        color: #f38ba8;
        font-size: 0.85rem;
        margin: 0;
      }
      .form-success {
        color: #a6e3a1;
        font-size: 0.85rem;
        margin: 0;
      }
    `,
  ],
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
