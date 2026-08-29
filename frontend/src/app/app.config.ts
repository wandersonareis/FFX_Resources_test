import { ApplicationConfig } from "@angular/core"
import { provideRouter } from "@angular/router"
import { providePrimeNG } from 'primeng/config';
import { routes } from "./app.routes"

import { FFXPreset } from "../theme/ffx.theme";
import { environment } from "../environments/environment";

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    providePrimeNG({
      theme: {
        preset: FFXPreset
      },
      license: environment.PRIMENG_LICENSE_KEY
    })
  ]
}
