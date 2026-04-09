import { TestBed } from "@angular/core/testing";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import { HttpClient, provideHttpClient, withInterceptors } from "@angular/common/http";
import { computed } from "@angular/core";
import { authInterceptor } from "./auth.interceptor";
import { AuthService } from "./auth.service";

describe("authInterceptor", () => {
  let http: HttpClient;
  let httpMock: HttpTestingController;
  let authServiceMock: jasmine.SpyObj<AuthService>;

  function setToken(token: string | null): void {
    Object.defineProperty(authServiceMock, "accessToken", {
      value: computed(() => token),
      configurable: true,
    });
  }

  beforeEach(() => {
    authServiceMock = jasmine.createSpyObj<AuthService>("AuthService", ["logout"], {
      accessToken: computed(() => "test-token"),
      isAuthenticated: computed(() => true),
    });

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([authInterceptor])),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authServiceMock },
      ],
    });

    http = TestBed.inject(HttpClient);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  // ── Happy path ────────────────────────────────────────────────────────────

  it("adds Authorization header when a token is present", () => {
    http.get("/api/test").subscribe();
    const req = httpMock.expectOne("/api/test");
    expect(req.request.headers.get("Authorization")).toBe("Bearer test-token");
    req.flush({});
  });

  // ── Edge case: no token ───────────────────────────────────────────────────

  it("does NOT add Authorization header when accessToken is null", () => {
    setToken(null);

    http.get("/api/test").subscribe();
    const req = httpMock.expectOne("/api/test");
    expect(req.request.headers.has("Authorization")).toBeFalse();
    req.flush({});
  });

  // ── Error: 401 triggers logout ────────────────────────────────────────────

  it("calls auth.logout() when the response is 401", () => {
    http.get("/api/protected").subscribe({ error: () => {} });
    const req = httpMock.expectOne("/api/protected");
    req.flush({ error: "Unauthorized" }, { status: 401, statusText: "Unauthorized" });
    expect(authServiceMock.logout).toHaveBeenCalled();
  });

  it("does NOT call auth.logout() for non-401 errors", () => {
    http.get("/api/data").subscribe({ error: () => {} });
    const req = httpMock.expectOne("/api/data");
    req.flush({ error: "Server Error" }, { status: 500, statusText: "Internal Server Error" });
    expect(authServiceMock.logout).not.toHaveBeenCalled();
  });
});
