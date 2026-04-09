import { Component, inject } from "@angular/core";
import { RouterLink, RouterLinkActive } from "@angular/router";
import { AuthService } from "../../core/auth/auth.service";

@Component({
  selector: "tk-navbar",
  standalone: true,
  imports: [RouterLink, RouterLinkActive],
  templateUrl: "./navbar.component.html",
  styleUrl: "./navbar.component.scss",
})
export class NavbarComponent {
  protected readonly auth = inject(AuthService);
}
