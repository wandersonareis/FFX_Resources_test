import { TestBed } from '@angular/core/testing';
import { FfxContextMenuService } from './ffx-context-menu.service';


describe('FfxContextMenuService', () => {
  let service: FfxContextMenuService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(FfxContextMenuService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
