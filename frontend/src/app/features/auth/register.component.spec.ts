import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { RouterTestingModule } from "@angular/router/testing";
import { RegisterComponent } from "./register.component";
import { AuthService } from "../../core/auth/auth.service";
import { Router } from "@angular/router";
import { signal } from "@angular/core";
import { of, throwError } from "rxjs";

describe("RegisterComponent", () => {
  let fixture: ComponentFixture<RegisterComponent>;
  let authSpy: jasmine.SpyObj<AuthService>;
  let router: Router;

  beforeEach(async () => {
    authSpy = jasmine.createSpyObj<AuthService>(
      "AuthService",
      ["register"],
      { isAuthenticated: signal(false) },
    );

    await TestBed.configureTestingModule({
      imports: [RegisterComponent, RouterTestingModule],
      providers: [{ provide: AuthService, useValue: authSpy }],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, "navigate");

    fixture = TestBed.createComponent(RegisterComponent);
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

  it("navigates to /dashboard on successful registration", () => {
    authSpy.register.and.returnValue(of({ access_token: "tok", refresh_token: "ref" } as any));
    fillForm("new@example.com", "securepass");
    getSubmitButton().click();
    fixture.detectChanges();
    expect(router.navigate).toHaveBeenCalledWith(["/dashboard"]);
  });

  it("calls auth.register with trimmed credentials", () => {
    authSpy.register.and.returnValue(of({ access_token: "t", refresh_token: "r" } as any));
    fillForm("new@example.com", "securepass");
    getSubmitButton().click();
    expect(authSpy.register).toHaveBeenCalledWith("new@example.com", "securepass");
  });

  // ── Validation edge cases ────────────────────────────────────────────────

  it("disables submit when password is shorter than 8 characters", () => {
    fillForm("new@example.com", "short");
    expect(getSubmitButton().disabled).toBeTrue();
  });

  it("enables submit when password is exactly 8 characters", () => {
    authSpy.register.and.returnValue(of({ access_token: "t", refresh_token: "r" } as any));
    fillForm("new@example.com", "exactly8");
    expect(getSubmitButton().disabled).toBeFalse();
  });

  it("disables submit with invalid email format", () => {
    fillForm("notanemail", "validpassword");
    expect(getSubmitButton().disabled).toBeTrue();
  });

  // ── Error cases ──────────────────────────────────────────────────────────

  it("shows error message when registration fails", () => {
    authSpy.register.and.returnValue(
      throwError(() => ({ error: { error: "Email already in use" } })),
    );
    fillForm("exists@example.com", "password1");
    getSubmitButton().click();
    fixture.detectChanges();
    const errorEl = fixture.nativeElement.querySelector(".error");
    expect(errorEl.textContent).toContain("Email already in use");
  });

  it("re-enables submit button after registration error", () => {
    authSpy.register.and.returnValue(
      throwError(() => ({ error: { error: "Server error" } })),
    );
    fillForm("u@example.com", "password1");
    getSubmitButton().click();
    fixture.detectChanges();
    expect(getSubmitButton().disabled).toBeFalse();
  });

  it("falls back to 'Registration failed' when error has no message", () => {
    authSpy.register.and.returnValue(throwError(() => ({})));
    fillForm("u@example.com", "password1");
    getSubmitButton().click();
    fixture.detectChanges();
    const errorEl = fixture.nativeElement.querySelector(".error");
    expect(errorEl.textContent).toContain("Registration failed");
  });

  it("does not navigate on failed registration", () => {
    authSpy.register.and.returnValue(throwError(() => ({ error: { error: "err" } })));
    fillForm("u@example.com", "password1");
    getSubmitButton().click();
    fixture.detectChanges();
    expect(router.navigate).not.toHaveBeenCalled();
  });
});
