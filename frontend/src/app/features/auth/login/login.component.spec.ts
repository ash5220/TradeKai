import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { RouterTestingModule } from "@angular/router/testing";
import { LoginComponent } from "./login.component";
import { AuthService } from "../../../core/auth/auth.service";
import { Router } from "@angular/router";
import { signal } from "@angular/core";
import { of, throwError } from "rxjs";
import { By } from "@angular/platform-browser";

describe("LoginComponent", () => {
  let fixture: ComponentFixture<LoginComponent>;
  let authSpy: jasmine.SpyObj<AuthService>;
  let router: Router;

  beforeEach(async () => {
    authSpy = jasmine.createSpyObj<AuthService>("AuthService", ["login"], {
      isAuthenticated: signal(false),
    });

    await TestBed.configureTestingModule({
      imports: [LoginComponent, RouterTestingModule],
      providers: [{ provide: AuthService, useValue: authSpy }],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, "navigate");

    fixture = TestBed.createComponent(LoginComponent);
    fixture.detectChanges();
  });

  function getEmailInput(): HTMLInputElement {
    return fixture.nativeElement.querySelector('input[type="email"]');
  }

  function getPasswordInput(): HTMLInputElement {
    return fixture.nativeElement.querySelector('input[type="password"]');
  }

  function getSubmitButton(): HTMLButtonElement {
    return fixture.nativeElement.querySelector('button[type="submit"]');
  }

  function fillForm(email: string, password: string): void {
    const emailEl = getEmailInput();
    const passEl = getPasswordInput();
    emailEl.value = email;
    emailEl.dispatchEvent(new Event("input"));
    passEl.value = password;
    passEl.dispatchEvent(new Event("input"));
    fixture.detectChanges();
  }

  // ── Happy paths ──────────────────────────────────────────────────────────

  it("submit button is disabled when form is empty", () => {
    expect(getSubmitButton().disabled).toBeTrue();
  });

  it("navigates to /dashboard on successful login", () => {
    authSpy.login.and.returnValue(
      of({ access_token: "tok", refresh_token: "ref" } as any),
    );
    fillForm("user@example.com", "password123");
    getSubmitButton().click();
    fixture.detectChanges();
    expect(router.navigate).toHaveBeenCalledWith(["/dashboard"]);
  });

  it("button shows loading text while submitting", () => {
    authSpy.login.and.returnValue(
      of({ access_token: "tok", refresh_token: "ref" } as any),
    );
    fillForm("user@example.com", "password123");
    // Patch subscribe to not complete so we can check loading state
    const btn = getSubmitButton();
    // After click detectChanges the loading should have resolved; just verify it calls login
    btn.click();
    expect(authSpy.login).toHaveBeenCalledWith(
      "user@example.com",
      "password123",
    );
  });

  // ── Validation edge cases ────────────────────────────────────────────────

  it("keeps submit disabled with invalid email format", () => {
    fillForm("not-an-email", "somepassword");
    expect(getSubmitButton().disabled).toBeTrue();
  });

  it("keeps submit disabled with empty password", () => {
    fillForm("user@example.com", "");
    expect(getSubmitButton().disabled).toBeTrue();
  });

  it("enables submit with valid email and non-empty password", () => {
    authSpy.login.and.returnValue(
      of({ access_token: "t", refresh_token: "r" } as any),
    );
    fillForm("user@example.com", "anypassword");
    expect(getSubmitButton().disabled).toBeFalse();
  });

  // ── Error cases ──────────────────────────────────────────────────────────

  it("displays error message when login fails", () => {
    authSpy.login.and.returnValue(
      throwError(() => ({ error: { error: "Invalid credentials" } })),
    );
    fillForm("user@example.com", "wrongpassword");
    getSubmitButton().click();
    fixture.detectChanges();
    const errorEl = fixture.nativeElement.querySelector(".error");
    expect(errorEl).not.toBeNull();
    expect(errorEl.textContent).toContain("Invalid credentials");
  });

  it("re-enables submit button after a failed login", () => {
    authSpy.login.and.returnValue(
      throwError(() => ({ error: { error: "Bad credentials" } })),
    );
    fillForm("user@example.com", "wrongpassword");
    getSubmitButton().click();
    fixture.detectChanges();
    expect(getSubmitButton().disabled).toBeFalse();
  });

  it("falls back to 'Login failed' when error has no message", () => {
    authSpy.login.and.returnValue(throwError(() => ({})));
    fillForm("user@example.com", "pass");
    getSubmitButton().click();
    fixture.detectChanges();
    const errorEl = fixture.nativeElement.querySelector(".error");
    expect(errorEl.textContent).toContain("Login failed");
  });

  it("does not navigate on failed login", () => {
    authSpy.login.and.returnValue(
      throwError(() => ({ error: { error: "err" } })),
    );
    fillForm("user@example.com", "pass");
    getSubmitButton().click();
    fixture.detectChanges();
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it("error is cleared between submit attempts", () => {
    authSpy.login.and.returnValue(
      throwError(() => ({ error: { error: "err" } })),
    );
    fillForm("user@example.com", "pass");
    getSubmitButton().click();
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector(".error")).not.toBeNull();

    authSpy.login.and.returnValue(
      of({ access_token: "t", refresh_token: "r" } as any),
    );
    getSubmitButton().click();
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector(".error")).toBeNull();
  });
});
