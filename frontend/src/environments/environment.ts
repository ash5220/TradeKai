import type { Environment } from "./environment.interface";

export type { Environment };

export const environment: Environment = {
  production: false,
  apiUrl: "/api/v1",
  wsUrl: `ws://${window.location.host}/ws`,
};
