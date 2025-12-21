import { Component, OnInit } from '@angular/core';
import { TechService } from './tech.service';
import { Technology } from './tech.model';
import { NgForOf } from '@angular/common';
import { environment } from '../../environments/environment';


@Component({
    selector: 'app-tech',
    imports: [
        NgForOf
    ],
    templateUrl: './tech.component.html'
})
export class TechComponent implements OnInit {

  public title: string = 'lagident';
  technologies: Technology[] = [];

  constructor(private readonly techService: TechService) { }

  ngOnInit() {
    this.techService.getTechnologies().subscribe(value => {
      this.technologies = value;
    });
  }
}
