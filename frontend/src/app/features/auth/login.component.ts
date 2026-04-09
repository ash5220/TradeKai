import { Component, inject, signal } from "@angular/core";
import { FormBuilder, ReactiveFormsModule, Validators } from "@angular/forms";
import { Router, RouterLink } from "@angular/router";
import { AuthService } from "../../core/auth/auth.service";

@Component({
  selector: "tk-login",
  standalone: true,
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: "./login.component.html",
  styleUrl: "./login.component.scss",
})
export class LoginComponent {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly fb = inject(FormBuilder);

  protected readonly form = this.fb.nonNullable.group({
    email: ["", [Validators.required, Validators.email]],
    password: ["", [Validators.required]],
  });

  protected readonly loading = signal(false);
  protected readonly error = signal<string | null>(null);

  protected submit(): void {
    if (this.form.invalid) return;
    const { email, password } = this.form.getRawValue();
    this.loading.set(true);
    this.error.set(null);

    this.auth.login(email, password).subscribe({
      next: () => this.router.navigate(["/dashboard"]),
      error: (err: { error?: { error?: string } }) => {
        this.error.set(err?.error?.error ?? "Login failed");
        this.loading.set(false);
      },
    });
  }
}
