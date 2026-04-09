import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { StrategyListComponent } from "./strategy-list.component";
import { ApiService, StrategyInfo } from "../../../core/api/api.service";
import { of, throwError } from "rxjs";

const mockStrategies: StrategyInfo[] = [
  { name: "RSI", symbol: "AAPL", active: false, running: false },
  { name: "MACD", symbol: "TSLA", active: true, running: true },
];

describe("StrategyListComponent", () => {
  let fixture: ComponentFixture<StrategyListComponent>;
  let apiSpy: jasmine.SpyObj<ApiService>;

  async function setup(strategies: StrategyInfo[]): Promise<void> {
    apiSpy = jasmine.createSpyObj<ApiService>("ApiService", [
      "getStrategies",
      "startStrategy",
      "stopStrategy",
    ]);
    apiSpy.getStrategies.and.returnValue(of(strategies));
    apiSpy.startStrategy.and.returnValue(of(undefined as any));
    apiSpy.stopStrategy.and.returnValue(of(undefined as any));

    await TestBed.configureTestingModule({
      imports: [StrategyListComponent],
      providers: [{ provide: ApiService, useValue: apiSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(StrategyListComponent);
    fixture.detectChanges();
  }

  // ── Happy paths ──────────────────────────────────────────────────────────

  it("renders each strategy item", async () => {
    await setup(mockStrategies);
    const items = fixture.nativeElement.querySelectorAll(".strategy-item");
    expect(items.length).toBe(2);
  });

  it("shows strategy name and symbol", async () => {
    await setup(mockStrategies);
    const names = fixture.nativeElement.querySelectorAll(".strategy-name");
    expect(names[0].textContent.trim()).toBe("RSI");
    expect(names[1].textContent.trim()).toBe("MACD");
  });

  it("shows 'Active' badge for active strategy", async () => {
    await setup(mockStrategies);
    const badges = fixture.nativeElement.querySelectorAll(".badge");
    // Second strategy is active
    expect(badges[1].textContent.trim()).toBe("Active");
    expect(badges[1].classList).toContain("badge-active");
  });

  it("shows 'Idle' badge for inactive strategy", async () => {
    await setup(mockStrategies);
    const badges = fixture.nativeElement.querySelectorAll(".badge");
    expect(badges[0].textContent.trim()).toBe("Idle");
    expect(badges[0].classList).toContain("badge-idle");
  });

  it("calls startStrategy when Start button is clicked", async () => {
    await setup(mockStrategies);
    const startBtns = fixture.nativeElement.querySelectorAll(".btn-success-sm");
    startBtns[0].click();
    expect(apiSpy.startStrategy).toHaveBeenCalledWith("RSI", ["AAPL"]);
  });

  it("updates strategy badge to Active after start", async () => {
    await setup(mockStrategies);
    const startBtns = fixture.nativeElement.querySelectorAll(".btn-success-sm");
    startBtns[0].click();
    fixture.detectChanges();
    const badge = fixture.nativeElement.querySelectorAll(".badge")[0];
    expect(badge.textContent.trim()).toBe("Active");
  });

  it("calls stopStrategy when Stop button is clicked", async () => {
    await setup(mockStrategies);
    const stopBtns = fixture.nativeElement.querySelectorAll(".btn-danger-sm");
    stopBtns[0].click();
    expect(apiSpy.stopStrategy).toHaveBeenCalledWith("MACD");
  });

  it("updates strategy badge to Idle after stop", async () => {
    await setup(mockStrategies);
    const stopBtns = fixture.nativeElement.querySelectorAll(".btn-danger-sm");
    stopBtns[0].click();
    fixture.detectChanges();
    const badge = fixture.nativeElement.querySelectorAll(".badge")[1];
    expect(badge.textContent.trim()).toBe("Idle");
  });

  // ── Edge cases ───────────────────────────────────────────────────────────

  it("shows empty message when no strategies exist", async () => {
    await setup([]);
    const empty = fixture.nativeElement.querySelector(".empty");
    expect(empty).not.toBeNull();
    expect(
      fixture.nativeElement.querySelectorAll(".strategy-item").length,
    ).toBe(0);
  });

  // ── Error cases ──────────────────────────────────────────────────────────

  it("strategy state does not change when startStrategy fails", async () => {
    apiSpy = jasmine.createSpyObj<ApiService>("ApiService", [
      "getStrategies",
      "startStrategy",
      "stopStrategy",
    ]);
    apiSpy.getStrategies.and.returnValue(of(mockStrategies as any));
    apiSpy.startStrategy.and.returnValue(throwError(() => new Error("fail")));
    apiSpy.stopStrategy.and.returnValue(of(undefined as any));

    await TestBed.resetTestingModule();
    await TestBed.configureTestingModule({
      imports: [StrategyListComponent],
      providers: [{ provide: ApiService, useValue: apiSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(StrategyListComponent);
    fixture.detectChanges();

    const startBtns = fixture.nativeElement.querySelectorAll(".btn-success-sm");
    startBtns[0].click();
    fixture.detectChanges();

    // RSI should remain Idle
    const badge = fixture.nativeElement.querySelectorAll(".badge")[0];
    expect(badge.textContent.trim()).toBe("Idle");
  });
});
