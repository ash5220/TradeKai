import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { HistoryComponent } from "./history.component";
import { ApiService } from "../../core/api/api.service";
import { of, throwError } from "rxjs";
import type { Order } from "../../shared/models/order.model";

const mockOrders: Order[] = [
  {
    id: "1",
    symbol: "AAPL",
    side: "buy",
    type: "market",
    qty: 10,
    status: "filled",
    created_at: "2024-01-01T10:00:00Z",
  } as Order,
  {
    id: "2",
    symbol: "TSLA",
    side: "sell",
    type: "limit",
    qty: 5,
    status: "pending",
    created_at: "2024-01-01T11:00:00Z",
  } as Order,
  {
    id: "3",
    symbol: "AAPL",
    side: "sell",
    type: "market",
    qty: 3,
    status: "cancelled",
    created_at: "2024-01-01T12:00:00Z",
  } as Order,
];

describe("HistoryComponent", () => {
  let fixture: ComponentFixture<HistoryComponent>;
  let apiSpy: jasmine.SpyObj<ApiService>;

  async function setup(orders: Order[], shouldFail = false): Promise<void> {
    apiSpy = jasmine.createSpyObj<ApiService>("ApiService", ["getOrders"]);

    if (shouldFail) {
      apiSpy.getOrders.and.returnValue(
        throwError(() => new Error("Network error")),
      );
    } else {
      apiSpy.getOrders.and.returnValue(of(orders));
    }

    await TestBed.configureTestingModule({
      imports: [HistoryComponent],
      providers: [{ provide: ApiService, useValue: apiSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(HistoryComponent);
    fixture.detectChanges();
    // Wait for async operations
    await fixture.whenStable();
    fixture.detectChanges();
  }

  // ── Happy paths ──────────────────────────────────────────────────────────

  it("calls getOrders with limit=200 on init", async () => {
    await setup(mockOrders);
    expect(apiSpy.getOrders).toHaveBeenCalledWith(200);
  });

  it("renders a row for each order", async () => {
    await setup(mockOrders);
    const rows = fixture.nativeElement.querySelectorAll("tbody tr");
    expect(rows.length).toBe(3);
  });

  it("shows symbol column values", async () => {
    await setup(mockOrders);
    const symbols = fixture.nativeElement.querySelectorAll("td.symbol");
    expect(symbols[0].textContent.trim()).toBe("AAPL");
  });

  it("status badge has correct class", async () => {
    await setup(mockOrders);
    const filled = fixture.nativeElement.querySelector(".badge-filled");
    expect(filled).not.toBeNull();
  });

  // ── Filtering edge cases ─────────────────────────────────────────────────

  it("filters orders by symbol (case-insensitive)", async () => {
    await setup(mockOrders);
    fixture.componentInstance.filterSymbol.set("aapl"); // lowercase
    fixture.detectChanges();
    const rows = fixture.nativeElement.querySelectorAll("tbody tr");
    // AAPL orders: ids 1 and 3
    expect(rows.length).toBe(2);
  });

  it("filters orders by side 'buy'", async () => {
    await setup(mockOrders);
    fixture.componentInstance.filterSide.set("buy");
    fixture.detectChanges();
    const rows = fixture.nativeElement.querySelectorAll("tbody tr");
    expect(rows.length).toBe(1);
  });

  it("filters orders by side 'sell'", async () => {
    await setup(mockOrders);
    fixture.componentInstance.filterSide.set("sell");
    fixture.detectChanges();
    const rows = fixture.nativeElement.querySelectorAll("tbody tr");
    expect(rows.length).toBe(2);
  });

  it("returns all orders when filters are cleared", async () => {
    await setup(mockOrders);
    fixture.componentInstance.filterSymbol.set("AAPL");
    fixture.detectChanges();
    fixture.componentInstance.filterSymbol.set("");
    fixture.detectChanges();
    const rows = fixture.nativeElement.querySelectorAll("tbody tr");
    expect(rows.length).toBe(3);
  });

  it("shows 'No trades found' when filter matches nothing", async () => {
    await setup(mockOrders);
    fixture.componentInstance.filterSymbol.set("ZZZZ");
    fixture.detectChanges();
    const stateMsg = fixture.nativeElement.querySelector(".state-msg");
    expect(stateMsg).not.toBeNull();
    expect(stateMsg.textContent).toContain("No trades found");
  });

  it("combines symbol and side filters", async () => {
    await setup(mockOrders);
    fixture.componentInstance.filterSymbol.set("AAPL");
    fixture.componentInstance.filterSide.set("sell");
    fixture.detectChanges();
    const rows = fixture.nativeElement.querySelectorAll("tbody tr");
    // Only order id=3: AAPL + sell
    expect(rows.length).toBe(1);
  });

  // ── Error cases ──────────────────────────────────────────────────────────

  it("shows 'No trades found' when API fails", async () => {
    await setup([], true);
    const stateMsg = fixture.nativeElement.querySelector(".state-msg");
    expect(stateMsg).not.toBeNull();
  });

  it("does not throw when API fails", async () => {
    await setup([], true);
    expect(fixture.componentInstance).toBeTruthy();
  });

  // ── statusClass helper ───────────────────────────────────────────────────

  it("statusClass returns correct class for each status", async () => {
    await setup([]);
    const comp = fixture.componentInstance;
    expect(comp.statusClass("filled")).toBe("badge-filled");
    expect(comp.statusClass("pending")).toBe("badge-pending");
    expect(comp.statusClass("submitted")).toBe("badge-submitted");
    expect(comp.statusClass("cancelled")).toBe("badge-cancelled");
    expect(comp.statusClass("rejected")).toBe("badge-rejected");
    expect(comp.statusClass("unknown")).toBe("badge-pending");
  });
});
