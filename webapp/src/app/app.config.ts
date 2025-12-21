import { ApplicationConfig, importProvidersFrom, isDevMode } from '@angular/core';
import { provideRouter, withHashLocation } from '@angular/router';
import { provideHttpClient, withInterceptors } from "@angular/common/http";

import { routes } from './app.routes';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';

import { provideToastr } from 'ngx-toastr';

import { definePreset } from '@primeng/themes';
import { providePrimeNG } from 'primeng/config';
import Aura from '@primeng/themes/aura';
import Lara from '@primeng/themes/lara';

const MyPreset = definePreset(Lara, {
    semantic: {
        primary: {
            50: '{blue.50}',
            100: '{blue.100}',
            200: '{blue.200}',
            300: '{blue.300}',
            400: '{blue.400}',
            500: '{blue.500}',
            600: '{blue.600}',
            700: '{blue.700}',
            800: '{blue.800}',
            900: '{blue.900}',
            950: '{blue.950}'
        },
    },
});

export const appConfig: ApplicationConfig = {
    providers: [
        { provide: Window, useValue: window },
      
        // For now we use HashLocationStrategy as I was not able to define the correct route fir the app in app.go
        provideRouter(routes, withHashLocation()),

        provideHttpClient(),
        provideToastr({
            timeOut: 3000,
            closeButton: true,
            progressBar: true,
            progressAnimation: 'decreasing',
            positionClass: 'toast-top-right',
            preventDuplicates: true
        }),
        provideAnimationsAsync(),
        providePrimeNG({
            theme: {
              preset: MyPreset,
            },
            ripple: true
          }),
    ]
};
