import { ApplicationConfig, provideExperimentalZonelessChangeDetection } from "@angular/core"
import { provideRouter } from "@angular/router"
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { providePrimeNG } from 'primeng/config';
import { routes } from "./app.routes"

import { FFXPreset } from "../theme/ffx.theme";

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideAnimationsAsync(),
    provideExperimentalZonelessChangeDetection(),
    providePrimeNG({
      theme: {
        preset: FFXPreset
      }
    })
  ]
}
