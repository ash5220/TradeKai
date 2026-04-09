import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { OrderFormComponent } from "./order-form.component";
import { ApiService } from "../../../core/api/api.service";
import { of, throwError } from "rxjs";

describe("OrderFormComponent", () => {
  let fixture: ComponentFixture<OrderFormComponent>;
  let apiSpy: jasmine.SpyObj<ApiService>;

  beforeEach(async () => {
    apiSpy = jasmine.createSpyObj<ApiService>("ApiService", ["placeOrder"]);
    apiSpy.placeOrder.and.returnValue(of({} as any));

    await TestBed.configureTestingModule({
      imports: [OrderFormComponent],
      providers: [{ provide: ApiService, useValue: apiSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(OrderFormComponent);
    fixture.detectChanges();
  });

  function getSubmitButton(): HTMLButtonElement {
    return fixture.nativeElement.querySelector('button[type="submit"]');
  }

  function setFormField(controlName: string, value: unknown): void {
    fixture.componentInstance.form.get(controlName)?.setValue(value);
    fixture.detectChanges();
  }

  function fillValidMarketOrder(): void {
    setFormField("symbol", "AAPL");
    setFormField("side", "buy");
    setFormField("type", "market");
    setFormField("quantity", 1);
  }

  // ── Happy paths ──────────────────────────────────────────────────────────

  it("submit button is disabled on load (empty symbol)", () => {
    expect(getSubmitButton().disabled).toBeTrue();
  });

  it("submits a valid market order and calls placeOrder", () => {
    fillValidMarketOrder();
    getSubmitButton().click();
    expect(apiSpy.placeOrder).toHaveBeenCalledWith(
      jasmine.objectContaining({
        symbol: "AAPL",
        side: "buy",
        type: "market",
        qty: 1,
      }),
    );
  });

  it("shows success message after successful order", () => {
    fillValidMarketOrder();
    getSubmitButton().click();
    fixture.detectChanges();
    const successEl = fixture.nativeElement.querySelector(".form-success");
    expect(successEl).not.toBeNull();
    expect(successEl.textContent).toContain("Order submitted");
  });

  it("resets the form after successful order", () => {
    fillValidMarketOrder();
    getSubmitButton().click();
    fixture.detectChanges();
    expect(fixture.componentInstance.form.get("symbol")?.value).toBeFalsy();
  });

  it("includes limit_price for limit orders", () => {
    setFormField("symbol", "AAPL");
    setFormField("type", "limit");
    setFormField("quantity", 5);
    setFormField("limitPrice", 150.5);
    getSubmitButton().click();
    expect(apiSpy.placeOrder).toHaveBeenCalledWith(
      jasmine.objectContaining({ limit_price: 150.5 }),
    );
  });

  // ── Edge cases ───────────────────────────────────────────────────────────

  it("shows limit price field when order type is limit", () => {
    setFormField("symbol", "AAPL");
    setFormField("type", "limit");
    const limitPriceField = fixture.nativeElement.querySelector('input[formcontrolname="limitPrice"]');
    expect(limitPriceField).not.toBeNull();
  });

  it("hides limit price field when order type is market", () => {
    setFormField("type", "market");
    const limitPriceField = fixture.nativeElement.querySelector('input[formcontrolname="limitPrice"]');
    expect(limitPriceField).toBeNull();
  });

  it("quantity of 0.0001 is valid (boundary)", () => {
    setFormField("symbol", "AAPL");
    setFormField("quantity", 0.0001);
    expect(fixture.componentInstance.form.get("quantity")?.valid).toBeTrue();
  });

  it("quantity of 0.00009 is invalid (below minimum)", () => {
    setFormField("quantity", 0.00009);
    expect(fixture.componentInstance.form.get("quantity")?.valid).toBeFalse();
  });

  it("symbol must match uppercase letter pattern", () => {
    setFormField("symbol", "12345");
    expect(fixture.componentInstance.form.get("symbol")?.valid).toBeFalse();
  });

  it("valid symbol 'AAPL' passes pattern validation", () => {
    setFormField("symbol", "AAPL");
    expect(fixture.componentInstance.form.get("symbol")?.valid).toBeTrue();
  });

  // ── Error cases ──────────────────────────────────────────────────────────

  it("shows error message when placeOrder fails", () => {
    apiSpy.placeOrder.and.returnValue(
      throwError(() => new Error("Order rejected")),
    );
    fillValidMarketOrder();
    getSubmitButton().click();
    fixture.detectChanges();
    const errorEl = fixture.nativeElement.querySelector(".form-error");
    expect(errorEl).not.toBeNull();
    expect(errorEl.textContent).toContain("Order rejected");
  });

  it("re-enables submit button after error", () => {
    apiSpy.placeOrder.and.returnValue(throwError(() => new Error("fail")));
    fillValidMarketOrder();
    getSubmitButton().click();
    fixture.detectChanges();
    expect(getSubmitButton().disabled).toBeFalse();
  });

  it("does not show success message on error", () => {
    apiSpy.placeOrder.and.returnValue(throwError(() => new Error("fail")));
    fillValidMarketOrder();
    getSubmitButton().click();
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector(".form-success")).toBeNull();
  });
});
