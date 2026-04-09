import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { RouterTestingModule } from "@angular/router/testing";
import { DashboardComponent } from "./dashboard.component";
import { WebSocketService } from "../../core/websocket/websocket.service";
import { ApiService } from "../../core/api/api.service";
import { signal } from "@angular/core";
import { of } from "rxjs";
import { By } from "@angular/platform-browser";
import { PriceChartComponent } from "./components/price-chart.component";
import { PositionTableComponent } from "./components/position-table.component";
import { PnlCardComponent } from "./components/pnl-card.component";
import { SystemHealthComponent } from "./components/system-health.component";

describe("DashboardComponent", () => {
  let fixture: ComponentFixture<DashboardComponent>;
  let wsSpy: jasmine.SpyObj<WebSocketService>;
  let apiSpy: jasmine.SpyObj<ApiService>;

  beforeEach(async () => {
    wsSpy = jasmine.createSpyObj<WebSocketService>(
      "WebSocketService",
      ["connect", "subscribe", "disconnect", "unsubscribe"],
      {
        connectionState: signal("disconnected" as const),
        ticks: signal(new Map()),
        latestOrders: signal([]),
        isConnected: signal(false),
      },
    );

    apiSpy = jasmine.createSpyObj<ApiService>("ApiService", [
      "getPositions",
      "getPnL",
      "getCandles",
      "getStrategies",
      "placeOrder",
    ]);
    apiSpy.getPositions.and.returnValue(of([]));
    apiSpy.getPnL.and.returnValue(of({ daily_realized_pnl: 0 }));
    apiSpy.getCandles.and.returnValue(of([]));
    apiSpy.getStrategies.and.returnValue(of([]));

    await TestBed.configureTestingModule({
      imports: [DashboardComponent, RouterTestingModule],
      providers: [
        { provide: WebSocketService, useValue: wsSpy },
        { provide: ApiService, useValue: apiSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardComponent);
    fixture.detectChanges();
  });

  it("renders the pnl-card component", () => {
    expect(
      fixture.debugElement.query(By.directive(PnlCardComponent)),
    ).not.toBeNull();
  });

  it("renders the system-health component", () => {
    expect(
      fixture.debugElement.query(By.directive(SystemHealthComponent)),
    ).not.toBeNull();
  });

  it("renders the price-chart component", () => {
    expect(
      fixture.debugElement.query(By.directive(PriceChartComponent)),
    ).not.toBeNull();
  });

  it("renders the position-table component", () => {
    expect(
      fixture.debugElement.query(By.directive(PositionTableComponent)),
    ).not.toBeNull();
  });

  it("calls ws.connect on ngOnInit", () => {
    expect(wsSpy.connect).toHaveBeenCalledTimes(1);
  });

  it("subscribes to default symbols on ngOnInit", () => {
    const expectedSymbols = ["AAPL", "TSLA", "MSFT", "GOOGL"];
    for (const sym of expectedSymbols) {
      expect(wsSpy.subscribe).toHaveBeenCalledWith(`ticks:${sym}`);
    }
  });
});
