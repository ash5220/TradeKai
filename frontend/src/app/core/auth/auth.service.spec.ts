import { TestBed } from "@angular/core/testing";
import {
  HttpClientTestingModule,
  HttpTestingController,
} from "@angular/common/http/testing";
import { RouterTestingModule } from "@angular/router/testing";
import { AuthService } from "./auth.service";
import { Router } from "@angular/router";

describe("AuthService", () => {
  let service: AuthService;
  let httpMock: HttpTestingController;
  let router: Router;

  beforeEach(() => {
    localStorage.clear();

    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule, RouterTestingModule],
      providers: [AuthService],
    });

    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
    router = TestBed.inject(Router);
    spyOn(router, "navigate");
  });

  afterEach(() => {
    httpMock.verify();
    localStorage.clear();
  });

  // ── Initial state ────────────────────────────────────────────────────────

  it("isAuthenticated is false when no token in localStorage", () => {
    expect(service.isAuthenticated()).toBeFalse();
  });

  it("isAuthenticated is true when access_token is present in localStorage", () => {
    localStorage.setItem("access_token", "existing-token");
    // Re-instantiate service so it picks up the stored token
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule, RouterTestingModule],
      providers: [AuthService],
    });
    const freshService = TestBed.inject(AuthService);
    expect(freshService.isAuthenticated()).toBeTrue();
  });

  // ── register ─────────────────────────────────────────────────────────────

  it("register() POSTs to /auth/register and stores tokens", () => {
    service.register("user@test.com", "password123").subscribe();
    const req = httpMock.expectOne((r) => r.url.includes("/auth/register"));
    expect(req.request.method).toBe("POST");
    expect(req.request.body).toEqual({ email: "user@test.com", password: "password123" });

    req.flush({ access_token: "acc", refresh_token: "ref" });

    expect(localStorage.getItem("access_token")).toBe("acc");
    expect(localStorage.getItem("refresh_token")).toBe("ref");
    expect(service.isAuthenticated()).toBeTrue();
  });

  it("register() sets isAuthenticated to true on success", () => {
    service.register("user@test.com", "pass").subscribe();
    const req = httpMock.expectOne((r) => r.url.includes("/auth/register"));
    req.flush({ access_token: "tok", refresh_token: "ref" });
    expect(service.isAuthenticated()).toBeTrue();
  });

  it("register() errors propagate to subscriber on HTTP 409", (done) => {
    service.register("dup@test.com", "pass").subscribe({
      error: (err) => {
        expect(err.status).toBe(409);
        expect(service.isAuthenticated()).toBeFalse();
        done();
      },
    });
    const req = httpMock.expectOne((r) => r.url.includes("/auth/register"));
    req.flush({ error: "Email taken" }, { status: 409, statusText: "Conflict" });
  });

  // ── login ─────────────────────────────────────────────────────────────────

  it("login() POSTs to /auth/login and stores tokens", () => {
    service.login("user@test.com", "password123").subscribe();
    const req = httpMock.expectOne((r) => r.url.includes("/auth/login"));
    expect(req.request.method).toBe("POST");
    req.flush({ access_token: "acc2", refresh_token: "ref2" });

    expect(localStorage.getItem("access_token")).toBe("acc2");
    expect(service.accessToken()).toBe("acc2");
  });

  it("login() returns 401 error when credentials are wrong", (done) => {
    service.login("user@test.com", "wrong").subscribe({
      error: (err) => {
        expect(err.status).toBe(401);
        done();
      },
    });
    const req = httpMock.expectOne((r) => r.url.includes("/auth/login"));
    req.flush({ error: "Unauthorized" }, { status: 401, statusText: "Unauthorized" });
  });

  // ── logout ────────────────────────────────────────────────────────────────

  it("logout() clears tokens from localStorage", () => {
    localStorage.setItem("access_token", "tok");
    localStorage.setItem("refresh_token", "ref");
    service.logout();
    expect(localStorage.getItem("access_token")).toBeNull();
    expect(localStorage.getItem("refresh_token")).toBeNull();
  });

  it("logout() sets isAuthenticated to false", () => {
    localStorage.setItem("access_token", "tok");
    localStorage.setItem("refresh_token", "ref");
    service.logout();
    expect(service.isAuthenticated()).toBeFalse();
  });

  it("logout() navigates to /auth/login", () => {
    service.logout();
    expect(router.navigate).toHaveBeenCalledWith(["/auth/login"]);
  });

  // ── refresh ───────────────────────────────────────────────────────────────

  it("refresh() POSTs refresh_token and stores new tokens", () => {
    localStorage.setItem("access_token", "old");
    localStorage.setItem("refresh_token", "old-ref");

    // Re-instantiate to pick up localStorage tokens
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule, RouterTestingModule],
      providers: [AuthService],
    });
    const freshService = TestBed.inject(AuthService);
    const freshHttp = TestBed.inject(HttpTestingController);
    spyOn(TestBed.inject(Router), "navigate");

    freshService.refresh()?.subscribe();
    const req = freshHttp.expectOne((r) => r.url.includes("/auth/refresh"));
    expect(req.request.body).toEqual({ refresh_token: "old-ref" });
    req.flush({ access_token: "new-acc", refresh_token: "new-ref" });

    expect(localStorage.getItem("access_token")).toBe("new-acc");
    freshHttp.verify();
  });

  it("refresh() calls logout when no refresh token exists", () => {
    // Service was created with empty localStorage — no refresh token is available
    const logoutSpy = spyOn(service, "logout").and.callThrough();
    service.refresh();
    expect(logoutSpy).toHaveBeenCalled();
  });
});
