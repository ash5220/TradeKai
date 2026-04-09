import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { RouterTestingModule } from "@angular/router/testing";
import { TradingComponent } from "./trading.component";
import { ApiService } from "../../core/api/api.service";
import { WebSocketService } from "../../core/websocket/websocket.service";
import { signal } from "@angular/core";
import { of } from "rxjs";
import { By } from "@angular/platform-browser";
import { StrategyListComponent } from "./components/strategy-list.component";
import { OrderFormComponent } from "./components/order-form.component";
import { TradeLogComponent } from "./components/trade-log.component";

describe("TradingComponent", () => {
  let fixture: ComponentFixture<TradingComponent>;

  beforeEach(async () => {
    const apiSpy = jasmine.createSpyObj<ApiService>("ApiService", [
      "getStrategies",
      "placeOrder",
      "startStrategy",
      "stopStrategy",
    ]);
    apiSpy.getStrategies.and.returnValue(of([]));
    apiSpy.placeOrder.and.returnValue(of({} as any));

    const wsSpy = jasmine.createSpyObj<WebSocketService>(
      "WebSocketService",
      ["connect", "disconnect"],
      {
        connectionState: signal("connected" as const),
        ticks: signal(new Map()),
        latestOrders: signal([]),
        isConnected: signal(true),
      },
    );

    await TestBed.configureTestingModule({
      imports: [TradingComponent, RouterTestingModule],
      providers: [
        { provide: ApiService, useValue: apiSpy },
        { provide: WebSocketService, useValue: wsSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(TradingComponent);
    fixture.detectChanges();
  });

  it("renders the strategy-list component", () => {
    expect(fixture.debugElement.query(By.directive(StrategyListComponent))).not.toBeNull();
  });

  it("renders the order-form component", () => {
    expect(fixture.debugElement.query(By.directive(OrderFormComponent))).not.toBeNull();
  });

  it("renders the trade-log component", () => {
    expect(fixture.debugElement.query(By.directive(TradeLogComponent))).not.toBeNull();
  });

  it("has correct grid layout with sidebar and main areas", () => {
    const layout = fixture.nativeElement.querySelector(".trading-layout");
    const sidebar = fixture.nativeElement.querySelector(".trading-sidebar");
    const main = fixture.nativeElement.querySelector(".trading-main");
    expect(layout).not.toBeNull();
    expect(sidebar).not.toBeNull();
    expect(main).not.toBeNull();
  });
});
