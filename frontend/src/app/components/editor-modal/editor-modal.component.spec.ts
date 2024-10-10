import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EditorModalComponent } from './editor-modal.component';

describe('EditorModalComponent', () => {
  let component: EditorModalComponent;
  let fixture: ComponentFixture<EditorModalComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EditorModalComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(EditorModalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
