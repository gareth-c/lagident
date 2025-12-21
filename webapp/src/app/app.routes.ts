import { Component, inject } from "@angular/core";
import { Routes } from '@angular/router';


export const routes: Routes = [{
    path: '',
    loadComponent: () => import('./targets/targets-index/targets-index.component').then(m => m.TargetsIndexComponent),
}, {
    path: 'tech',
    loadComponent: () => import('./tech/tech.component').then(m => m.TechComponent),
}
];
