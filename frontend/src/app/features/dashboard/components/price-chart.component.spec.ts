import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { PriceChartComponent } from "./price-chart.component";
import { ApiService } from "../../../core/api/api.service";
import { WebSocketService } from "../../../core/websocket/websocket.service";
import { signal } from "@angular/core";
import { of, throwError } from "rxjs";
import type { Candle, Tick } from "../../../shared/models/market.model";

describe("PriceChartComponent", () => {
  let fixture: ComponentFixture<PriceChartComponent>;
  let apiSpy: jasmine.SpyObj<ApiService>;
  let wsSpy: jasmine.SpyObj<WebSocketService>;

  const sampleCandles: Candle[] = [
    { symbol: "AAPL", interval: "1m", ts: "2024-01-01T09:30:00Z", open: 100, high: 105, low: 99, close: 103, volume: 1000 },
    { symbol: "AAPL", interval: "1m", ts: "2024-01-01T09:31:00Z", open: 103, high: 106, low: 102, close: 105, volume: 1200 },
  ];

  beforeEach(async () => {
    apiSpy = jasmine.createSpyObj<ApiService>("ApiService", ["getCandles"]);
    apiSpy.getCandles.and.returnValue(of(sampleCandles));

    wsSpy = jasmine.createSpyObj<WebSocketService>(
      "WebSocketService",
      ["subscribe", "unsubscribe"],
      {
        ticks: signal(new Map<string, Tick>()),
        latestOrders: signal([]),
        connectionState: signal("connected" as const),
        isConnected: signal(true),
      },
    );

    await TestBed.configureTestingModule({
      imports: [PriceChartComponent],
      providers: [
        { provide: ApiService, useValue: apiSpy },
        { provide: WebSocketService, useValue: wsSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(PriceChartComponent);

    // Stub out chart initialisation so JSDOM (no canvas) doesn't throw
    spyOn(fixture.componentInstance as any, "initChart").and.stub();
  });

  // ── Happy paths ──────────────────────────────────────────────────────────

  it("calls apiService.getCandles for default symbol on init", () => {
    fixture.detectChanges();
    expect(apiSpy.getCandles).toHaveBeenCalledWith("AAPL", "1m");
  });

  it("component mounts without errors", () => {
    expect(() => fixture.detectChanges()).not.toThrow();
  });

  it("exposes the expected list of symbols", () => {
    expect(fixture.componentInstance.symbols).toContain("AAPL");
    expect(fixture.componentInstance.symbols).toContain("TSLA");
  });

  it("exposes the expected list of intervals", () => {
    expect(fixture.componentInstance.intervals).toContain("1m");
    expect(fixture.componentInstance.intervals).toContain("5m");
  });

  it("calls getCandles with new symbol when symbol changes", () => {
    fixture.detectChanges();
    apiSpy.getCandles.calls.reset();
    apiSpy.getCandles.and.returnValue(of([]));

    fixture.componentInstance.onSymbolChange("TSLA");
    expect(apiSpy.getCandles).toHaveBeenCalledWith("TSLA", jasmine.any(String));
  });

  it("calls getCandles when interval changes", () => {
    fixture.detectChanges();
    apiSpy.getCandles.calls.reset();
    apiSpy.getCandles.and.returnValue(of([]));

    fixture.componentInstance.onIntervalChange("5m");
    expect(apiSpy.getCandles).toHaveBeenCalledTimes(1);
  });

  it("unsubscribes from old symbol and subscribes to new on symbol change", () => {
    fixture.detectChanges();
    apiSpy.getCandles.and.returnValue(of([]));
    fixture.componentInstance.onSymbolChange("TSLA");
    expect(wsSpy.unsubscribe).toHaveBeenCalledWith("ticks:AAPL");
    expect(wsSpy.subscribe).toHaveBeenCalledWith("ticks:TSLA");
  });

  // ── Edge cases ───────────────────────────────────────────────────────────

  it("does not crash on empty candles array", () => {
    apiSpy.getCandles.and.returnValue(of([]));
    expect(() => fixture.detectChanges()).not.toThrow();
  });

  // ── Error cases ──────────────────────────────────────────────────────────

  it("does not throw when getCandles fails", () => {
    apiSpy.getCandles.and.returnValue(throwError(() => new Error("Network error")));
    expect(() => fixture.detectChanges()).not.toThrow();
  });
});


