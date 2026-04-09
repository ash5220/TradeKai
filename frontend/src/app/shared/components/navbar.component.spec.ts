import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { RouterTestingModule } from "@angular/router/testing";
import { NavbarComponent } from "./navbar.component";
import { AuthService } from "../../core/auth/auth.service";
import { Router } from "@angular/router";
import { signal } from "@angular/core";
import { By } from "@angular/platform-browser";

describe("NavbarComponent", () => {
  let fixture: ComponentFixture<NavbarComponent>;
  let authSpy: jasmine.SpyObj<AuthService>;
  let router: Router;

  function setup(authenticated: boolean): void {
    authSpy = jasmine.createSpyObj<AuthService>(
      "AuthService",
      ["logout"],
      { isAuthenticated: signal(authenticated) },
    );
    authSpy.logout.and.callFake(() => {});
  }

  async function createFixture(): Promise<void> {
    await TestBed.configureTestingModule({
      imports: [NavbarComponent, RouterTestingModule],
      providers: [{ provide: AuthService, useValue: authSpy }],
    }).compileComponents();

    router = TestBed.inject(Router);
    fixture = TestBed.createComponent(NavbarComponent);
    fixture.detectChanges();
  }

  describe("when authenticated", () => {
    beforeEach(async () => {
      setup(true);
      await createFixture();
    });

    it("shows the brand link", () => {
      const brand = fixture.nativeElement.querySelector(".brand");
      expect(brand.textContent.trim()).toBe("TradeKai");
    });

    it("renders all three nav links", () => {
      const links: NodeListOf<HTMLAnchorElement> =
        fixture.nativeElement.querySelectorAll(".nav-links a");
      const hrefs = Array.from(links).map((l) =>
        l.getAttribute("routerlink") ?? l.getAttribute("ng-reflect-router-link"),
      );
      expect(hrefs).toContain("/dashboard");
      expect(hrefs).toContain("/trading");
      expect(hrefs).toContain("/history");
    });

    it("shows the logout button", () => {
      const btn = fixture.nativeElement.querySelector(".btn-logout");
      expect(btn).not.toBeNull();
    });

    it("calls auth.logout when logout button is clicked", () => {
      const btn: HTMLButtonElement =
        fixture.nativeElement.querySelector(".btn-logout");
      btn.click();
      expect(authSpy.logout).toHaveBeenCalledTimes(1);
    });
  });

  describe("when not authenticated", () => {
    beforeEach(async () => {
      setup(false);
      await createFixture();
    });

    it("does not show nav links", () => {
      const links = fixture.nativeElement.querySelectorAll(".nav-links");
      expect(links.length).toBe(0);
    });

    it("does not show the logout button", () => {
      const btn = fixture.nativeElement.querySelector(".btn-logout");
      expect(btn).toBeNull();
    });

    it("still shows the brand link", () => {
      const brand = fixture.nativeElement.querySelector(".brand");
      expect(brand).not.toBeNull();
    });
  });
});
