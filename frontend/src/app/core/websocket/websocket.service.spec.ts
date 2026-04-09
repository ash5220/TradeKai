import { TestBed, fakeAsync, tick } from "@angular/core/testing";
import { computed, signal } from "@angular/core";
import { WebSocketService } from "./websocket.service";
import { AuthService } from "../auth/auth.service";

// ── FakeWebSocket ──────────────────────────────────────────────────────────

class FakeWebSocket {
  static OPEN = 1;
  static CONNECTING = 0;
  static CLOSED = 3;

  readyState = FakeWebSocket.CONNECTING;
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onerror: (() => void) | null = null;
  onclose: (() => void) | null = null;

  send = jasmine.createSpy("send");
  close = jasmine.createSpy("close").and.callFake(() => {
    this.readyState = FakeWebSocket.CLOSED;
    this.onclose?.();
  });

  triggerOpen(): void {
    this.readyState = FakeWebSocket.OPEN;
    this.onopen?.();
  }

  triggerMessage(data: unknown): void {
    this.onmessage?.({ data: JSON.stringify(data) });
  }

  triggerError(): void {
    this.onerror?.();
  }

  triggerClose(): void {
    this.readyState = FakeWebSocket.CLOSED;
    this.onclose?.();
  }
}

// ──────────────────────────────────────────────────────────────────────────

