import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../core/auth/auth.service';

@Component({
  selector: 'tk-register',
  standalone: true,
  imports: [ReactiveFormsModule, RouterLink],
  template: `
    <div class="auth-container">
      <div class="auth-card">
        <h1>TradeKai</h1>
        <h2>Create Account</h2>
        <form [formGroup]="form" (ngSubmit)="submit()">
          <label>
            Email
            <input type="email" formControlName="email" placeholder="you@example.com" />
          </label>
          <label>
            Password
            <input type="password" formControlName="password" placeholder="At least 8 characters" />
          </label>
          @if (error()) {
            <p class="error">{{ error() }}</p>
          }
          <button type="submit" [disabled]="form.invalid || loading()">
            {{ loading() ? 'Creating account…' : 'Register' }}
          </button>
        </form>
        <p class="link">Already have an account? <a routerLink="/auth/login">Sign in</a></p>
      </div>
    </div>
  `,
  styles: [`
    .auth-container {
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 80vh;
    }
    .auth-card {
      background: #1a1d27;
      border: 1px solid #2a2d3a;
      border-radius: 8px;
      padding: 2rem;
      width: 100%;
      max-width: 420px;
    }
    h1 { color: #00d1b2; margin: 0 0 0.25rem; }
    h2 { color: #fff; margin: 0 0 1.5rem; font-weight: 400; }
    label { display: flex; flex-direction: column; gap: 0.4rem; margin-bottom: 1rem; color: #aaa; font-size: 0.875rem; }
    input {
      background: #0f1117;
      border: 1px solid #2a2d3a;
      border-radius: 4px;
      color: #fff;
      padding: 0.6rem 0.75rem;
      font-size: 1rem;
      &:focus { outline: none; border-color: #00d1b2; }
    }
    button {
      width: 100%;
      background: #00d1b2;
      color: #000;
      border: none;
      border-radius: 4px;
      padding: 0.75rem;
      font-size: 1rem;
      font-weight: 600;
      cursor: pointer;
      margin-top: 0.5rem;
      &:disabled { opacity: 0.5; cursor: not-allowed; }
    }
    .error { color: #ff6b6b; font-size: 0.875rem; margin: 0 0 0.75rem; }
    .link { text-align: center; color: #aaa; font-size: 0.875rem; a { color: #00d1b2; } }
  `],
})
export class RegisterComponent {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly fb = inject(FormBuilder);

  protected readonly form = this.fb.nonNullable.group({
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(8)]],
  });

  protected readonly loading = signal(false);
  protected readonly error = signal<string | null>(null);

  protected submit(): void {
    if (this.form.invalid) return;
    const { email, password } = this.form.getRawValue();
    this.loading.set(true);
    this.error.set(null);

    this.auth.register(email, password).subscribe({
      next: () => this.router.navigate(['/dashboard']),
      error: (err: { error?: { error?: string } }) => {
        this.error.set(err?.error?.error ?? 'Registration failed');
        this.loading.set(false);
      },
    });
  }
}
