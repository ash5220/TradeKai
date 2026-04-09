import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { SystemHealthComponent } from "./system-health.component";
import {
  WebSocketService,
  ConnectionState,
} from "../../../core/websocket/websocket.service";
import { signal, WritableSignal } from "@angular/core";

describe("SystemHealthComponent", () => {
  let fixture: ComponentFixture<SystemHealthComponent>;
  let connectionStateSignal: WritableSignal<ConnectionState>;

  async function createFixtureWith(state: ConnectionState): Promise<void> {
    connectionStateSignal = signal<ConnectionState>(state);

    const wsSpy = jasmine.createSpyObj<WebSocketService>(
      "WebSocketService",
      ["connect", "disconnect"],
      {
        connectionState: connectionStateSignal.asReadonly(),
        ticks: signal(new Map()),
        latestOrders: signal([]),
        isConnected: signal(state === "connected"),
      },
    );

    await TestBed.configureTestingModule({
      imports: [SystemHealthComponent],
      providers: [{ provide: WebSocketService, useValue: wsSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(SystemHealthComponent);
    fixture.detectChanges();
  }

  // ── Happy paths ──────────────────────────────────────────────────────────

  it("shows 'dot-connected' and 'Live' when connected", async () => {
    await createFixtureWith("connected");
    const dot: HTMLElement = fixture.nativeElement.querySelector(".status-dot");
    const text: HTMLElement =
      fixture.nativeElement.querySelector(".status-text");
    expect(dot.classList).toContain("dot-connected");
    expect(text.textContent?.trim()).toBe("Live");
  });

  // ── Edge cases ───────────────────────────────────────────────────────────

  it("shows 'dot-connecting' and 'Connecting…' when connecting", async () => {
    await createFixtureWith("connecting");
    const dot: HTMLElement = fixture.nativeElement.querySelector(".status-dot");
    const text: HTMLElement =
      fixture.nativeElement.querySelector(".status-text");
    expect(dot.classList).toContain("dot-connecting");
    expect(text.textContent?.trim()).toBe("Connecting…");
  });

  it("shows 'dot-disconnected' and 'Disconnected' when disconnected", async () => {
    await createFixtureWith("disconnected");
    const dot: HTMLElement = fixture.nativeElement.querySelector(".status-dot");
    const text: HTMLElement =
      fixture.nativeElement.querySelector(".status-text");
    expect(dot.classList).toContain("dot-disconnected");
    expect(text.textContent?.trim()).toBe("Disconnected");
  });

  it("shows 'dot-error' and 'Error' when in error state", async () => {
    await createFixtureWith("error");
    const dot: HTMLElement = fixture.nativeElement.querySelector(".status-dot");
    const text: HTMLElement =
      fixture.nativeElement.querySelector(".status-text");
    expect(dot.classList).toContain("dot-error");
    expect(text.textContent?.trim()).toBe("Error");
  });

  it("shows reconnect hint only when state is error", async () => {
    await createFixtureWith("error");
    const hint: HTMLElement | null =
      fixture.nativeElement.querySelector(".status-hint");
    expect(hint).not.toBeNull();
  });

  it("does not show reconnect hint when connected", async () => {
    await createFixtureWith("connected");
    const hint: HTMLElement | null =
      fixture.nativeElement.querySelector(".status-hint");
    expect(hint).toBeNull();
  });

  it("does not show reconnect hint when disconnected", async () => {
    await createFixtureWith("disconnected");
    const hint: HTMLElement | null =
      fixture.nativeElement.querySelector(".status-hint");
    expect(hint).toBeNull();
  });
});