describe("WebSocketService", () => {
  let service: WebSocketService;
  let authServiceMock: jasmine.SpyObj<AuthService>;
  let fakeWs: FakeWebSocket;
  let wsSpy: jasmine.Spy;

  beforeEach(() => {
    fakeWs = new FakeWebSocket();

    // Replace the global WebSocket constructor so new WebSocket(...) returns our fake
    const FakeWebSocketConstructor = function (this: FakeWebSocket) {
      return fakeWs;
    } as unknown as typeof WebSocket;
    (FakeWebSocketConstructor as any).OPEN = 1;
    (FakeWebSocketConstructor as any).CONNECTING = 0;
    (FakeWebSocketConstructor as any).CLOSED = 3;
    (FakeWebSocketConstructor as any).CLOSING = 2;

    wsSpy = spyOn(window as any, "WebSocket").and.returnValue(fakeWs);

    authServiceMock = jasmine.createSpyObj<AuthService>(
      "AuthService",
      ["logout"],
      {
        accessToken: computed(() => "test-token"),
        isAuthenticated: computed(() => true),
      },
    );

    TestBed.configureTestingModule({
      providers: [
        WebSocketService,
        { provide: AuthService, useValue: authServiceMock },
      ],
    });

    service = TestBed.inject(WebSocketService);
  });

  afterEach(() => {
    service.ngOnDestroy();
  });

  // ── connect ───────────────────────────────────────────────────────────────

  it("connect() creates WebSocket with token appended to URL", () => {
    service.connect();
    expect(wsSpy).toHaveBeenCalledWith(
      jasmine.stringContaining("token=test-token"),
    );
  });

  it("connect() sets connectionState to 'connecting'", () => {
    service.connect();
    expect(service.connectionState()).toBe("connecting");
  });

  it("connect() does nothing when no access token", () => {
    // Override the mock auth to return null token for this call
    Object.defineProperty(authServiceMock, "accessToken", {
      value: computed(() => null),
      configurable: true,
    });
    // Reset the spy call count
    wsSpy.calls.reset();
    service.connect();
    // WebSocket constructor should not have been called
    expect(wsSpy).not.toHaveBeenCalled();
    // Restore token for other tests
    Object.defineProperty(authServiceMock, "accessToken", {
      value: computed(() => "test-token"),
      configurable: true,
    });
  });

  it("does not open second WebSocket when OPEN connection already exists", () => {
    service.connect();
    fakeWs.triggerOpen();
    const callsBefore = wsSpy.calls.count();
    service.connect();
    expect(wsSpy.calls.count()).toBe(callsBefore);
  });

  // ── onopen ────────────────────────────────────────────────────────────────

  it("connectionState becomes 'connected' on open", () => {
    service.connect();
    fakeWs.triggerOpen();
    expect(service.connectionState()).toBe("connected");
  });

  // ── tick message ──────────────────────────────────────────────────────────

  it("updates ticks signal when a tick message is received", () => {
    service.connect();
    fakeWs.triggerOpen();
    const tick = {
      symbol: "AAPL",
      price: 150,
      volume: 1000,
      timestamp: "2024-01-01T00:00:00Z",
    };
    fakeWs.triggerMessage({ type: "tick", payload: tick });
    expect(service.ticks().get("AAPL")).toEqual(tick as any);
  });

  it("handles multiple tick messages for different symbols", () => {
    service.connect();
    fakeWs.triggerOpen();
    fakeWs.triggerMessage({
      type: "tick",
      payload: { symbol: "AAPL", price: 150 },
    });
    fakeWs.triggerMessage({
      type: "tick",
      payload: { symbol: "TSLA", price: 200 },
    });
    expect(service.ticks().size).toBe(2);
  });

  it("overwrites existing tick for the same symbol", () => {
    service.connect();
    fakeWs.triggerOpen();
    fakeWs.triggerMessage({
      type: "tick",
      payload: { symbol: "AAPL", price: 150 },
    });
    fakeWs.triggerMessage({
      type: "tick",
      payload: { symbol: "AAPL", price: 155 },
    });
    expect(service.ticks().get("AAPL")!.price).toBe(155);
  });

  // ── order_update message ──────────────────────────────────────────────────

  it("prepends order to latestOrders on order_update", () => {
    service.connect();
    fakeWs.triggerOpen();
    const order = {
      id: "o1",
      symbol: "AAPL",
      status: "filled",
      qty: 1,
      side: "buy",
    };
    fakeWs.triggerMessage({ type: "order_update", payload: order });
    expect(service.latestOrders()[0]).toEqual(order as any);
  });

  it("caps latestOrders at 100 entries", () => {
    service.connect();
    fakeWs.triggerOpen();
    for (let i = 0; i < 105; i++) {
      fakeWs.triggerMessage({ type: "order_update", payload: { id: `o${i}` } });
    }
    expect(service.latestOrders().length).toBe(100);
  });

  // ── onerror ───────────────────────────────────────────────────────────────

  it("sets connectionState to 'error' on onerror", () => {
    service.connect();
    fakeWs.triggerError();
    expect(service.connectionState()).toBe("error");
  });

  // ── onclose + reconnect ───────────────────────────────────────────────────

  it("sets connectionState to 'disconnected' on close", fakeAsync(() => {
    service.connect();
    fakeWs.triggerOpen();
    fakeWs.triggerClose();
    expect(service.connectionState()).toBe("disconnected");
    tick(30_000); // consume any pending timers
  }));

  it("schedules reconnect after close (non-destroyed)", fakeAsync(() => {
    service.connect();
    fakeWs.triggerOpen();
    fakeWs.triggerClose();
    // After 1s the reconnect fires
    tick(1000);
    expect(wsSpy.calls.count()).toBe(2);
    tick(30_000);
  }));

  // ── disconnect() ──────────────────────────────────────────────────────────

  it("disconnect() prevents reconnect after close", fakeAsync(() => {
    service.connect();
    fakeWs.triggerOpen();
    service.disconnect();
    const callCount = wsSpy.calls.count();
    tick(30_000);
    expect(wsSpy.calls.count()).toBe(callCount);
  }));

  // ── malformed JSON ────────────────────────────────────────────────────────

  it("does not crash when onmessage receives invalid JSON", () => {
    service.connect();
    fakeWs.triggerOpen();
    expect(() => {
      service["ws"]!.onmessage!({ data: "not-valid-json" } as any);
    }).not.toThrow();
    expect(service.ticks().size).toBe(0);
  });

  // ── subscribe/unsubscribe ─────────────────────────────────────────────────

  it("subscribe() sends subscribe action when connected", () => {
    service.connect();
    fakeWs.triggerOpen();
    service.subscribe("ticks:AAPL");
    expect(fakeWs.send).toHaveBeenCalledWith(
      JSON.stringify({ action: "subscribe", room: "ticks:AAPL" }),
    );
  });

  it("unsubscribe() sends unsubscribe action when connected", () => {
    service.connect();
    fakeWs.triggerOpen();
    service.unsubscribe("ticks:AAPL");
    expect(fakeWs.send).toHaveBeenCalledWith(
      JSON.stringify({ action: "unsubscribe", room: "ticks:AAPL" }),
    );
  });

  it("subscribe() does nothing when socket is not OPEN", () => {
    service.connect();
    // socket is still CONNECTING, not OPEN
    service.subscribe("ticks:AAPL");
    expect(fakeWs.send).not.toHaveBeenCalled();
  });
});
