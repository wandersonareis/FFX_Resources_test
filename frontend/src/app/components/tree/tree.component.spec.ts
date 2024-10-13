import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FfxTreeComponent } from './tree.component';

describe('TreeContextMenuDemoComponent', () => {
  let component: FfxTreeComponent;
  let fixture: ComponentFixture<FfxTreeComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FfxTreeComponent]
    })
      .compileComponents();

    fixture = TestBed.createComponent(FfxTreeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
