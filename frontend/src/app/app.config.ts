import { ApplicationConfig } from "@angular/core"
import { provideRouter } from "@angular/router"
import { provideAnimations } from "@angular/platform-browser/animations"
import { MatIconRegistry } from "@angular/material/icon"
import { routes } from "./app.routes"

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideAnimations(),
    {
      provide: 'APP_INITIALIZER',
      useFactory: (registry: MatIconRegistry) => () => {
        registry.setDefaultFontSetClass('material-symbols-outlined');
      },
      deps: [MatIconRegistry],
      multi: true,
    },
  ]
}
