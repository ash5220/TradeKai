import { TestBed } from "@angular/core/testing";
import { Router, UrlTree } from "@angular/router";
import { RouterTestingModule } from "@angular/router/testing";
import { authGuard } from "./auth.guard";
import { AuthService } from "./auth.service";
import { computed } from "@angular/core";

describe("authGuard", () => {
  let router: Router;
  let authServiceMock: jasmine.SpyObj<AuthService>;

  function runGuard(): ReturnType<typeof authGuard> {
    return TestBed.runInInjectionContext(() => authGuard({} as any, {} as any));
  }

  beforeEach(() => {
    authServiceMock = jasmine.createSpyObj<AuthService>(
      "AuthService",
      ["logout"],
      {
        isAuthenticated: computed(() => false),
        accessToken: computed(() => null),
      },
    );

    TestBed.configureTestingModule({
      imports: [RouterTestingModule],
      providers: [{ provide: AuthService, useValue: authServiceMock }],
    });

    router = TestBed.inject(Router);
  });

  // ── Happy path ────────────────────────────────────────────────────────────

  it("returns true when the user is authenticated", () => {
    // Override isAuthenticated to return true
    Object.defineProperty(authServiceMock, "isAuthenticated", {
      value: computed(() => true),
    });

    const result = runGuard();
    expect(result).toBeTrue();
  });

  // ── Redirect ──────────────────────────────────────────────────────────────

  it("returns a UrlTree to /auth/login when the user is not authenticated", () => {
    const result = runGuard();
    expect(result).toBeInstanceOf(UrlTree);
    expect((result as UrlTree).toString()).toBe("/auth/login");
  });
});
