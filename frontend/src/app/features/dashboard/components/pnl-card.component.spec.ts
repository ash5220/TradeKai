import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { PnlCardComponent } from "./pnl-card.component";
import { ApiService } from "../../../core/api/api.service";
import { of, throwError } from "rxjs";

describe("PnlCardComponent", () => {
  let fixture: ComponentFixture<PnlCardComponent>;
  let apiSpy: jasmine.SpyObj<ApiService>;

  function createFixtureWith(pnlValue?: number, shouldFail = false): void {
    apiSpy = jasmine.createSpyObj<ApiService>("ApiService", ["getPnL"]);

    if (shouldFail) {
      apiSpy.getPnL.and.returnValue(throwError(() => new Error("Network error")));
    } else {
      apiSpy.getPnL.and.returnValue(
        of({ daily_realized_pnl: pnlValue ?? 0 }),
      );
    }

    TestBed.configureTestingModule({
      imports: [PnlCardComponent],
      providers: [{ provide: ApiService, useValue: apiSpy }],
    });

    fixture = TestBed.createComponent(PnlCardComponent);
    fixture.detectChanges();
  }

  // ── Happy paths ──────────────────────────────────────────────────────────

  it("applies 'positive' class when daily PnL is positive", () => {
    createFixtureWith(150.5);
    const pnlEl: HTMLElement = fixture.nativeElement.querySelector(".metric-pnl");
    expect(pnlEl.classList).toContain("positive");
    expect(pnlEl.classList).not.toContain("negative");
  });

  it("applies 'negative' class when daily PnL is negative", () => {
    createFixtureWith(-75.25);
    const pnlEl: HTMLElement = fixture.nativeElement.querySelector(".metric-pnl");
    expect(pnlEl.classList).toContain("negative");
    expect(pnlEl.classList).not.toContain("positive");
  });

  it("calls apiService.getPnL on init", () => {
    createFixtureWith(0);
    expect(apiSpy.getPnL).toHaveBeenCalledTimes(1);
  });

  // ── Edge cases ───────────────────────────────────────────────────────────

  it("applies 'positive' class when PnL is exactly zero", () => {
    createFixtureWith(0);
    const pnlEl: HTMLElement = fixture.nativeElement.querySelector(".metric-pnl");
    expect(pnlEl.classList).toContain("positive");
  });

  // ── Error cases ──────────────────────────────────────────────────────────

  it("shows no PnL value on API error (summary remains null)", () => {
    createFixtureWith(undefined, true);
    // The summary signal stays null — the metric-pnl span should have no number content
    const pnlEl: HTMLElement = fixture.nativeElement.querySelector(".metric-pnl");
    // Angular's decimal pipe renders null/undefined as empty string
    expect(pnlEl.textContent?.trim() ?? "").toBe("");
  });

  it("does not throw when API fails", () => {
    expect(() => createFixtureWith(undefined, true)).not.toThrow();
  });
});
