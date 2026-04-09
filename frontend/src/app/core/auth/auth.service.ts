import { computed, inject, signal } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Router } from "@angular/router";
import { Injectable } from "@angular/core";
import { tap } from "rxjs/operators";
import { environment } from "../../../environments/environment";

interface TokenPair {
  access_token: string;
  refresh_token: string;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
}

@Injectable({ providedIn: "root" })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly router = inject(Router);

  private readonly _state = signal<AuthState>({
    accessToken: localStorage.getItem("access_token"),
    refreshToken: localStorage.getItem("refresh_token"),
  });

  readonly isAuthenticated = computed(() => this._state().accessToken !== null);
  readonly accessToken = computed(() => this._state().accessToken);

  register(email: string, password: string) {
    return this.http
      .post<TokenPair>(`${environment.apiUrl}/auth/register`, {
        email,
        password,
      })
      .pipe(tap((pair) => this.storeTokens(pair)));
  }

  login(email: string, password: string) {
    return this.http
      .post<TokenPair>(`${environment.apiUrl}/auth/login`, { email, password })
      .pipe(tap((pair) => this.storeTokens(pair)));
  }

  refresh() {
    const refreshToken = this._state().refreshToken;
    if (!refreshToken) {
      this.logout();
      return;
    }
    return this.http
      .post<TokenPair>(`${environment.apiUrl}/auth/refresh`, {
        refresh_token: refreshToken,
      })
      .pipe(tap((pair) => this.storeTokens(pair)));
  }

  logout(): void {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    this._state.set({ accessToken: null, refreshToken: null });
    this.router.navigate(["/auth/login"]);
  }

  private storeTokens(pair: TokenPair): void {
    localStorage.setItem("access_token", pair.access_token);
    localStorage.setItem("refresh_token", pair.refresh_token);
    this._state.set({
      accessToken: pair.access_token,
      refreshToken: pair.refresh_token,
    });
  }
}
