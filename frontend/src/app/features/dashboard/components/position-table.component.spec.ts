import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { PositionTableComponent } from "./position-table.component";
import { ApiService } from "../../../core/api/api.service";
import { WebSocketService } from "../../../core/websocket/websocket.service";
import { signal, WritableSignal } from "@angular/core";
import { of, throwError } from "rxjs";
import type { Position } from "../../../shared/models/position.model";
import type { Tick } from "../../../shared/models/market.model";

const mockPositions: Position[] = [
  { user_id: "u1", symbol: "AAPL", qty: 10, avg_price: 150, realized_pnl: 0, updated_at: "2024-01-01", unrealized_pnl: 50, current_price: 155 },
  { user_id: "u1", symbol: "TSLA", qty: 5, avg_price: 200, realized_pnl: 0, updated_at: "2024-01-01", unrealized_pnl: -25, current_price: 195 },
];

describe("PositionTableComponent", () => {
  let fixture: ComponentFixture<PositionTableComponent>;
  let apiSpy: jasmine.SpyObj<ApiService>;
  let ticksSignal: WritableSignal<Map<string, Tick>>;

  async function setup(positions: Position[], shouldFail = false): Promise<void> {
    apiSpy = jasmine.createSpyObj<ApiService>("ApiService", [
      "getPositions",
      "placeOrder",
    ]);

    if (shouldFail) {
      apiSpy.getPositions.and.returnValue(throwError(() => new Error("fail")));
    } else {
      apiSpy.getPositions.and.returnValue(of(positions));
    }
    apiSpy.placeOrder.and.returnValue(of({} as any));

    ticksSignal = signal<Map<string, Tick>>(
      new Map([["AAPL", { symbol: "AAPL", price: 160, volume: 500, timestamp: "" }]]),
    );

    const wsSpy = jasmine.createSpyObj<WebSocketService>(
      "WebSocketService",
      ["connect", "disconnect"],
      {
        ticks: ticksSignal.asReadonly(),
        latestOrders: signal([]),
        connectionState: signal("connected" as const),
        isConnected: signal(true),
      },
    );

    await TestBed.configureTestingModule({
      imports: [PositionTableComponent],
      providers: [
        { provide: ApiService, useValue: apiSpy },
        { provide: WebSocketService, useValue: wsSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(PositionTableComponent);
    fixture.detectChanges();
  }

  // ── Happy paths ──────────────────────────────────────────────────────────

  it("renders a row per position", async () => {
    await setup(mockPositions);
    const rows = fixture.nativeElement.querySelectorAll("tbody tr");
    expect(rows.length).toBe(2);
  });

  it("shows symbol with .symbol class", async () => {
    await setup(mockPositions);
    const symbols = fixture.nativeElement.querySelectorAll("td.symbol");
    expect(symbols[0].textContent.trim()).toBe("AAPL");
  });

  it("shows positive PnL with pnl-positive class", async () => {
    await setup(mockPositions);
    const pnlCells = fixture.nativeElement.querySelectorAll("td.pnl-positive");
    expect(pnlCells.length).toBeGreaterThan(0);
  });

  it("shows negative PnL with pnl-negative class", async () => {
    await setup(mockPositions);
    const pnlCells = fixture.nativeElement.querySelectorAll("td.pnl-negative");
    expect(pnlCells.length).toBeGreaterThan(0);
  });

  it("calls placeOrder with sell market order when Close is clicked", async () => {
    await setup(mockPositions);
    const closeBtns: NodeListOf<HTMLButtonElement> =
      fixture.nativeElement.querySelectorAll(".btn-danger-sm");
    closeBtns[0].click();
    expect(apiSpy.placeOrder).toHaveBeenCalledWith(
      jasmine.objectContaining({ symbol: "AAPL", side: "sell", type: "market" }),
    );
  });

  it("refreshes positions after closing a position", async () => {
    await setup(mockPositions);
    apiSpy.getPositions.calls.reset();
    apiSpy.getPositions.and.returnValue(of([]));

    const closeBtns: NodeListOf<HTMLButtonElement> =
      fixture.nativeElement.querySelectorAll(".btn-danger-sm");
    closeBtns[0].click();
    fixture.detectChanges();

    expect(apiSpy.getPositions).toHaveBeenCalledTimes(1);
  });

  // ── Edge cases ───────────────────────────────────────────────────────────

  it("shows empty state when there are no positions", async () => {
    await setup([]);
    const empty = fixture.nativeElement.querySelector(".empty");
    expect(empty).not.toBeNull();
    const rows = fixture.nativeElement.querySelectorAll("tbody tr");
    expect(rows.length).toBe(0);
  });

  it("displays current price from ticks signal", async () => {
    await setup([mockPositions[0]]);
    // AAPL tick price is 160
    const cells: NodeListOf<HTMLTableCellElement> =
      fixture.nativeElement.querySelectorAll("tbody td");
    // Find the cell with 160.00
    const priceCell = Array.from(cells).find((c) =>
      c.textContent?.includes("160"),
    );
    expect(priceCell).not.toBeUndefined();
  });

  // ── Error cases ──────────────────────────────────────────────────────────

  it("does not throw when getPositions fails", async () => {
    await setup([], true);
    // Component should remain mounted even when getPositions errors
    expect(fixture.componentInstance).toBeTruthy();
  });
});
