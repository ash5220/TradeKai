import type { Environment } from "./environment.interface";

export const environment: Environment = {
  production: true,
  apiUrl: "/api/v1",
  wsUrl: `wss://${window.location.host}/ws`,
};
