import { TestBed } from '@angular/core/testing';

import { HistogramsService } from './histograms.service';

describe('HistogramsService', () => {
  let service: HistogramsService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(HistogramsService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
