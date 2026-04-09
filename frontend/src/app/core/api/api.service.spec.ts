import { TestBed } from "@angular/core/testing";
import {
  HttpClientTestingModule,
  HttpTestingController,
} from "@angular/common/http/testing";
import { ApiService, PlaceOrderRequest } from "./api.service";

describe("ApiService", () => {
  let service: ApiService;
  let httpMock: HttpTestingController;
  const base = "/api/v1";

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [ApiService],
    });
    service = TestBed.inject(ApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  // ── Orders ────────────────────────────────────────────────────────────────

  it("getOrders() sends GET with default limit=50 and offset=0", () => {
    service.getOrders().subscribe();
    const req = httpMock.expectOne(
      (r) =>
        r.url === `${base}/orders` &&
        r.params.get("limit") === "50" &&
        r.params.get("offset") === "0",
    );
    expect(req.request.method).toBe("GET");
    req.flush([]);
  });

  it("getOrders() accepts custom limit and offset", () => {
    service.getOrders(200, 10).subscribe();
    const req = httpMock.expectOne(
      (r) =>
        r.url === `${base}/orders` &&
        r.params.get("limit") === "200" &&
        r.params.get("offset") === "10",
    );
    expect(req.request.method).toBe("GET");
    req.flush([]);
  });

  it("placeOrder() sends POST to /orders with body", () => {
    const orderReq: PlaceOrderRequest = {
      symbol: "AAPL",
      side: "buy",
      type: "market",
      qty: 1,
    };
    service.placeOrder(orderReq).subscribe();
    const req = httpMock.expectOne(`${base}/orders`);
    expect(req.request.method).toBe("POST");
    expect(req.request.body).toEqual(orderReq);
    req.flush({ id: "123" });
  });

  it("placeOrder() includes limit_price when type is limit", () => {
    const orderReq: PlaceOrderRequest = {
      symbol: "TSLA",
      side: "buy",
      type: "limit",
      qty: 5,
      limit_price: 250.5,
    };
    service.placeOrder(orderReq).subscribe();
    const req = httpMock.expectOne(`${base}/orders`);
    expect(req.request.body.limit_price).toBe(250.5);
    req.flush({ id: "456" });
  });

  it("cancelOrder() sends DELETE to /orders/:id", () => {
    service.cancelOrder("abc-123").subscribe();
    const req = httpMock.expectOne(`${base}/orders/abc-123`);
    expect(req.request.method).toBe("DELETE");
    req.flush(null);
  });

  // ── Portfolio ─────────────────────────────────────────────────────────────

  it("getPositions() sends GET to /portfolio/positions", () => {
    service.getPositions().subscribe();
    const req = httpMock.expectOne(`${base}/portfolio/positions`);
    expect(req.request.method).toBe("GET");
    req.flush([]);
  });

  it("getPnL() sends GET to /portfolio/pnl", () => {
    service.getPnL().subscribe();
    const req = httpMock.expectOne(`${base}/portfolio/pnl`);
    expect(req.request.method).toBe("GET");
    req.flush({ daily_realized_pnl: 100 });
  });

  // ── Market ────────────────────────────────────────────────────────────────

  it("getCandles() sends GET with symbol, interval and limit", () => {
    service.getCandles("AAPL", "1m").subscribe();
    const req = httpMock.expectOne(
      (r) =>
        r.url === `${base}/market/candles/AAPL` &&
        r.params.get("interval") === "1m" &&
        r.params.get("limit") === "500",
    );
    expect(req.request.method).toBe("GET");
    req.flush([]);
  });

  it("getCandles() appends optional from/to params when provided", () => {
    service
      .getCandles("MSFT", "5m", "2024-01-01", "2024-01-02", 100)
      .subscribe();
    const req = httpMock.expectOne(
      (r) =>
        r.url === `${base}/market/candles/MSFT` &&
        r.params.get("from") === "2024-01-01" &&
        r.params.get("to") === "2024-01-02" &&
        r.params.get("limit") === "100",
    );
    expect(req.request.method).toBe("GET");
    expect(req.request.params.get("interval")).toBe("5m");
    req.flush([]);
  });

  // ── Strategies ────────────────────────────────────────────────────────────

  it("getStrategies() sends GET to /strategies", () => {
    service.getStrategies().subscribe();
    const req = httpMock.expectOne(`${base}/strategies`);
    expect(req.request.method).toBe("GET");
    req.flush([]);
  });

  it("startStrategy() sends POST to /strategies/:name/start with symbols", () => {
    service.startStrategy("RSI", ["AAPL", "TSLA"]).subscribe();
    const req = httpMock.expectOne(`${base}/strategies/RSI/start`);
    expect(req.request.method).toBe("POST");
    expect(req.request.body).toEqual({ symbols: ["AAPL", "TSLA"] });
    req.flush(null);
  });

  it("stopStrategy() sends POST to /strategies/:name/stop", () => {
    service.stopStrategy("MACD").subscribe();
    const req = httpMock.expectOne(`${base}/strategies/MACD/stop`);
    expect(req.request.method).toBe("POST");
    req.flush(null);
  });

  // ── Error propagation ─────────────────────────────────────────────────────

  it("getOrders() propagates HTTP 500 to subscriber", (done) => {
    service.getOrders().subscribe({
      error: (err) => {
        expect(err.status).toBe(500);
        done();
      },
    });
    const req = httpMock.expectOne((r) => r.url === `${base}/orders`);
    req.flush(
      { error: "Server Error" },
      { status: 500, statusText: "Internal Server Error" },
    );
  });

  it("placeOrder() propagates HTTP 400 to subscriber", (done) => {
    const orderReq: PlaceOrderRequest = {
      symbol: "AAPL",
      side: "buy",
      type: "market",
      qty: 1,
    };
    service.placeOrder(orderReq).subscribe({
      error: (err) => {
        expect(err.status).toBe(400);
        done();
      },
    });
    const req = httpMock.expectOne(`${base}/orders`);
    req.flush(
      { error: "Bad Request" },
      { status: 400, statusText: "Bad Request" },
    );
  });
});
