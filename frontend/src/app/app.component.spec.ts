import { TestBed } from "@angular/core/testing";
import { ComponentFixture } from "@angular/core/testing";
import { RouterTestingModule } from "@angular/router/testing";
import { AppComponent } from "./app.component";
import { NavbarComponent } from "./shared/components/navbar.component";
import { AuthService } from "./core/auth/auth.service";
import { signal } from "@angular/core";
import { By } from "@angular/platform-browser";

describe("AppComponent", () => {
  let fixture: ComponentFixture<AppComponent>;
  let authSpy: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    authSpy = jasmine.createSpyObj<AuthService>("AuthService", ["logout"], {
      isAuthenticated: signal(false),
    });

    await TestBed.configureTestingModule({
      imports: [AppComponent, RouterTestingModule],
      providers: [{ provide: AuthService, useValue: authSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(AppComponent);
    fixture.detectChanges();
  });

  it("renders the navbar component", () => {
    const navbar = fixture.debugElement.query(By.directive(NavbarComponent));
    expect(navbar).not.toBeNull();
  });

  it("renders a router-outlet", () => {
    const outlet = fixture.nativeElement.querySelector("router-outlet");
    expect(outlet).not.toBeNull();
  });

  it("wraps content in a main.main-content element", () => {
    const main = fixture.nativeElement.querySelector("main.main-content");
    expect(main).not.toBeNull();
  });
});
