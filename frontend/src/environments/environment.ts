export interface Environment {
  production: boolean;
  apiUrl: string;
  wsUrl: string;
}

export const environment: Environment = {
  production: false,
  apiUrl: "/api/v1",
  wsUrl: `ws://${window.location.host}/ws`,
};
