import { Routes } from "@angular/router";

export const HISTORY_ROUTES: Routes = [
  {
    path: "",
    loadComponent: () =>
      import("./history.component").then((m) => m.HistoryComponent),
  },
];
