import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { TradeLogComponent } from "./trade-log.component";
import { WebSocketService } from "../../../core/websocket/websocket.service";
import { signal, WritableSignal } from "@angular/core";
import type { Order } from "../../../shared/models/order.model";

const mockOrders: Order[] = [
  {
    id: "1",
    symbol: "AAPL",
    side: "buy",
    type: "market",
    qty: 10,
    status: "filled",
    created_at: "2024-01-01T10:00:00Z",
    filled_avg_price: 150,
  } as Order,
  {
    id: "2",
    symbol: "TSLA",
    side: "sell",
    type: "limit",
    qty: 5,
    status: "pending",
    created_at: "2024-01-01T10:01:00Z",
    filled_avg_price: null,
  } as unknown as Order,
];

describe("TradeLogComponent", () => {
  let fixture: ComponentFixture<TradeLogComponent>;
  let ordersSignal: WritableSignal<Order[]>;

  async function setup(orders: Order[]): Promise<void> {
    ordersSignal = signal<Order[]>(orders);

    const wsSpy = jasmine.createSpyObj<WebSocketService>(
      "WebSocketService",
      ["connect", "disconnect"],
      {
        latestOrders: ordersSignal.asReadonly(),
        ticks: signal(new Map()),
        connectionState: signal("connected" as const),
        isConnected: signal(true),
      },
    );

    await TestBed.configureTestingModule({
      imports: [TradeLogComponent],
      providers: [{ provide: WebSocketService, useValue: wsSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(TradeLogComponent);
    fixture.detectChanges();
  }

  // ── Happy paths ──────────────────────────────────────────────────────────

  it("renders one row per order", async () => {
    await setup(mockOrders);
    const rows = fixture.nativeElement.querySelectorAll("tbody tr");
    expect(rows.length).toBe(2);
  });

  it("shows symbol with .symbol class", async () => {
    await setup(mockOrders);
    const symbols = fixture.nativeElement.querySelectorAll("td.symbol");
    expect(symbols[0].textContent.trim()).toBe("AAPL");
  });

  it("applies 'side-buy' class for buy orders", async () => {
    await setup(mockOrders);
    const buyCells = fixture.nativeElement.querySelectorAll("td.side-buy");
    expect(buyCells.length).toBeGreaterThan(0);
  });

  it("applies 'side-sell' class for sell orders", async () => {
    await setup(mockOrders);
    const sellCells = fixture.nativeElement.querySelectorAll("td.side-sell");
    expect(sellCells.length).toBeGreaterThan(0);
  });

  it("applies 'badge-filled' class for filled status", async () => {
    await setup(mockOrders);
    const filledBadges =
      fixture.nativeElement.querySelectorAll(".badge-filled");
    expect(filledBadges.length).toBeGreaterThan(0);
  });

  it("applies 'badge-pending' class for pending status", async () => {
    await setup(mockOrders);
    const pendingBadges =
      fixture.nativeElement.querySelectorAll(".badge-pending");
    expect(pendingBadges.length).toBeGreaterThan(0);
  });

  // ── Edge cases ───────────────────────────────────────────────────────────

  it("shows empty waiting message when there are no orders", async () => {
    await setup([]);
    const empty = fixture.nativeElement.querySelector(".empty");
    expect(empty).not.toBeNull();
    const rows = fixture.nativeElement.querySelectorAll("tbody tr");
    expect(rows.length).toBe(0);
  });

  it("applies all known status classes correctly", async () => {
    const allStatuses: Order[] = [
      "filled",
      "pending",
      "submitted",
      "cancelled",
      "rejected",
    ].map(
      (status, i) =>
        ({
          id: String(i),
          symbol: "AAPL",
          side: "buy",
          type: "market",
          qty: 1,
          status,
          created_at: new Date().toISOString(),
        }) as unknown as Order,
    );

    await setup(allStatuses);
    const component = fixture.componentInstance;
    expect(component.statusClass("filled")).toBe("badge-filled");
    expect(component.statusClass("pending")).toBe("badge-pending");
    expect(component.statusClass("submitted")).toBe("badge-submitted");
    expect(component.statusClass("cancelled")).toBe("badge-cancelled");
    expect(component.statusClass("rejected")).toBe("badge-rejected");
  });

  it("falls back to badge-pending for unknown status", async () => {
    await setup([]);
    expect(fixture.componentInstance.statusClass("unknown_status")).toBe(
      "badge-pending",
    );
  });
});
