import { TestBed } from '@angular/core/testing';

import { CompressService } from './compress.service';

describe('CompressService', () => {
  let service: CompressService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(CompressService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
